package grpc

import (
	"context"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/usecase"
	userpb "github.com/IK-akx/pharmacy-proto-gen/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	user, err := h.registerUC.Execute(ctx, req.Email, req.Password, req.FirstName, req.LastName)
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
	result, err := h.loginUC.Execute(ctx, req.Email, req.Password)
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
	user, err := h.profileUC.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &userpb.UserResponse{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

func (h *UserHandler) UpdateUserProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UserResponse, error) {
	user, err := h.profileUC.UpdateProfile(ctx, req.UserId, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userpb.UserResponse{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

func (h *UserHandler) ValidateToken(ctx context.Context, req *userpb.ValidateTokenRequest) (*userpb.ValidateTokenResponse, error) {
	userID, role, err := h.loginUC.ValidateToken(req.Token)
	if err != nil {
		return &userpb.ValidateTokenResponse{Valid: false}, nil
	}

	return &userpb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
		Role:   role,
	}, nil
}
