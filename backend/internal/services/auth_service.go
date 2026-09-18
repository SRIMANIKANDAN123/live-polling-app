package services

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"

	"live-polling-app/backend/internal/models"
	"live-polling-app/backend/internal/repositories"
	"live-polling-app/backend/internal/utils"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

var (
	ErrInvalidEmail = errors.New("please provide a valid email address")
	ErrWeakPassword = errors.New("password must be at least 8 characters")
	ErrEmailInUse   = errors.New("an account with this email already exists")
	ErrInvalidCreds = errors.New("invalid email or password")
)

type AuthService struct {
	users     *repositories.UserRepository
	jwtSecret string
}

func NewAuthService(users *repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*models.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !emailRegex.MatchString(email) {
		return nil, "", ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, "", ErrWeakPassword
	}

	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, "", ErrEmailInUse
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	user := &models.User{Email: email, PasswordHash: hash}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := utils.GenerateToken(s.jwtSecret, user.ID.Hex(), user.Email)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, "", ErrInvalidCreds
		}
		return nil, "", err
	}
	if !utils.CheckPassword(user.PasswordHash, password) {
		return nil, "", ErrInvalidCreds
	}
	token, err := utils.GenerateToken(s.jwtSecret, user.ID.Hex(), user.Email)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}
