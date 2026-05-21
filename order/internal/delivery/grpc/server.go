package grpc

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/service"
	orderpb "github.com/IK-akx/pharmacy-proto-gen/order"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server wraps the gRPC server
type Server struct {
	server   *grpc.Server
	listener net.Listener
	logger   *zap.Logger
	addr     string
}

// NewServer creates a new gRPC server with interceptors
func NewServer(
	addr string,
	orderSvc service.OrderService,
	balanceSvc service.BalanceService,
	logger *zap.Logger,
) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	interceptor := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		logger.Info("gRPC request",
			zap.String("method", info.FullMethod),
		)

		// Recovery от паник
		defer func() {
			if r := recover(); r != nil {
				logger.Error("gRPC handler panicked",
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		resp, err = handler(ctx, req)
		if err != nil {
			logger.Error("gRPC error",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
		}
		return resp, err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor),
	)

	// Register services
	orderHandler := NewOrderHandler(orderSvc, balanceSvc, logger)
	orderpb.RegisterOrderServiceServer(grpcServer, orderHandler)

	return &Server{
		server:   grpcServer,
		listener: listener,
		logger:   logger,
		addr:     addr,
	}, nil
}

// Start starts the gRPC server in a goroutine
func (s *Server) Start() {
	s.logger.Info("gRPC server starting", zap.String("addr", s.addr))
	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			s.logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.logger.Info("gRPC server stopping")
	s.server.GracefulStop()
}
