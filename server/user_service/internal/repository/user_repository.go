package repository

import (
	"fmt"
	"user_service/internal/models"

	"github.com/jmoiron/sqlx"
)

type UserRepositoryIn interface {
	CreateUser(username, password string) (int, error)
	FindUserByUsername(username string) *models.UserDb
	UpdateAvatar(newAvatar []byte, userId int) error
}

type UserRepository struct {
	Db *sqlx.DB
}

var _ UserRepositoryIn = &UserRepository{}

func (ur *UserRepository) CreateUser(username, password string) (int, error) {
	const op = "repository.CreateUser"

	var userId int
	stmt, err := ur.Db.Prepare(`INSERT INTO app_user (username, password, avatar) VALUES ($1, $2, $3) RETURNING id`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	err = stmt.QueryRow(username, password, nil).Scan(&userId)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	stmt.Close()
	return userId, nil
}

func (ur *UserRepository) FindUserByUsername(username string) *models.UserDb {
	var candidate models.UserDb
	stmt, err := ur.Db.Prepare("SELECT id, username, password FROM app_user WHERE username=$1")
	if err != nil {
		return nil
	}
	err = stmt.QueryRow(username).Scan(&candidate.Id, &candidate.Username, &candidate.Password)
	if err != nil {
		return nil
	}
	stmt.Close()
	return &candidate
}

func (ur *UserRepository) UpdateAvatar(newAvatar []byte, userId int) error {
	_, err := ur.Db.Exec("UPDATE app_user SET avatar = $1 WHERE id = $2", newAvatar, userId)
	if err != nil {
		return err
	}
	return nil
}
