package services_test

import (
	"context"
	"testing"

	"github.com/beohoang98/moneyapp/internal/services"
)

func TestLogin_ValidCredentials(t *testing.T) {
	db := setupTestDB(t)
	as := services.NewAuthService(db, "test-secret", 24)

	result, err := as.Login(context.Background(), "admin", "changeme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if result.ExpiresAt == "" {
		t.Fatal("expected non-empty expires_at")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	db := setupTestDB(t)
	as := services.NewAuthService(db, "test-secret", 24)

	_, err := as.Login(context.Background(), "admin", "wrongpassword")
	if err != services.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_InvalidUsername(t *testing.T) {
	db := setupTestDB(t)
	as := services.NewAuthService(db, "test-secret", 24)

	_, err := as.Login(context.Background(), "nonexistent", "changeme")
	if err != services.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestValidateToken_Valid(t *testing.T) {
	db := setupTestDB(t)
	as := services.NewAuthService(db, "test-secret", 24)

	result, _ := as.Login(context.Background(), "admin", "changeme")
	userID, err := as.ValidateToken(result.Token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 1 {
		t.Fatalf("expected userID 1, got %d", userID)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	db := setupTestDB(t)
	as := services.NewAuthService(db, "test-secret", 24)

	_, err := as.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	db := setupTestDB(t)
	as1 := services.NewAuthService(db, "secret-1", 24)
	as2 := services.NewAuthService(db, "secret-2", 24)

	result, _ := as1.Login(context.Background(), "admin", "changeme")
	_, err := as2.ValidateToken(result.Token)
	if err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}
