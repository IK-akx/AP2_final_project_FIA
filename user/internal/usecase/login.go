package usecase

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/ports"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrWrongPassword = errors.New("wrong password")
)

type LoginUseCase struct {
	repo ports.UserRepository
}

func NewLoginUseCase(repo ports.UserRepository) *LoginUseCase {
	return &LoginUseCase{repo: repo}
}

type LoginResult struct {
	User  *domain.User
	Token string
}

func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrWrongPassword
	}

	token, err := uc.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{User: user, Token: token}, nil
}

func (uc *LoginUseCase) GenerateToken(userID, role string) (string, error) {
	return generateJWT(userID, role)
}

func (uc *LoginUseCase) ValidateToken(tokenStr string) (string, string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret"
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", errors.New("user_id not found in token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return "", "", errors.New("role not found in token")
	}

	return userID, role, nil
}

func generateJWT(userID, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret"
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
