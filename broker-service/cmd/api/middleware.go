package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

// Recovers incase of panic and handles it
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Adds few security headers
func (app *application) secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

// Provides client based rate limiting
func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			// If a client has'nt been within last 5 mins, delete their entry from map
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 5*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			// Use the realip.FromRequest() function to get the client's real IP address
			ip := realip.FromRequest(r)

			mu.Lock()

			if _, ok := clients[ip]; !ok {
				// Allows an average of 2 requests per second, with a maximum of 4 requests in a single ‘burst’ (by default)
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			// Update last seen time for client
			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

// Handles CORS issues
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		// Response will be different according to either this header exists or not
		// Bcs this header determines either the current request is preflight or not
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
			for _, trustedOrigin := range app.config.cors.trustedOrigins {
				if origin == trustedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// Treat preflight requests differently
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")

						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						// Some browsers doesn't support 204. So prefer 200 here
						w.WriteHeader(http.StatusOK)

						return
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Authenticates user and passes user info to request context
func (app *application) authenticate(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUserInfo(r, 0, "", false, false)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")

		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		tokenString := headerParts[1]

		// Load the public key
		keysDir := filepath.Join(".", "keys")
		publicKeyPath := filepath.Join(keysDir, "public.pem")
		publicBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicBytes)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				app.serverErrorResponse(w, r, err)
			}

			return publicKey, nil
		})

		var claims jwt.MapClaims

		if claimsMap, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			claims = claimsMap
		} else {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		userID := claims["user_id"].(int64)
		userName := claims["user_name"].(string)
		userActivated := claims["user_activated"].(bool)

		r = app.contextSetUserInfo(r, userID, userName, userActivated, true)

		next.ServeHTTP(w, r)

	})
}

// Checks if a user authenticated
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAuthenticated := app.contextGetUserAuthenticatedStatus(r)

		if !userAuthenticated {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Checks if a user is both authenticated and activated
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, userActivated := app.contextGetUserInfo(r)

		if !userActivated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}
