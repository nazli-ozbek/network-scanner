package repository

import (
	"database/sql"
	"errors"
	"network-scanner/logger"
	"network-scanner/model"

	"strings"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
}

type SQLiteUserRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewSQLiteUserRepository(db *sql.DB, logger logger.Logger) *SQLiteUserRepository {
	if err := ensureUsersTable(db); err != nil {
		logger.Error("failed to create users table", err)
	}
	return &SQLiteUserRepository{db: db, logger: logger}
}

func ensureUsersTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		);
	`)
	return err
}

func (r *SQLiteUserRepository) Create(user *model.User) error {
	user.ID = uuid.New().String()
	_, err := r.db.Exec("INSERT INTO users (id, username, password) VALUES (?, ?, ?)",
		user.ID, user.Username, user.Password)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			return errors.New("user already exists")
		}
		return err
	}
	return nil
}

func (r *SQLiteUserRepository) FindByUsername(username string) (*model.User, error) {
	row := r.db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username)

	var user model.User
	if err := row.Scan(&user.ID, &user.Username, &user.Password); err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
