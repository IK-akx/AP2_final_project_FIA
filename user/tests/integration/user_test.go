//go:build integration
// +build integration

package integration

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/adapters/postgres"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5435/user_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	testDB, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Printf("Integration tests skipped: cannot connect to test DB: %v", err)
		os.Exit(0)
	}
	defer testDB.Close()

	_, _ = testDB.Exec(ctx, "TRUNCATE users, notification_logs RESTART IDENTITY CASCADE")

	code := m.Run()

	_, _ = testDB.Exec(ctx, "TRUNCATE users, notification_logs RESTART IDENTITY CASCADE")

	os.Exit(code)
}

func TestIntegration_UserRegistrationAndLogin(t *testing.T) {
	userRepo := postgres.NewUserRepo(testDB)
	registerUC := usecase.NewRegisterUseCase(userRepo)
	loginUC := usecase.NewLoginUseCase(userRepo)

	email := "integration@example.com"
	password := "testpass123"
	firstName := "Integration"
	lastName := "Test"

	user, err := registerUC.Execute(context.Background(), email, password, firstName, lastName)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
	if user.FirstName != firstName {
		t.Errorf("Expected first name %s, got %s", firstName, user.FirstName)
	}

	result, err := loginUC.Execute(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.User.ID != user.ID {
		t.Errorf("User ID mismatch: expected %s, got %s", user.ID, result.User.ID)
	}
	if result.Token == "" {
		t.Error("Expected token, got empty string")
	}

	t.Logf("Registration and login successful for user: %s", user.ID)
}

func TestIntegration_RegisterDuplicateEmail(t *testing.T) {
	userRepo := postgres.NewUserRepo(testDB)
	registerUC := usecase.NewRegisterUseCase(userRepo)

	email := "duplicate@example.com"

	_, err := registerUC.Execute(context.Background(), email, "pass1", "First", "User")
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	_, err = registerUC.Execute(context.Background(), email, "pass2", "Second", "User")
	if err != usecase.ErrUserExists {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}

	t.Log("Duplicate email correctly rejected")
}

func TestIntegration_LoginWrongPassword(t *testing.T) {
	userRepo := postgres.NewUserRepo(testDB)
	registerUC := usecase.NewRegisterUseCase(userRepo)
	loginUC := usecase.NewLoginUseCase(userRepo)

	email := "wrongpass@example.com"
	password := "correctpass"

	_, err := registerUC.Execute(context.Background(), email, password, "Test", "User")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	_, err = loginUC.Execute(context.Background(), email, "wrongpass")
	if err != usecase.ErrWrongPassword {
		t.Errorf("Expected ErrWrongPassword, got %v", err)
	}

	t.Log("Wrong password correctly rejected")
}

func TestIntegration_GetAndUpdateProfile(t *testing.T) {
	userRepo := postgres.NewUserRepo(testDB)
	registerUC := usecase.NewRegisterUseCase(userRepo)
	profileUC := usecase.NewProfileUseCase(userRepo, nil) // nil cache для теста

	user, err := registerUC.Execute(context.Background(), "profile@example.com", "pass", "OldName", "OldLast")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	profile, err := profileUC.GetProfile(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}
	if profile.FirstName != "OldName" {
		t.Errorf("Expected OldName, got %s", profile.FirstName)
	}

	updated, err := profileUC.UpdateProfile(context.Background(), user.ID, "NewName", "NewLast", "+123456789")
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}
	if updated.FirstName != "NewName" {
		t.Errorf("Expected NewName, got %s", updated.FirstName)
	}
	if updated.Phone != "+123456789" {
		t.Errorf("Expected phone +123456789, got %s", updated.Phone)
	}

	t.Log("Profile update successful")
}

func TestIntegration_NotificationLogSave(t *testing.T) {
	notificationRepo := postgres.NewNotificationRepo(testDB)

	logEntry := domain.NotificationLog{
		ID:       "test-id-123",
		UserID:   "test-user-id",
		Type:     "order.created",
		Subject:  "Test Subject",
		Body:     "Test Body",
		SentAt:   time.Now(),
		Status:   "sent",
		ErrorMsg: "",
	}

	err := notificationRepo.Save(context.Background(), logEntry)
	if err != nil {
		t.Fatalf("Failed to save notification log: %v", err)
	}

	var count int
	err = testDB.QueryRow(context.Background(), "SELECT COUNT(*) FROM notification_logs WHERE id = $1", "test-id-123").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}

	t.Log("Notification log saved successfully")
}
