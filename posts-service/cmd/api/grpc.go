package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/AthfanFasee/posts/internal/data"
	"github.com/AthfanFasee/posts/internal/validator"
	posts "github.com/AthfanFasee/posts/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PostsService struct {
	posts.UnimplementedPostsServiceServer
	Models      data.Models
	DB          *sql.DB
	Application *application
}

type RabbitPayload struct {
	Name string
	Data any
}

// Gets all the posts
func (u *PostsService) GetPosts(ctx context.Context, req *posts.Empty) (*posts.GetPostsResponse, error) {

	// Get filter values from gRPC metadata
	var userID string
	var page string
	var limit string
	var sort string
	var title string

	AssignMetadataValue(ctx, "user-id", &userID)
	AssignMetadataValue(ctx, "page", &userID)
	AssignMetadataValue(ctx, "limit", &userID)
	AssignMetadataValue(ctx, "sort", &userID)
	AssignMetadataValue(ctx, "title", &userID)

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user ID: "+userID)
	}
	pageInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid page value "+page)
	}
	limitInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid limit value: "+limit)
	}

	filters := data.Filters{
		ID:           userIDInt,
		Page:         int(pageInt),
		Limit:        int(limitInt),
		Sort:         sort,
		SortSafeList: []string{"id", "title", "readtime", "likescount", "-id", "-title", "-readtime", "-likescount"},
	}

	result, metaData, err := u.Models.Posts.GetAll(title, filters)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	// Get posts ready to send via gRPC
	grpcPosts := make([]*posts.Post, len(result))
	for i, post := range result {
		grpcPosts[i] = ConvertPostToGRPCResponse(post)
	}

	grpcMetaData := &posts.Metadata{
		CurrentPage:  int32(metaData.CurrentPage),
		PageSize:     int32(metaData.PageSize),
		FirstPage:    int32(metaData.FirstPage),
		LastPage:     int32(metaData.LastPage),
		TotalRecords: int32(metaData.TotalRecords),
	}

	res := &posts.GetPostsResponse{Posts: grpcPosts, MetaData: grpcMetaData}
	return res, nil
}

// Gets posts ready to send via gRPC
func ConvertPostToGRPCResponse(post *data.Post) *posts.Post {
	return &posts.Post{
		Id:        post.ID,
		Title:     post.Title,
		PostText:  post.PostText,
		Img:       post.Img,
		ReadTime:  post.ReadTime,
		LikedBy:   post.LikedBy,
		UserId:    post.UserID,
		UserName:  post.UserName,
		CreatedAt: timestamppb.New(post.CreatedAt),
	}
}

// Assings metadata value for the specific key to the target
func AssignMetadataValue(ctx context.Context, key string, target *string) {
	var values []string

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values = md.Get(key)
	}

	if len(values) > 0 {
		*target = values[0]
	}
}

// Gets a single post by id
func (u *PostsService) GetPost(ctx context.Context, req *posts.GetPostRequest) (*posts.GetPostResponse, error) {
	id := req.GetId()
	result, err := u.Models.Posts.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	post := posts.Post{
		Id:        result.ID,
		Title:     result.Title,
		PostText:  result.PostText,
		Img:       result.Img,
		ReadTime:  result.ReadTime,
		LikedBy:   result.LikedBy,
		UserId:    result.UserID,
		UserName:  result.UserName,
		CreatedAt: timestamppb.New(result.CreatedAt),
	}

	res := &posts.GetPostResponse{Post: &post}
	return res, nil
}

