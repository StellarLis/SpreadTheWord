package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"user_service/internal/amqp"
	"user_service/internal/models"
	"user_service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceIn interface {
	Register(username, password string) (string, error)
	SignIn(username, password string) (string, error)
	UpdateAvatar(newAvatar []byte, userId int) error
	GetDataFromToken(token string) (int, string, error)
}

type UserService struct {
	UserRepository repository.UserRepositoryIn
	JwtSecretKey   string
	Amqp           *amqp.Amqp
}

var _ UserServiceIn = &UserService{}

func (us *UserService) Register(username, password string) (string, error) {
	if err := us.validate(username, password); err != nil {
		return "", err
	}
	candidate := us.UserRepository.FindUserByUsername(username)
	if candidate != nil {
		return "", errors.New("user with that username already exists")
	}
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	hashedPassword := string(hashBytes)
	userId, err := us.UserRepository.CreateUser(username, hashedPassword)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	message := models.CreatePhotoMessage{UserId: userId, Username: username}
	bytesJson, err := json.Marshal(&message)
	if err != nil {
		return "", err
	}
	err = us.Amqp.Channel.PublishWithContext(ctx,
		"",
		us.Amqp.Queue.Name,
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        bytesJson,
		},
	)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":    userId,
		"username":  username,
		"timestamp": time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(us.JwtSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (us *UserService) SignIn(username, password string) (string, error) {
	if err := us.validate(username, password); err != nil {
		return "", err
	}
	candidate := us.UserRepository.FindUserByUsername(username)
	if candidate == nil {
		return "", errors.New("invalid username or password")
	}
	err := bcrypt.CompareHashAndPassword([]byte(candidate.Password),
		[]byte(password))
	if err != nil {
		return "", errors.New("invalid username or password")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":    candidate.Id,
		"username":  username,
		"timestamp": time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(us.JwtSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (us *UserService) UpdateAvatar(newAvatar []byte, userId int) error {
	err := us.UserRepository.UpdateAvatar(newAvatar, userId)
	if err != nil {
		return err
	}
	return nil
}

func (us *UserService) GetDataFromToken(token string) (int, string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(us.JwtSecretKey), nil
	})
	if err != nil {
		return 0, "", err
	}
	return int(claims["userId"].(float64)), claims["username"].(string), nil
}

func (us *UserService) validate(username, password string) error {
	if len(username) < 6 || len(username) > 26 {
		return errors.New("username's length should be bigger than 5 and less than 27")
	}
	if len(password) < 6 || len(password) > 26 {
		return errors.New("password's length should be bigger than 5 and less than 27")
	}
	return nil
}
