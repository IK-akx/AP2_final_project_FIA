package grpc

import (
	"context"
	"fmt"

	orderpb "github.com/IK-akx/pharmacy-proto-gen/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient struct {
	client orderpb.OrderServiceClient
}

func NewOrderClient(addr string) (*OrderClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	return &OrderClient{
		client: orderpb.NewOrderServiceClient(conn),
	}, nil
}

// InitBalance вызывает InitBalance в Order Service для создания начального баланса
func (c *OrderClient) InitBalance(ctx context.Context, userID string) error {
	_, err := c.client.InitBalance(ctx, &orderpb.InitBalanceRequest{
		UserId: userID,
	})
	return err
}
