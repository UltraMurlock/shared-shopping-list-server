package auth

import (
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Claims struct {
	Login string `json:"login,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTManager(secret []byte, ttl time.Duration) (*JWTManager, error) {
	cfg := Config{Secret: secret, TTL: ttl}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &JWTManager{secret: secret, ttl: ttl}, nil
}

func (m *JWTManager) Generate(userID uuid.UUID, login string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)

	claims := Claims{
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth: sign token: %w", err)
	}

	return signed, expiresAt, nil
}

func (m *JWTManager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method: %v", t.Header["alg"])
		}

		return m.secret, nil
	})

	if errors.Is(err, jwt.ErrTokenExpired) {
		return nil, fmt.Errorf("auth: parse token: %w", ErrExpiredToken)
	}

	if err != nil {
		return nil, fmt.Errorf("auth: parse token: %w", ErrInvalidToken)
	}

	return claims, nil
}

func (c *Claims) UserID() (uuid.UUID, error) {
	id, err := uuid.Parse(c.Subject)
	if err != nil {
		return uuid.Nil(), fmt.Errorf("auth: parse subject: %w", ErrInvalidToken)
	}

	return id, nil
}
