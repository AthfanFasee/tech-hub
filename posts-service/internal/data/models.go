package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Comments interface {
		GetAllForPost(postID int64) ([]*Comment, error)
		Insert(comment *Comment) error
		Delete(id int64) error
	}
	Posts interface {
		GetAll(title string, filters Filters) ([]*Post, Metadata, error)
		Get(id int64) (*Post, error)
		Insert(post *Post) error
		Update(post *Post) error
		Delete(id int64) error
		DeleteForUser(userID int64) error
		AddLike(post *Post, userID int64) error
		RemoveLike(post *Post, userID int64) error
	}
}

// Creates a new Models instance
func NewModels(db *sql.DB) Models {
	return Models{
		Comments: CommentModel{DB: db},
		Posts:    PostModel{DB: db},
	}
}
