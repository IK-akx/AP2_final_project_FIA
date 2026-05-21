package grpc

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"time"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/service"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/pkg/metrics"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/pkg/tracer"
	orderpb "github.com/IK-akx/pharmacy-proto-gen/order"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
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
	prometheusPort string,
) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Общий interceptor с logging, metrics и tracing
	interceptor := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		// Tracing: создаём span для каждого gRPC запроса
		ctx, span := tracer.StartSpan(ctx, info.FullMethod)
		defer span.End()

		span.SetAttributes(
			attribute.String("grpc.method", info.FullMethod),
		)

		// Логирование
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
				span.SetStatus(codes.Error, "handler panicked")
				span.RecordError(fmt.Errorf("panic: %v", r))
				err = status.Errorf(grpccodes.Internal, "internal server error")
			}
		}()

		resp, err = handler(ctx, req)

		// Метрики
		duration := time.Since(start).Seconds()
		statusStr := "success"
		if err != nil {
			statusStr = "error"
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "success")
		}

		metrics.GRPCRequestsTotal.WithLabelValues(info.FullMethod, statusStr).Inc()
		metrics.GRPCRequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

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

	// Prometheus metrics HTTP endpoint
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsAddr := fmt.Sprintf(":%s", prometheusPort)
		logger.Info("Prometheus metrics endpoint", zap.String("addr", metricsAddr))
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			logger.Error("Prometheus metrics server failed", zap.Error(err))
		}
	}()

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
