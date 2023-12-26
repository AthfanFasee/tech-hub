package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateEmail = errors.New("duplicate email")
)

type Models struct {
	Users interface {
		Insert(user *User) (int, error)
		GetByEmail(email string) (*User, error)
		GetOne(id int64) (*User, error)
		DeleteByID(id int64) error
		Update(user *User) error
		ResetPassword(password string, id int64) error
	}
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
	}
}
