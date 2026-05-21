package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	productpb "github.com/IK-akx/pharmacy-proto-gen/product"
)

// ProductClient implements ProductServiceClient interface via gRPC
type ProductClient struct {
	client productpb.ProductServiceClient
	logger *zap.Logger
}

func NewProductClient(addr string, logger *zap.Logger) (*ProductClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product service: %w", err)
	}

	client := productpb.NewProductServiceClient(conn)

	return &ProductClient{
		client: client,
		logger: logger,
	}, nil
}

func (c *ProductClient) CheckAvailability(ctx context.Context, productID uuid.UUID, quantity int32) (bool, int32, error) {
	resp, err := c.client.CheckAvailability(ctx, &productpb.CheckAvailabilityRequest{
		ProductId: productID.String(),
		Quantity:  quantity,
	})
	if err != nil {
		c.logger.Error("failed to check product availability",
			zap.String("product_id", productID.String()),
			zap.Int32("quantity", quantity),
			zap.Error(err),
		)
		return false, 0, fmt.Errorf("failed to check availability: %w", err)
	}

	return resp.Available, resp.CurrentStock, nil
}

func (c *ProductClient) UpdateStock(ctx context.Context, productID uuid.UUID, delta int32) error {
	resp, err := c.client.UpdateStock(ctx, &productpb.UpdateStockRequest{
		ProductId:     productID.String(),
		QuantityDelta: delta,
	})
	if err != nil {
		c.logger.Error("failed to update product stock",
			zap.String("product_id", productID.String()),
			zap.Int32("delta", delta),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update stock: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("stock update failed for product %s", productID.String())
	}

	return nil
}
