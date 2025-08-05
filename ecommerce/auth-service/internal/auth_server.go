package internal

import (
	"context"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal/repository"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
)

type AuthServer struct {
	pb.UnimplementedAuthenticationServer
	db repository.AuthDB
	logger logger.Logger
}

func NewAuthServer(db repository.AuthDB, logger logger.Logger) *AuthServer {
	return &AuthServer{
		db: db,
		logger: logger,
	}
}

// Login authenticates a user with the given username and password.
//
// Parameters:
//   - ctx: request context.
//   - in: LoginRequest containing username and password.
//
// Returns:
//   - LoginResponse indicating success or failure.
//   - error if an internal failure occurs.
func (s *AuthServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	success, err := s.db.Login(in.Username, in.Password)

	if err != nil {
		s.logger.Error("Internal error during login for user '%s': %v", in.Username, err)
		return &pb.LoginResponse{
			Success: false,
			ErrorMessage: "Internal server error during the login",
		}, err
	}

	if !success {
		s.logger.Warn("Invalid login attempt for user '%s'", in.Username)
		return &pb.LoginResponse{
			Success: false,
			ErrorMessage: "Invalid username or password",
		}, nil
	}

	s.logger.Info("Successful login for the user '%s'", in.Username)
	return &pb.LoginResponse{
		Success: true,
	}, nil
}

// Register creates a new user with the provided registration details.
//
// Parameters:
//   - ctx: request context.
//   - in: RegisterRequest containing username, password, email, and phone.
//
// Returns:
//   - RegisterResponse indicating success or failure.
//   - error if an internal failure occurs.
func (s *AuthServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	success, err := s.db.Register(in.Username, in.Password, in.Email, in.Phone)

	if err != nil {
		s.logger.Error("Internal error during registration for user '%s': %v", in.Username, err)
		return &pb.RegisterResponse{
			Success: false,
			ErrorMessage: "Internal server error during the registration",
		}, err
	}

	if !success {
		s.logger.Warn("Registration failed for user '%s': username or email already exists", in.Username)
		return &pb.RegisterResponse{
			Success: false,
			ErrorMessage: "Username or email already exists",
		}, nil
	}

	s.logger.Info("Successful registration for user '%s'", in.Username)
	return &pb.RegisterResponse{
		Success: true,
	}, nil
}

// ChangePassword updates the user's password after verifying the current one.
//
// Parameters:
//   - ctx: request context.
//   - in: ChangePasswordRequest containing username, old password, and new password.
//
// Returns:
//   - ChangePasswordResponse indicating success or failure.
//   - error if an internal failure occurs.
func (s *AuthServer) ChangePassword(ctx context.Context, in *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	success, err := s.db.ChangePassword(in.Username, in.OldPassword, in.NewPassword)

	if err != nil {
		s.logger.Error("Internal error while changing password for user '%s': %v", in.Username, err)
		return &pb.ChangePasswordResponse{
			Success: false,
			ErrorMessage: "Internal server error while changing password",
		}, err
	}

	if !success {
		s.logger.Warn("Password change failed: incorrect current password for user '%s'", in.Username)
		return &pb.ChangePasswordResponse{
			Success: false,
			ErrorMessage: "Incorrect current password",
		}, nil
	}

	s.logger.Info("Successful password change for user '%s'", in.Username)
	return &pb.ChangePasswordResponse{
		Success: true,
	}, nil
}