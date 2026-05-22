package middleware

import (
	"net/http"
	"strings"

	userpb "github.com/IK-akx/pharmacy-proto-gen/user"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func JWTAuth(userServiceAddr string) gin.HandlerFunc {
	conn, _ := grpc.NewClient(userServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	client := userpb.NewUserServiceClient(conn)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid format"})
			c.Abort()
			return
		}

		resp, _ := client.ValidateToken(c.Request.Context(), &userpb.ValidateTokenRequest{
			Token: parts[1],
		})

		if resp == nil || !resp.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", resp.UserId)
		c.Set("role", resp.Role)
		c.Next()
	}
}
