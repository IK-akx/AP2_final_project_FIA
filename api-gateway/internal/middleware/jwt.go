package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth проверяет JWT токен из заголовка Authorization
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		// Сохраняем user_id и роль в контексте
		userID, _ := claims["user_id"].(string)
		role, _ := claims["role"].(string)

		c.Set("user_id", userID)
		c.Set("role", role)

		c.Next()
	}
}

// OptionalAuth — не блокирует запрос если токена нет (для публичных эндпоинтов)
func OptionalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err == nil && token.Valid {
			claims, _ := token.Claims.(jwt.MapClaims)
			userID, _ := claims["user_id"].(string)
			role, _ := claims["role"].(string)
			c.Set("user_id", userID)
			c.Set("role", role)
		}

		c.Next()
	}
}
