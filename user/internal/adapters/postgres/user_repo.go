package postgres

import (
	"context"
	"errors"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Save(ctx context.Context, u domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, phone, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		u.ID, u.Email, u.PasswordHash,
		u.FirstName, u.LastName, u.Phone, u.Role,
		u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, first_name, last_name, phone, role, created_at, updated_at
		FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash,
			&u.FirstName, &u.LastName, &u.Phone, &u.Role,
			&u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, first_name, last_name, phone, role, created_at, updated_at
		FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash,
			&u.FirstName, &u.LastName, &u.Phone, &u.Role,
			&u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, u domain.User) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET first_name = $1, last_name = $2, phone = $3, role = $4, updated_at = $5
		WHERE id = $6`,
		u.FirstName, u.LastName, u.Phone, u.Role, u.UpdatedAt, u.ID,
	)
	return err
}
