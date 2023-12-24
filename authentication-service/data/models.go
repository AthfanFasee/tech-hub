package data

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateEmail = errors.New("duplicate email")
)

type Models struct {
	Tokens interface {
		Insert(token *Token) error
		New(userID int64, timeToLive time.Duration, scope string) (*Token, error)
		DeleteAllForUser(scope string, userID int64) error
	}
	Users interface {
		Insert(user *User) (int, error)
		GetByEmail(email string) (*User, error)
		GetOne(id int) (*User, error)
		DeleteByID(id int) error
		Update(user *User) error
		GetForToken(tokenScope, tokenPlainText string) (*User, error)
		ResetPassword(password string, id int) error
	}
}

func NewModels(db *sql.DB) Models {
	return Models{
		Tokens: TokenModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
