package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/AthfanFasee/authentication/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

const dbTimeout = time.Second * 3

var db *sql.DB

// User is the structure which holds one user from the database.
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  Password  `json:"-"`
	Admin     bool      `json:"admin"`
	Bio       string    `json:"bio"`
	Avatar    string    `json:"avatar"`
	Activated bool      `json:"activated"`
	CreatedAt time.Time `json:"createdAt"`
	Version   int32     `json:"-"`
}

type Password struct {
	plainText *string
	hash      []byte
}

type UserModel struct {
	DB *sql.DB
}

// Returns one user by email
func (u UserModel) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, email, name, admin, activated, bio, avatar, created_at FROM users WHERE email = $1`

	var user User
	row := db.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Admin,
		&user.Activated,
		&user.Bio,
		&user.Avatar,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Returns one user by id
func (u UserModel) GetOne(id int64) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, email, name, admin, activated, bio, avatar, created_at FROM users WHERE id = $1`

	var user User
	row := db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Admin,
		&user.Activated,
		&user.Bio,
		&user.Avatar,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Insert inserts a new user into the database, and returns the ID of the newly inserted row
func (u UserModel) Insert(user *User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password.hash), 12)
	if err != nil {
		return 0, err
	}

	var newID int
	stmt := `INSERT INTO users (email, name, password_hash, admin, activated, bio, avatar, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	err = db.QueryRowContext(ctx, stmt,
		user.Email,
		user.Name,
		hashedPassword,
		false,
		user.Activated,
		user.Bio,
		user.Avatar,
		time.Now(),
	).Scan(&newID)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return 0, ErrDuplicateEmail
		default:
			return 0, err
		}
	}

	return newID, nil
}

// Updates one user in the database, using the information stored in the receiver u
func (u UserModel) Update(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `UPDATE users SET
		email = $1,
		name = $2,
		admin = $3
		activated = $4,
		bio = $5,
		avatar = $6,
		updated_at = $5
		WHERE id = $6 AND version = $7
	`

	result, err := db.ExecContext(ctx, stmt,
		user.Email,
		user.Name,
		user.Admin,
		user.Activated,
		user.Bio,
		user.Avatar,
		time.Now(),
		user.ID,
		user.Version,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrEditConflict
	}

	return nil
}

// Deletes one user from the database, by ID
func (u UserModel) DeleteByID(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM users WHERE id = $1`

	_, err := db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}

// Returns a user by token
func (u UserModel) GetForToken(tokenScope, tokenPlainText string) (*User, error) {
	query := `
	SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.admin, users.bio, users.avatar
	FROM users
	INNER JOIN tokens
	ON users.id = tokens.user_id
	WHERE tokens.hash = $1
	AND tokens.scope = $2
	AND tokens.expiry > $3`

	tokenHash := sha256.Sum256([]byte(tokenPlainText))

	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Admin,
		&user.Bio,
		&user.Avatar,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// ResetPassword is the method we will use to change a user's password.
func (u UserModel) ResetPassword(password string, id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `UPDATE users SET password = $1 WHERE id = $2`
	_, err = db.ExecContext(ctx, stmt, hashedPassword, id)
	if err != nil {
		return err
	}

	return nil
}

// Sets plain and hashed passwords
func (p *Password) Set(plainTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return err
	}

	p.plainText = &plainTextPassword
	p.hash = hash

	return nil
}

// PasswordMatches compares a user supplied password with the hash we have stored for a given user.
func (p *Password) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// invalid password
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

// Validation related.
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "email must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "password must be provided")
	v.Check(len(password) >= 6, "password", "password must be at least 6 bytes long")
	v.Check(len(password) <= 72, "password", "password must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "name must be provided")
	v.Check(len(user.Name) <= 100, "name", "name must not be more than 100 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plainText != nil {
		ValidatePasswordPlaintext(v, *user.Password.plainText)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}
