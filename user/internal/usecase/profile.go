package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/ports"
)

var ErrProfileNotFound = errors.New("user not found")

type ProfileUseCase struct {
	repo ports.UserRepository
}

func NewProfileUseCase(repo ports.UserRepository) *ProfileUseCase {
	return &ProfileUseCase{repo: repo}
}

func (uc *ProfileUseCase) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := uc.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrProfileNotFound
	}
	return user, nil
}

func (uc *ProfileUseCase) UpdateProfile(ctx context.Context, userID, firstName, lastName, phone string) (*domain.User, error) {
	user, err := uc.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrProfileNotFound
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, *user); err != nil {
		return nil, err
	}
	return user, nil
}
