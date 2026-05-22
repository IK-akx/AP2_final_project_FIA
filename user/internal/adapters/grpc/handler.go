package grpc

import (
	"context"
	"time"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/usecase"
	userpb "github.com/IK-akx/pharmacy-proto-gen/user"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Total number of registered users",
		},
	)
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	registerUC *usecase.RegisterUseCase
	loginUC    *usecase.LoginUseCase
	profileUC  *usecase.ProfileUseCase
}

func NewUserHandler(
	registerUC *usecase.RegisterUseCase,
	loginUC *usecase.LoginUseCase,
	profileUC *usecase.ProfileUseCase,
) *UserHandler {
	return &UserHandler{
		registerUC: registerUC,
		loginUC:    loginUC,
		profileUC:  profileUC,
	}
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *userpb.RegisterRequest) (*userpb.AuthResponse, error) {
	start := time.Now()

	user, err := h.registerUC.Execute(ctx, req.Email, req.Password, req.FirstName, req.LastName)

	duration := time.Since(start).Seconds()
	statusLabel := "success"
	if err != nil {
		statusLabel = "error"
	}

	requestsTotal.WithLabelValues("RegisterUser", statusLabel).Inc()
	requestDuration.WithLabelValues("RegisterUser").Observe(duration)

	if err != nil {
		switch err {
		case usecase.ErrUserExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case usecase.ErrInvalidEmail:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	token, err := h.loginUC.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	activeUsers.Inc()

	return &userpb.AuthResponse{
		Token:     token,
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

func (h *UserHandler) LoginUser(ctx context.Context, req *userpb.LoginRequest) (*userpb.AuthResponse, error) {
	start := time.Now()

	result, err := h.loginUC.Execute(ctx, req.Email, req.Password)

	duration := time.Since(start).Seconds()
	statusLabel := "success"
	if err != nil {
		statusLabel = "error"
	}

	requestsTotal.WithLabelValues("LoginUser", statusLabel).Inc()
	requestDuration.WithLabelValues("LoginUser").Observe(duration)

	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case usecase.ErrWrongPassword:
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &userpb.AuthResponse{
		Token:     result.Token,
		UserId:    result.User.ID,
		Email:     result.User.Email,
		FirstName: result.User.FirstName,
		LastName:  result.User.LastName,
		Role:      result.User.Role,
	}, nil
}

func (h *UserHandler) GetUserProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.UserResponse, error) {
	start := time.Now()

	user, err := h.profileUC.GetProfile(ctx, req.UserId)

	duration := time.Since(start).Seconds()
	statusLabel := "success"
	if err != nil {
		statusLabel = "error"
	}

	requestsTotal.WithLabelValues("GetUserProfile", statusLabel).Inc()
	requestDuration.WithLabelValues("GetUserProfile").Observe(duration)

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &userpb.UserResponse{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		Role:      user.Role,
	}, nil
}

func (h *UserHandler) UpdateUserProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UserResponse, error) {
	start := time.Now()

	user, err := h.profileUC.UpdateProfile(ctx, req.UserId, req.FirstName, req.LastName, req.Phone)

	duration := time.Since(start).Seconds()
	statusLabel := "success"
	if err != nil {
		statusLabel = "error"
	}

	requestsTotal.WithLabelValues("UpdateUserProfile", statusLabel).Inc()
	requestDuration.WithLabelValues("UpdateUserProfile").Observe(duration)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userpb.UserResponse{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		Role:      user.Role,
	}, nil
}

func (h *UserHandler) ValidateToken(ctx context.Context, req *userpb.ValidateTokenRequest) (*userpb.ValidateTokenResponse, error) {
	start := time.Now()

	userID, role, err := h.loginUC.ValidateToken(req.Token)

	duration := time.Since(start).Seconds()
	statusLabel := "success"
	if err != nil {
		statusLabel = "error"
	}

	requestsTotal.WithLabelValues("ValidateToken", statusLabel).Inc()
	requestDuration.WithLabelValues("ValidateToken").Observe(duration)

	if err != nil {
		return &userpb.ValidateTokenResponse{Valid: false}, nil
	}

	return &userpb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
		Role:   role,
	}, nil
}
