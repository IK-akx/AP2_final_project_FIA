package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	// TODO: add gRPC client when Product Service will bw done
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

// ListProducts — GET /products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"products": []gin.H{},
		"message":  "product service not yet integrated",
	})
}

// GetProduct — GET /products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "product service not yet integrated",
	})
}
