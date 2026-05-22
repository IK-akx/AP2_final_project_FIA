package ports

import (
	"context"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
)

type UserRepository interface {
	Save(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, u domain.User) error
}

type NotificationRepository interface {
	Save(ctx context.Context, log domain.NotificationLog) error
}
