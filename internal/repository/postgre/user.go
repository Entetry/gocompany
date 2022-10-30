package postgre

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"entetry/gotest/internal/model"
)

// UserRepository user repository interface
type UserRepository interface {
	Create(ctx context.Context, username, pwdHash, email string) (uuid.UUID, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

// User User postgres repository struct
type User struct {
	db *pgxpool.Pool
}

// NewUserRepository creates new user repository object
func NewUserRepository(db *pgxpool.Pool) *User {
	return &User{
		db: db,
	}
}

// Create insert user record in db
func (u *User) Create(ctx context.Context, username, pwdHash, email string) (uuid.UUID, error) {
	var user model.User
	user.ID = uuid.New()
	user.PasswordHash = pwdHash
	user.Email = email
	user.Username = username
	_, err := u.db.Exec(ctx, `INSERT INTO users (id, username, email, passwordHash) VALUES ($1, $2, $3, $4)`,
		user.ID, user.Username, user.Email, user.PasswordHash)
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot create User: %v", err)
	}
	return user.ID, nil
}

// GetByUsername return user by his username
func (u *User) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := u.db.QueryRow(ctx,
		`SELECT id, username, email, passwordHash FROM users WHERE username = $1`, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error in GetByUsername: %v", err)
	}
	return &user, nil
}
