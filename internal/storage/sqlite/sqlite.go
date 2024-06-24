package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexPavlikov/sso/internal/domain/models"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	// TODO: ...
	return 1, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	// TODO: ...
	return models.User{}, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	// TODO: ...
	return true, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	return models.App{}, nil
}
