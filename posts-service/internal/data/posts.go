package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AthfanFasee/posts/internal/validator"
	"github.com/lib/pq"
)

type Post struct {
	ID        int64
	Title     string
	PostText  string
	Img       string
	ReadTime  int32
	LikedBy   []int64
	UserID    int64
	UserName  string
	Version   int32
	CreatedAt time.Time
}

type UpdatePostRequestBody struct {
	Title    *string `json:"title"`
	PostText *string `json:"postText"`
	ReadTime *int32  `json:"readTime"`
	Img      *string `json:"img"`
}

type PostModel struct {
	DB *sql.DB
}

// Gets all posts
func (p PostModel) GetAll(title string, filters Filters) ([]*Post, Metadata, error) {
	// Get post data along with name of the user who created it
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, title, post_text, img, read_time, liked_by, user_id, user_name, created_at
	FROM posts
	WHERE (to_tsvector('english', title) @@ plainto_tsquery('english', $1) OR $1 = '')
	AND (created_by = $2 OR $2 = 0)
	ORDER BY %s %s, id %s
	LIMIT $3 OFFSET $4`, filters.sortParam(), filters.sortOrder(), filters.sortOrder())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, filters.ID, filters.limit(), filters.offset()}

	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	posts := []*Post{}

	for rows.Next() {
		var post Post
		err := rows.Scan(
			&totalRecords,
			&post.ID,
			&post.Title,
			&post.PostText,
			&post.Img,
			&post.ReadTime,
			pq.Array(&post.LikedBy),
			&post.UserID,
			&post.UserName,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.Limit)

	return posts, metadata, nil
}

// Gets a single post by id
func (p PostModel) Get(id int64) (*Post, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT id, title, post_text, img, read_time, liked_by, user_id, user_name, created_at, version
	FROM posts
	WHERE id = $1`

	var post Post

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.PostText,
		&post.Img,
		&post.ReadTime,
		pq.Array(&post.LikedBy),
		&post.UserID,
		&post.UserName,
		&post.CreatedAt,
		&post.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

// Inserts post
func (p PostModel) Insert(post *Post) error {
	query := `
		INSERT INTO posts (title, post_text, img, read_time, user_id, user_name) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, created_at`

	args := []interface{}{post.Title, post.PostText, post.Img, post.ReadTime, post.UserID, post.UserName}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, args...).Scan(&post.ID, &post.CreatedAt)
}

// Updates a single post by id
func (p PostModel) Update(post *Post) error {
	query := `
	UPDATE posts
	SET title = $1, post_text = $2, img = $3, read_time = $4, version = version + 1
	WHERE id = $5 AND version = $6`

	args := []interface{}{
		post.Title,
		post.PostText,
		post.Img,
		post.ReadTime,
		post.ID,
		post.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := p.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
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

// Deletes a single post by id
func (p PostModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE FROM posts
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := p.DB.ExecContext(ctx, query, id)
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

// Deletes all the posts for a single user
func (p PostModel) DeleteForUser(userID int64) error {
	if userID < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE FROM posts
	WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := p.DB.ExecContext(ctx, query, userID)
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

// Adds like to a single post
func (p PostModel) AddLike(post *Post, userID int64) error {
	// This SQL statement will prevent a user from liking a post twice
	query := `
	UPDATE posts SET 
	liked_by = (select array_agg(distinct x) from unnest(array_append(liked_by, $1)) t(x))
	WHERE id = $2
	RETURNING liked_by`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, userID, post.ID).Scan(pq.Array(&post.LikedBy))
}

// Adds dilike to a single post
func (p PostModel) RemoveLike(post *Post, userID int64) error {
	query := `
	UPDATE posts SET liked_by = array_remove(liked_by, $1)
	WHERE id = $2
	RETURNING liked_by`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, userID, post.ID).Scan(pq.Array(&post.LikedBy))
}

// Validates post data
func ValidatePost(v *validator.Validator, post *Post) {
	v.Check(post.Title != "", "title", "Title must be provided")
	v.Check(len(post.Title) <= 100, "title", "Title can only contain 100 characters or less")

	v.Check(post.PostText != "", "postText", "Text must be provided")
	v.Check(post.Img != "", "img", "Image must be provided")

	v.Check(post.ReadTime != 0, "readTime", "Read time must be provided")
	v.Check(post.ReadTime > 0, "readTime", "Read time must be provided")
}
