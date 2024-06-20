package storage

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type Storage struct {
	Db *sqlx.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS user_post(
	id SERIAL PRIMARY KEY,
	message TEXT NOT NULL,
	user_id INTEGER REFERENCES app_user (id)
);
`

func New() *Storage {
	connectionString := fmt.Sprintf(
		"user=%v password=%v dbname=%v port=%v host=host.docker.internal sslmode=disable",
		os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
	)
	logrus.Infoln(connectionString)
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		logrus.Fatalln("error while creating storage", err)
	}
	db.MustExec(schema)

	return &Storage{Db: db}
}

func (s *Storage) Stop() {
	err := s.Db.Close()
	if err != nil {
		logrus.Errorf("error while closing db instance")
	}
}
