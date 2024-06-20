package repository

import (
	"fmt"
	"post_service/internal/model"

	"github.com/jmoiron/sqlx"
)

type PostRepositoryIn interface {
	GetPostById(postId int) (*model.PostDb, error)
	NewPost(message string, userId int) error
	UpdatePost(postId int, newMessage string) error
	DeletePost(postId int) error
}

type PostRepository struct {
	Db *sqlx.DB
}

var _ PostRepositoryIn = &PostRepository{}

func (p *PostRepository) GetPostById(postId int) (*model.PostDb, error) {
	const op = "repository.GetPostById"

	var post *model.PostDb = &model.PostDb{}
	err := p.Db.QueryRow(`SELECT up.id, message, user_id, username, avatar FROM user_post AS up
	 INNER JOIN app_user AS au ON up.user_id = au.id WHERE up.id = $1`,
		postId).Scan(&post.PostId, &post.Message, &post.UserId, &post.Username, &post.Avatar)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return post, nil
}

func (p *PostRepository) NewPost(message string, userId int) error {
	const op = "repository.NewPost"

	stmt, err := p.Db.Prepare("INSERT INTO user_post (message, user_id) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(message, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	stmt.Close()
	return nil
}

func (p *PostRepository) UpdatePost(postId int, newMessage string) error {
	const op = "repository.UpdatePost"

	stmt, err := p.Db.Prepare("UPDATE user_post SET message = $1 WHERE id = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(newMessage, postId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	stmt.Close()
	return nil
}

func (p *PostRepository) DeletePost(postId int) error {
	const op = "repository.DeletePost"

	stmt, err := p.Db.Prepare("DELETE FROM user_post WHERE id = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(postId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	stmt.Close()
	return nil
}