// Creates a post
func (u *PostsService) CreatePost(ctx context.Context, req *posts.CreatePostRequest) (*posts.GetPostResponse, error) {
	title := req.GetTitle()
	postText := req.GetPostText()
	img := req.GetImg()
	readTime := req.GetReadTime()
	userID := req.GetUserId()
	userName := req.GetUserName()

	post := &data.Post{
		Title:    strings.TrimSpace(title),
		PostText: strings.TrimSpace(postText),
		Img:      img,
		ReadTime: readTime,
		UserID:   userID,
		UserName: userName,
	}

	v := validator.New()
	data.ValidatePost(v, post)

	err := u.Application.checkValidationStatus(v)
	if err != nil {
		return nil, err
	}

	err = u.Models.Posts.Insert(post)
	if err != nil {
		err := u.Application.ServerErrorResponse(err.Error())
		if err != nil {
			return nil, err
		}
	}

	postResponse := posts.Post{
		Id:        post.ID,
		Title:     post.Title,
		PostText:  post.PostText,
		Img:       post.Img,
		ReadTime:  post.ReadTime,
		LikedBy:   post.LikedBy,
		UserId:    post.UserID,
		UserName:  post.UserName,
		CreatedAt: timestamppb.New(post.CreatedAt),
	}

	res := &posts.GetPostResponse{Post: &postResponse}
	return res, nil
}

// Updates a post
func (u *PostsService) UpdatePost(ctx context.Context, req *posts.UpdatePostRequest) (*posts.GetPostResponse, error) {
	id := req.GetId()
	title := req.GetTitle()
	postText := req.GetPostText()
	img := req.GetImg()
	readTime := req.GetReadTime()

	// Check if a post with provided id exists.
	post, err := u.Application.Models.Posts.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	var input data.UpdatePostRequestBody

	input.Title = &title
	input.PostText = &postText
	input.ReadTime = &readTime
	input.Img = &img

	// Copy values from input to appropriate fields of post record only if they are not nil.
	if input.Title != nil {
		post.Title = strings.TrimSpace(*input.Title)
	}
	if input.PostText != nil {
		post.PostText = strings.TrimSpace(*input.PostText)
	}
	if input.Img != nil {
		post.Img = *input.Img
	}
	if input.ReadTime != nil {
		post.ReadTime = *input.ReadTime
	}

	v := validator.New()

	// Title and PostText must be provided by the client (other fields are optional when updating).
	if nil == input.Title {
		v.AddError("title", "must be provided")
	}
	if nil == input.PostText {
		v.AddError("postText", "must be provided")
	}

	data.ValidatePost(v, post)

	err = u.Application.checkValidationStatus(v)
	if err != nil {
		return nil, err
	}

	err = u.Models.Posts.Update(post)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			err := u.Application.EditConflictResponse(err.Error())
			if err != nil {
				return nil, err
			}
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	postResponse := posts.Post{
		Id:        post.ID,
		Title:     post.Title,
		PostText:  post.PostText,
		Img:       post.Img,
		ReadTime:  post.ReadTime,
		LikedBy:   post.LikedBy,
		UserId:    post.UserID,
		UserName:  post.UserName,
		CreatedAt: timestamppb.New(post.CreatedAt),
	}

	res := &posts.GetPostResponse{Post: &postResponse}
	return res, nil
}

// Deletes a single post by id
func (u *PostsService) DeletePost(ctx context.Context, req *posts.GetPostRequest) (*posts.DeletePostResponse, error) {
	id := req.GetId()

	err := u.Models.Posts.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	res := &posts.DeletePostResponse{Message: "post deleted successfully"}
	return res, nil
}

// Deletes all the posts for a single user
func (u *PostsService) DeletePostsForUser(ctx context.Context, req *posts.GetPostRequest) (*posts.DeletePostResponse, error) {
	id := req.GetId()

	err := u.Models.Posts.DeleteForUser(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	res := &posts.DeletePostResponse{Message: "post deleted successfully"}
	return res, nil
}

// Likes a single post by id
func (u *PostsService) LikePost(ctx context.Context, req *posts.LikePostRequest) (*posts.GetPostResponse, error) {
	id := req.GetId()
	userID := req.GetUserId()

	// Check if a post with provided id exists.
	post, err := u.Application.Models.Posts.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	err = u.Application.Models.Posts.AddLike(post, userID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("bad request: %v", err.Error()))
	}

	postResponse := &posts.Post{
		Id:        post.ID,
		Title:     post.Title,
		PostText:  post.PostText,
		Img:       post.Img,
		ReadTime:  post.ReadTime,
		LikedBy:   post.LikedBy,
		UserId:    post.UserID,
		UserName:  post.UserName,
		CreatedAt: timestamppb.New(post.CreatedAt),
	}

	res := &posts.GetPostResponse{Post: postResponse}
	return res, nil
}

