package handlers

import (
	"context"
	"net/http"
	"time"

	productpb "github.com/IK-akx/pharmacy-proto-gen/product"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductHandler struct {
	client productpb.ProductServiceClient
	logger *zap.Logger
}

func NewProductHandler(productServiceAddr string, logger *zap.Logger) (*ProductHandler, error) {
	conn, err := grpc.NewClient(productServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := productpb.NewProductServiceClient(conn)

	return &ProductHandler{
		client: client,
		logger: logger,
	}, nil
}

// ListProducts — GET /products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page := parseIntParam(c, "page", 1)
	limit := parseIntParam(c, "limit", 20)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ListProducts(ctx, &productpb.ListProductsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	})
	if err != nil {
		h.logger.Error("list products failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProduct — GET /products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetProduct(ctx, &productpb.GetProductRequest{Id: id})
	if err != nil {
		h.logger.Error("get product failed", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
