package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/ports"
)

var ErrProfileNotFound = errors.New("user not found")

type ProfileUseCase struct {
	repo  ports.UserRepository
	cache ports.Cache
}

func NewProfileUseCase(repo ports.UserRepository, cache ports.Cache) *ProfileUseCase {
	return &ProfileUseCase{repo: repo, cache: cache}
}

func (uc *ProfileUseCase) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	cacheKey := "profile:" + userID
	cached, err := uc.cache.Get(ctx, cacheKey)
	if err == nil {
		var user domain.User
		if err := json.Unmarshal(cached, &user); err == nil {
			return &user, nil
		}
	}

	user, err := uc.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrProfileNotFound
	}

	data, _ := json.Marshal(user)
	_ = uc.cache.Set(ctx, cacheKey, data, 300)

	return user, nil
}

func (uc *ProfileUseCase) UpdateProfile(ctx context.Context, userID, firstName, lastName, phone string) (*domain.User, error) {
	user, err := uc.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrProfileNotFound
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Phone = phone
	user.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, *user); err != nil {
		return nil, err
	}

	cacheKey := "profile:" + userID
	_ = uc.cache.Delete(ctx, cacheKey)

	return user, nil
}
