package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

type AuthService struct {
	db          *sql.DB
	jwtSecret   []byte
	tokenExpiry time.Duration
}

func NewAuthService(db *sql.DB, jwtSecret string, tokenExpiryHours int) *AuthService {
	return &AuthService{
		db:          db,
		jwtSecret:   []byte(jwtSecret),
		tokenExpiry: time.Duration(tokenExpiryHours) * time.Hour,
	}
}

type LoginResult struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	var userID int64
	var passwordHash string

	err := s.db.QueryRowContext(ctx,
		"SELECT id, password_hash FROM users WHERE username = ?", username,
	).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	expiresAt := time.Now().Add(s.tokenExpiry)
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	return &LoginResult{
		Token:     tokenString,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrTokenExpired
		}
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}

	sub, err := claims.GetSubject()
	if err != nil {
		subFloat, ok := claims["sub"].(float64)
		if !ok {
			return 0, ErrInvalidToken
		}
		return int64(subFloat), nil
	}
	_ = sub

	subFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, ErrInvalidToken
	}
	return int64(subFloat), nil
}
