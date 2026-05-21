package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	orderpb "github.com/IK-akx/pharmacy-proto-gen/order"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderHandler struct {
	client orderpb.OrderServiceClient
	logger *zap.Logger
}

func NewOrderHandler(orderServiceAddr string, logger *zap.Logger) (*OrderHandler, error) {
	conn, err := grpc.NewClient(orderServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := orderpb.NewOrderServiceClient(conn)

	return &OrderHandler{
		client: client,
		logger: logger,
	}, nil
}

// CreateOrder — POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req struct {
		Items []struct {
			ProductID string `json:"product_id" binding:"required"`
			Quantity  int32  `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	items := make([]*orderpb.OrderItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = &orderpb.OrderItemRequest{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	resp, err := h.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{
		UserId: userID,
		Items:  items,
	})

	if err != nil {
		h.logger.Error("failed to create order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         resp.Id,
		"user_id":    resp.UserId,
		"status":     resp.Status,
		"total":      resp.Total,
		"items":      resp.Items,
		"created_at": resp.CreatedAt,
	})
}

// GetOrder — GET /orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetOrder(ctx, &orderpb.GetOrderRequest{
		OrderId: orderID,
	})

	if err != nil {
		h.logger.Error("failed to get order", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         resp.Id,
		"user_id":    resp.UserId,
		"status":     resp.Status,
		"total":      resp.Total,
		"items":      resp.Items,
		"created_at": resp.CreatedAt,
		"updated_at": resp.UpdatedAt,
	})
}

// GetUserOrders — GET /orders/user/:userId
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID := c.Param("userId")
	page := parseIntParam(c, "page", 1)
	limit := parseIntParam(c, "limit", 20)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetUserOrders(ctx, &orderpb.GetUserOrdersRequest{
		UserId: userID,
		Page:   int32(page),
		Limit:  int32(limit),
	})

	if err != nil {
		h.logger.Error("failed to get user orders", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":      resp.Orders,
		"total_count": resp.TotalCount,
		"page":        resp.Page,
		"limit":       resp.Limit,
	})
}

// CancelOrder — PUT /orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.CancelOrder(ctx, &orderpb.CancelOrderRequest{
		OrderId: orderID,
	})

	if err != nil {
		h.logger.Error("failed to cancel order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     resp.Id,
		"status": resp.Status,
		"total":  resp.Total,
	})
}

// GetBalance — GET /users/:id/balance
func (h *OrderHandler) GetBalance(c *gin.Context) {
	userID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetUserBalance(ctx, &orderpb.GetBalanceRequest{
		UserId: userID,
	})

	if err != nil {
		h.logger.Error("failed to get balance", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    resp.UserId,
		"balance":    resp.Balance,
		"updated_at": resp.UpdatedAt,
	})
}

func parseIntParam(c *gin.Context, name string, defaultValue int) int {
	val := c.Query(name)
	if val == "" {
		return defaultValue
	}
	var result int
	if _, err := fmt.Sscanf(val, "%d", &result); err != nil {
		return defaultValue
	}
	return result
}
