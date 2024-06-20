package repository

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	Db *sqlx.DB
}

func New() *Repository {
	connectionString := fmt.Sprintf(
		"user=%v password=%v dbname=%v port=%v host=host.docker.internal sslmode=disable",
		os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
	)
	logrus.Infoln(connectionString)
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		logrus.Fatalln("error while creating storage", err)
	}

	return &Repository{Db: db}
}

func (r *Repository) UpdateDbAvatar(avatarBytes []byte, userId int) {
	_, err := r.Db.Exec("UPDATE app_user SET avatar = $1 WHERE id = $2", avatarBytes, userId)
	if err != nil {
		logrus.Fatalln(err)
	}
}
