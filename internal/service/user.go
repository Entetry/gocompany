// Package service contains service structs
package service

import (
	"context"
	"github.com/Entetry/gocompany/internal/repository"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/Entetry/gocompany/internal/model"
)

// User service struct
type User struct {
	userRepository repository.UserRepository
}

// NewUserService creates new User service
func NewUserService(userRepository repository.UserRepository) *User {
	return &User{
		userRepository: userRepository}
}

// GetByUsername return user by its username
func (u *User) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return u.userRepository.GetByUsername(ctx, username)
}

// Create user
func (u *User) Create(ctx context.Context, username, password, email string) (uuid.UUID, error) {
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		log.Println(err)
		return uuid.Nil, err
	}

	return u.userRepository.Create(ctx, username, string(pwdHash), email)
}