// Dislikes a single post by id
func (u *PostsService) DislikePost(ctx context.Context, req *posts.LikePostRequest) (*posts.GetPostResponse, error) {
	id := req.GetId()
	userID := req.GetUserId()

	// Check if a post with provided id exists.
	post, err := u.Application.Models.Posts.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	err = u.Application.Models.Posts.RemoveLike(post, userID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("bad request: %v", err.Error()))
	}

	postResponse := &posts.Post{
		Id:        post.ID,
		Title:     post.Title,
		PostText:  post.PostText,
		Img:       post.Img,
		ReadTime:  post.ReadTime,
		LikedBy:   post.LikedBy,
		UserId:    post.UserID,
		UserName:  post.UserName,
		CreatedAt: timestamppb.New(post.CreatedAt),
	}

	res := &posts.GetPostResponse{Post: postResponse}
	return res, nil
}

// Gets all the comments for a single post
func (u *PostsService) GetCommentsForPost(ctx context.Context, req *posts.GetPostRequest) (*posts.GetCommentsForPostResponse, error) {
	id := req.GetId()

	result, err := u.Models.Comments.GetAllForPost(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	// Get comments ready to send via gRPC
	grpcComments := make([]*posts.Comment, len(result))
	for i, comment := range result {
		grpcComments[i] = ConvertCommentToGRPCResponse(comment)
	}

	res := &posts.GetCommentsForPostResponse{Comments: grpcComments}
	return res, nil
}

// Gets comments ready to send via gRPC
func ConvertCommentToGRPCResponse(comment *data.Comment) *posts.Comment {
	return &posts.Comment{
		Id:       comment.ID,
		Text:     comment.Text,
		UserId:   comment.UserID,
		UserName: comment.UserName,
		PostId:   comment.PostID,
	}

}

// Creates a comment
func (u *PostsService) CreateComment(ctx context.Context, req *posts.CreateCommentRequest) (*posts.CreateCommentResponse, error) {
	text := req.GetText()
	userID := req.GetUserId()
	postID := req.GetPostId()

	comment := &data.Comment{
		Text:   strings.TrimSpace(text),
		UserID: userID,
		PostID: postID,
	}

	v := validator.New()
	data.ValidateComment(v, comment)

	err := u.Application.checkValidationStatus(v)
	if err != nil {
		return nil, err
	}

	err = u.Models.Comments.Insert(comment)
	if err != nil {
		err := u.Application.ServerErrorResponse(err.Error())
		if err != nil {
			return nil, err
		}
	}

	commentResponse := posts.Comment{
		Id:       comment.ID,
		Text:     comment.Text,
		UserId:   comment.UserID,
		UserName: comment.UserName,
		PostId:   comment.PostID,
	}

	res := &posts.CreateCommentResponse{Comment: &commentResponse}
	return res, nil
}

// Deletes a single comment by id
func (u *PostsService) DeleteComment(ctx context.Context, req *posts.GetPostRequest) (*posts.DeletePostResponse, error) {
	id := req.GetId()

	err := u.Models.Comments.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	res := &posts.DeletePostResponse{Message: "comment deleted successfully"}
	return res, nil
}

// Starts listening to gRPC calls
func (app *application) gRPCListen() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", app.config.gRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	s := grpc.NewServer()
	posts.RegisterPostsServiceServer(s, &PostsService{Models: app.Models, DB: app.DB, Application: app})

	// Health Check
	dbSystem := "database" // Specify the system as "database"
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(s, healthcheck)

	go func() {
		// Asynchronously inspect database status and toggle serving status as needed
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			// Perform a health check on the database
			err := app.DB.PingContext(ctx)
			cancel()

			// Set the serving status based on the database health check result
			if err != nil {
				healthcheck.SetServingStatus(dbSystem, healthpb.HealthCheckResponse_NOT_SERVING)
			} else {
				healthcheck.SetServingStatus(dbSystem, healthpb.HealthCheckResponse_SERVING)
			}

			// Change this in future, so that check happens only when errors occur
			time.Sleep(time.Second * 60)
		}
	}()

	log.Printf("gRPC server started on port %v", app.config.gRPCPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
