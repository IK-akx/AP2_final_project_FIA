package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	// TODO: add gRPC client when User Service will be done
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// Register — POST /auth/register
func (h *UserHandler) Register(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "user service not yet integrated — registration mocked",
		"token":   "mock-jwt-token",
	})
}

// Login — POST /auth/login
func (h *UserHandler) Login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "user service not yet integrated — login mocked",
		"token":   "mock-jwt-token",
	})
}

// GetProfile — GET /users/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"message": "user service not yet integrated",
	})
}
