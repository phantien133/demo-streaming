package auth

import (
	"context"
	"errors"
	"strings"

	"demo-streaming/internal/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type GormLoginService struct {
	db *gorm.DB
}

func NewGormLoginService(db *gorm.DB) *GormLoginService {
	return &GormLoginService{db: db}
}

func (s *GormLoginService) Execute(ctx context.Context, input LoginInput) (LoginOutput, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := input.Password
	if email == "" || password == "" {
		return LoginOutput{}, ErrInvalidCredentials
	}

	var user database.User
	if err := s.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LoginOutput{}, ErrInvalidCredentials
		}
		return LoginOutput{}, err
	}
	if user.PasswordHash == "" {
		return LoginOutput{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginOutput{}, ErrInvalidCredentials
	}

	return LoginOutput{
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}

