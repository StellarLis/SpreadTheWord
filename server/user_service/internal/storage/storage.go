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
CREATE TABLE IF NOT EXISTS app_user(
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
	avatar bytea
)
`

func New() *Storage {
	connectionString := fmt.Sprintf(
		"user=%v password=%v dbname=%v port=%v host=host.docker.internal sslmode=disable",
		os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
	)
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		logrus.Fatalln(err)
	}
	db.MustExec(schema)

	return &Storage{Db: db}
}

func (s *Storage) Stop() {
	const op = "storage.Stop"
	logrus.WithField("op", op).Info("closing database instance")
	s.Db.Close()
}
