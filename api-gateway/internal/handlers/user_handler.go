package handlers

import (
	"context"
	"net/http"
	"time"

	userpb "github.com/IK-akx/pharmacy-proto-gen/user"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserHandler struct {
	client userpb.UserServiceClient
	logger *zap.Logger
}

func NewUserHandler(userServiceAddr string, logger *zap.Logger) (*UserHandler, error) {
	conn, err := grpc.NewClient(userServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := userpb.NewUserServiceClient(conn)

	return &UserHandler{
		client: client,
		logger: logger,
	}, nil
}

// Register — POST /auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req userpb.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.RegisterUser(ctx, &req)
	if err != nil {
		h.logger.Error("register failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login — POST /auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req userpb.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.LoginUser(ctx, &req)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProfile — GET /users/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetUserProfile(ctx, &userpb.GetProfileRequest{
		UserId: userID,
	})
	if err != nil {
		h.logger.Error("get profile failed", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ValidateToken — вызывает User Service для проверки JWT
func (h *UserHandler) ValidateToken(c *gin.Context) {
	token := c.GetString("raw_token") // из middleware
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no token"})
		c.Abort()
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	resp, err := h.client.ValidateToken(ctx, &userpb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil || !resp.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	c.Set("user_id", resp.UserId)
	c.Set("role", resp.Role)
	c.Next()
}
