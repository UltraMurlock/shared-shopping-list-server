package service

import (
	"context"
	"errors"
	"fmt"
	"shared-shopping-list-server/internal/domain"
	"shared-shopping-list-server/internal/repository"
	"shared-shopping-list-server/internal/shared/auth"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LoginResult struct {
	User      *domain.User
	Token     string
	ExpiresAt time.Time
}

type AuthService struct {
	pool     *pgxpool.Pool
	userRepo *repository.UserRepo
	jwt      *auth.JWTManager
}

func NewAuthService(
	pool *pgxpool.Pool,
	userRepo *repository.UserRepo,
	jwt *auth.JWTManager,
) *AuthService {
	return &AuthService{pool: pool, userRepo: userRepo, jwt: jwt}
}

func (s *AuthService) Login(
	ctx context.Context,
	login, password string,
) (*LoginResult, error) {
	login = strings.TrimSpace(login)
	password = strings.TrimSpace(password)
	if login == "" || password == "" {
		return nil, fmt.Errorf("auth service: login: %w", domain.ErrValidationError)
	}

	u, err := s.userRepo.GetByLogin(ctx, s.pool, login)

	if errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("auth service: login: %w", domain.ErrInvalidCredentials)
	}

	if err != nil {
		return nil, fmt.Errorf("auth service: login: %w", err)
	}

	if err := auth.CheckPassword(u.PasswordHash, password); err != nil {
		return nil, fmt.Errorf("auth service: login: %w", domain.ErrInvalidCredentials)
	}

	token, expiresAt, err := s.jwt.Generate(u.ID, u.Login)
	if err != nil {
		return nil, fmt.Errorf("auth service: login: %w", err)
	}

	return &LoginResult{
		User:      u,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
