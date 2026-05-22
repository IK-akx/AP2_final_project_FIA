package usecase

import (
	"context"
	"testing"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
)

type mockUserRepo struct {
	users map[string]domain.User
}

func (m *mockUserRepo) Save(ctx context.Context, u domain.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return &u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u domain.User) error {
	m.users[u.ID] = u
	return nil
}

func TestRegisterUseCase_Execute_Success(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	uc := NewRegisterUseCase(repo)

	user, err := uc.Execute(context.Background(), "test@example.com", "password123", "John", "Doe")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
	if user.FirstName != "John" {
		t.Errorf("Expected first name John, got %s", user.FirstName)
	}
	if user.LastName != "Doe" {
		t.Errorf("Expected last name Doe, got %s", user.LastName)
	}
	if user.Role != "user" {
		t.Errorf("Expected role user, got %s", user.Role)
	}
	if user.Phone != "" {
		t.Errorf("Expected empty phone, got %s", user.Phone)
	}
}

func TestRegisterUseCase_Execute_EmptyEmail(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	uc := NewRegisterUseCase(repo)

	_, err := uc.Execute(context.Background(), "", "password123", "John", "Doe")
	if err != ErrEmptyField {
		t.Errorf("Expected ErrEmptyField, got %v", err)
	}
}

func TestRegisterUseCase_Execute_EmptyPassword(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	uc := NewRegisterUseCase(repo)

	_, err := uc.Execute(context.Background(), "test@example.com", "", "John", "Doe")
	if err != ErrEmptyField {
		t.Errorf("Expected ErrEmptyField, got %v", err)
	}
}

func TestRegisterUseCase_Execute_EmptyFirstName(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	uc := NewRegisterUseCase(repo)

	_, err := uc.Execute(context.Background(), "test@example.com", "password123", "", "Doe")
	if err != ErrEmptyField {
		t.Errorf("Expected ErrEmptyField, got %v", err)
	}
}

func TestRegisterUseCase_Execute_InvalidEmail(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	uc := NewRegisterUseCase(repo)

	_, err := uc.Execute(context.Background(), "not-an-email", "password123", "John", "Doe")
	if err != ErrInvalidEmail {
		t.Errorf("Expected ErrInvalidEmail, got %v", err)
	}
}

func TestRegisterUseCase_Execute_DuplicateEmail(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	uc := NewRegisterUseCase(repo)

	_, err := uc.Execute(context.Background(), "duplicate@example.com", "password123", "John", "Doe")
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	_, err = uc.Execute(context.Background(), "duplicate@example.com", "password456", "Jane", "Smith")
	if err != ErrUserExists {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}
}
