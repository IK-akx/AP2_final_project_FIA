package postgres

import (
	"context"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepo struct {
	db *pgxpool.Pool
}

func NewNotificationRepo(db *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Save(ctx context.Context, log domain.NotificationLog) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO notification_logs (id, user_id, type, subject, body, sent_at, status, error_msg)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ID, log.UserID, log.Type, log.Subject, log.Body, log.SentAt, log.Status, log.ErrorMsg,
	)
	return err
}
