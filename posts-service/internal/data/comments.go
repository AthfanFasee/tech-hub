package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AthfanFasee/posts/internal/validator"
)

type Comment struct {
	ID        int64
	Text      string
	UserID    int64
	UserName  string
	PostID    int64
	CreatedAt time.Time
}

type CommentModel struct {
	DB *sql.DB
}

// Gets all the comments for a single post
func (c CommentModel) GetAllForPost(postID int64) ([]*Comment, error) {
	query := `SELECT id, text, user_id, user_name, post_id
	FROM comments
	WHERE post_id = $1
	ORDER BY id DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, postID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	defer rows.Close()

	comments := []*Comment{}

	for rows.Next() {
		var comment Comment

		err := rows.Scan(
			&comment.ID,
			&comment.Text,
			&comment.UserID,
			comment.UserName,
			&comment.PostID,
		)

		if err != nil {
			return nil, err
		}

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// Creates a comment
func (c CommentModel) Insert(comment *Comment) error {
	query := `
	INSERT INTO comments (text, post_id, user_id, user_name)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	args := []interface{}{comment.Text, comment.PostID, comment.UserID, comment.UserName}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&comment.ID)
}

// Deletes a single comment
func (c CommentModel) Delete(id int64) error {
	query := `
	DELETE FROM comments
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Validates comment data
func ValidateComment(v *validator.Validator, comment *Comment) {
	v.Check(comment.Text != "", "text", "Comment cannot be empty")
	v.Check(len(comment.Text) <= 200, "text", "Comment can only contain 200 characters or less")

	v.Check(comment.PostID != 0, "post", "Post id must be provided")
	v.Check(comment.PostID > 0, "post", "Post id must be valid")
}
