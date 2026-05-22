package usecase

import (
	"context"
	"testing"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
)

type mockCache struct {
	data map[string][]byte
}

func (m *mockCache) Set(ctx context.Context, key string, value []byte, ttlSeconds int) error {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}
	m.data[key] = value
	return nil
}

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	if m.data == nil {
		return nil, context.Canceled
	}
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, context.Canceled
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	if m.data != nil {
		delete(m.data, key)
	}
	return nil
}

func TestProfileUseCase_GetProfile_Success(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	cache := &mockCache{}
	profileUC := NewProfileUseCase(repo, cache)
	registerUC := NewRegisterUseCase(repo)

	user, err := registerUC.Execute(context.Background(), "profile@example.com", "pass123", "Profile", "User")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	retrieved, err := profileUC.GetProfile(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, retrieved.ID)
	}
	if retrieved.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrieved.Email)
	}
	if retrieved.FirstName != "Profile" {
		t.Errorf("Expected first name Profile, got %s", retrieved.FirstName)
	}
}

func TestProfileUseCase_GetProfile_NotFound(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	cache := &mockCache{}
	profileUC := NewProfileUseCase(repo, cache)

	_, err := profileUC.GetProfile(context.Background(), "non-existent-id")
	if err != ErrProfileNotFound {
		t.Errorf("Expected ErrProfileNotFound, got %v", err)
	}
}

func TestProfileUseCase_UpdateProfile_Success(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	cache := &mockCache{}
	profileUC := NewProfileUseCase(repo, cache)
	registerUC := NewRegisterUseCase(repo)

	user, err := registerUC.Execute(context.Background(), "update@example.com", "pass123", "OldName", "OldLast")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	updated, err := profileUC.UpdateProfile(context.Background(), user.ID, "NewName", "NewLast", "+777777777")
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}

	if updated.FirstName != "NewName" {
		t.Errorf("Expected first name NewName, got %s", updated.FirstName)
	}
	if updated.LastName != "NewLast" {
		t.Errorf("Expected last name NewLast, got %s", updated.LastName)
	}
	if updated.Phone != "+777777777" {
		t.Errorf("Expected phone +777777777, got %s", updated.Phone)
	}
}

func TestProfileUseCase_UpdateProfile_NotFound(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	cache := &mockCache{}
	profileUC := NewProfileUseCase(repo, cache)

	_, err := profileUC.UpdateProfile(context.Background(), "non-existent-id", "Name", "Last", "123")
	if err != ErrProfileNotFound {
		t.Errorf("Expected ErrProfileNotFound, got %v", err)
	}
}

func TestProfileUseCase_UpdateProfile_PartialUpdate(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	cache := &mockCache{}
	profileUC := NewProfileUseCase(repo, cache)
	registerUC := NewRegisterUseCase(repo)

	user, err := registerUC.Execute(context.Background(), "partial@example.com", "pass123", "Original", "User")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	updated, err := profileUC.UpdateProfile(context.Background(), user.ID, "UpdatedName", "", "+123456789")
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}

	if updated.FirstName != "UpdatedName" {
		t.Errorf("Expected first name UpdatedName, got %s", updated.FirstName)
	}
	if updated.LastName != "" {
		t.Errorf("Expected last name empty, got %s", updated.LastName)
	}
	if updated.Phone != "+123456789" {
		t.Errorf("Expected phone +123456789, got %s", updated.Phone)
	}
}

func TestProfileUseCase_GetProfile_FromCache(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	cache := &mockCache{}
	profileUC := NewProfileUseCase(repo, cache)
	registerUC := NewRegisterUseCase(repo)

	user, err := registerUC.Execute(context.Background(), "cache@example.com", "pass123", "Cache", "Test")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	first, err := profileUC.GetProfile(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("First GetProfile failed: %v", err)
	}

	user.FirstName = "Changed"
	_ = repo.Update(context.Background(), *user)

	second, err := profileUC.GetProfile(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Second GetProfile failed: %v", err)
	}

	if second.FirstName != "Cache" {
		t.Errorf("Expected cached first name Cache, got %s", second.FirstName)
	}
	_ = first
}
