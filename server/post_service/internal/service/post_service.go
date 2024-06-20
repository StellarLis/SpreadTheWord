package service

import (
	"fmt"
	"post_service/internal/model"
	"post_service/internal/repository"
)

type PostServiceIn interface {
	GetPost(postId int) (*model.PostDb, error)
	NewPost(message string, userId int) error
	UpdatePost(postId int, newMessage string, userId int) error
	DeletePost(postId int, userId int) error
}

type PostService struct {
	PostRepository *repository.PostRepository
}

var _ PostServiceIn = &PostService{}

func (p *PostService) GetPost(postId int) (*model.PostDb, error) {
	const op = "service.GetPost"

	postDb, err := p.PostRepository.GetPostById(postId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return postDb, nil
}

func (p *PostService) NewPost(message string, userId int) error {
	const op = "service.NewPost"

	err := p.PostRepository.NewPost(message, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (p *PostService) UpdatePost(postId int, newMessage string, userId int) error {
	const op = "service.UpdatePost"

	postDb, err := p.PostRepository.GetPostById(postId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if postDb.UserId != userId {
		return fmt.Errorf("you are not an owner of this post")
	}
	err = p.PostRepository.UpdatePost(postId, newMessage)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

func (p *PostService) DeletePost(postId int, userId int) error {
	const op = "service.DeletePost"

	postDb, err := p.PostRepository.GetPostById(postId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if postDb.UserId != userId {
		return fmt.Errorf("you are not an owner of this post")
	}
	err = p.PostRepository.DeletePost(postId)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}
