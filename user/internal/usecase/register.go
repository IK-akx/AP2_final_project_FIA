package usecase

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/ports"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail = errors.New("invalid email")
	ErrUserExists   = errors.New("user already exists")
	ErrEmptyField   = errors.New("required field is empty")
)

type OrderServiceClient interface {
	InitBalance(ctx context.Context, userID string) error
}

type RegisterUseCase struct {
	repo        ports.UserRepository
	orderClient OrderServiceClient
}

func NewRegisterUseCase(repo ports.UserRepository, orderClient ...OrderServiceClient) *RegisterUseCase {
	var client OrderServiceClient
	if len(orderClient) > 0 {
		client = orderClient[0]
	}
	return &RegisterUseCase{
		repo:        repo,
		orderClient: client,
	}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, email, password, firstName, lastName string) (*domain.User, error) {
	if email == "" || password == "" || firstName == "" {
		return nil, ErrEmptyField
	}
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	existing, _ := uc.repo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        "",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := uc.repo.Save(ctx, user); err != nil {
		return nil, err
	}

	if uc.orderClient != nil {
		if err := uc.orderClient.InitBalance(ctx, user.ID); err != nil {
			log.Printf("Warning: failed to init balance for user %s: %v", user.ID, err)
		}
	}

	return &user, nil
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
