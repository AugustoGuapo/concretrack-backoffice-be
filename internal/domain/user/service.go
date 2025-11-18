package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidRole = errors.New("invalid role")

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Login(username string, password string) (User, error) {
	u, err := s.repo.GetByUsername(username)

	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	if u.Role != "admin" {
		return User{}, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return User{}, ErrInvalidRole
	}

	return *u, nil
} 