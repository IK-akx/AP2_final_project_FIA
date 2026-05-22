package usecase

import (
	"context"
	"testing"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
)

func TestLoginUseCase_Execute_Success(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	registerUC := NewRegisterUseCase(repo)
	loginUC := NewLoginUseCase(repo)

	registeredUser, err := registerUC.Execute(context.Background(), "login@example.com", "correctpass", "Login", "User")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	result, err := loginUC.Execute(context.Background(), "login@example.com", "correctpass")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.User.ID != registeredUser.ID {
		t.Errorf("Expected user ID %s, got %s", registeredUser.ID, result.User.ID)
	}
	if result.User.Email != "login@example.com" {
		t.Errorf("Expected email login@example.com, got %s", result.User.Email)
	}
	if result.Token == "" {
		t.Error("Expected token, got empty string")
	}
}

func TestLoginUseCase_Execute_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	loginUC := NewLoginUseCase(repo)

	_, err := loginUC.Execute(context.Background(), "nonexistent@example.com", "password123")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestLoginUseCase_Execute_WrongPassword(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	registerUC := NewRegisterUseCase(repo)
	loginUC := NewLoginUseCase(repo)

	_, err := registerUC.Execute(context.Background(), "wrongpass@example.com", "correctpass", "Test", "User")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	_, err = loginUC.Execute(context.Background(), "wrongpass@example.com", "wrongpass")
	if err != ErrWrongPassword {
		t.Errorf("Expected ErrWrongPassword, got %v", err)
	}
}

func TestLoginUseCase_GenerateToken(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	loginUC := NewLoginUseCase(repo)

	token, err := loginUC.GenerateToken("user-123", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Error("Expected token, got empty string")
	}
}

func TestLoginUseCase_ValidateToken_Valid(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	loginUC := NewLoginUseCase(repo)

	token, err := loginUC.GenerateToken("user-456", "user")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	userID, role, err := loginUC.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if userID != "user-456" {
		t.Errorf("Expected userID user-456, got %s", userID)
	}
	if role != "user" {
		t.Errorf("Expected role user, got %s", role)
	}
}

func TestLoginUseCase_ValidateToken_Invalid(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	loginUC := NewLoginUseCase(repo)

	_, _, err := loginUC.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestLoginUseCase_ValidateToken_Empty(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	loginUC := NewLoginUseCase(repo)

	_, _, err := loginUC.ValidateToken("")
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}
}
