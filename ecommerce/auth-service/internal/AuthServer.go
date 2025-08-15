package internal

import (
	"context"
	"fmt"
	"regexp"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal/interfaces"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer implements the authentication service gRPC server.
type AuthServer struct {
	pb.UnimplementedAuthenticationServiceServer
	db     interfaces.AuthServiceInterface
	logger logger.Logger
}

// regular expression to check the validity of an email
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// regular expression to check the validity of a phone number
var phoneRegex = regexp.MustCompile(`^\+?[0-9]+$`)

// NewAuthServer creates a new instance of AuthServer.
func NewAuthServer(db interfaces.AuthServiceInterface, logger logger.Logger) *AuthServer {
	return &AuthServer{
		db:     db,
		logger: logger,
	}
}

// Login authenticates a user with the given username and password.
func (s *AuthServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	if in.Username == "" || in.Password == "" {
		s.logger.Warn("Username or password empty in login request")
		return &pb.LoginResponse{
			Success:      false,
			User:         nil,
			ErrorMessage: "Username and password must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username and password must be provided and not empty")
	}

	succ, user, err := s.db.Login(in.Username, in.Password)

	if err != nil {
		s.logger.Error("Internal error during login for user '%s': %v", in.Username, err)
		return &pb.LoginResponse{
			Success:      false,
			User:         nil,
			ErrorMessage: "Internal server error during the login",
		}, err
	}
	if !succ {
		s.logger.Warn("Invalid login attempt for user '%s'", in.Username)
		return &pb.LoginResponse{
			Success:      false,
			User:         nil,
			ErrorMessage: "Invalid username or password",
		}, nil
	}

	s.logger.Info("Successful login for the user '%s'", in.Username)
	return &pb.LoginResponse{
		Success: true,
		User:    user,
	}, nil
}

// Register creates a new user with the provided registration details.
func (s *AuthServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if in.Username == "" || in.Password == "" || in.Email == "" || in.Phone == "" {
		s.logger.Warn("Username, password, email or phone empty in register request")
		return &pb.RegisterResponse{
			Success:      false,
			ErrorMessage: "Username, password, email and phone must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username, password, email and phone must be provided and not empty")
	}

	if !isValidEmail(in.Email) {
		s.logger.Warn("Invalid email format in register request: %s", in.Email)
		return &pb.RegisterResponse{
			Success:      false,
			ErrorMessage: "Invalid email format",
		}, status.Error(codes.InvalidArgument, "Invalid email format")
	}

	if !isValidPhone(in.Phone) {
		s.logger.Warn("Invalid phone number in register request: %s", in.Phone)
		return &pb.RegisterResponse{
			Success:      false,
			ErrorMessage: "Invalid phone number",
		}, status.Error(codes.InvalidArgument, "Invalid phone number")
	}

	succ, err := s.db.Register(in.Username, in.Password, in.Email, in.Phone)
	if err != nil {
		s.logger.Error("Internal error during registration for user '%s': %v", in.Username, err)
		return &pb.RegisterResponse{
			Success:      false,
			ErrorMessage: "Internal server error during the registration",
		}, err
	}
	if !succ {
		s.logger.Warn("Registration failed for user '%s': username or email already exists", in.Username)
		return &pb.RegisterResponse{
			Success:      false,
			ErrorMessage: "Username or email already exists",
		}, nil
	}

	s.logger.Info("Successful registration for user '%s'", in.Username)
	return &pb.RegisterResponse{
		Success: true,
	}, nil
}

// ChangePassword updates the user's password after verifying the current one.
func (s *AuthServer) ChangePassword(ctx context.Context, in *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	if in.Username == "" || in.OldPassword == "" || in.NewPassword == "" {
		s.logger.Warn("Username, old password or new password empty in change password request")
		return &pb.ChangePasswordResponse{
			Success:      false,
			ErrorMessage: "Username, old password and new password must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username, old password and new password must be provided and not empty")
	}

	succ, err := s.db.ChangePassword(in.Username, in.OldPassword, in.NewPassword)

	if err != nil {
		s.logger.Error("Internal error while changing password for user '%s': %v", in.Username, err)
		return &pb.ChangePasswordResponse{
			Success:      false,
			ErrorMessage: "Internal server error while changing password",
		}, err
	}
	if !succ {
		s.logger.Warn("Password change failed: incorrect current password for user '%s'", in.Username)
		return &pb.ChangePasswordResponse{
			Success:      false,
			ErrorMessage: "Incorrect current password",
		}, nil
	}

	s.logger.Info("Successful password change for user '%s'", in.Username)
	return &pb.ChangePasswordResponse{
		Success: true,
	}, nil
}

// UpdateUser updates a user's email, phone, or role based on the provided fields.
func (s *AuthServer) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if in.Username == "" {
		s.logger.Warn("Username empty in update user request")
		return &pb.UpdateUserResponse{
			Success:      false,
			ErrorMessage: "Username must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username must be provided and not empty")
	}

	if in.Email != "" && !isValidEmail(in.Email) {
		s.logger.Warn("Invalid email format in update user request: %s", in.Email)
		return &pb.UpdateUserResponse{
			Success:      false,
			ErrorMessage: "Invalid email format",
		}, status.Error(codes.InvalidArgument, "Invalid email format")
	}

	if in.Phone != "" && !isValidPhone(in.Phone) {
		s.logger.Warn("Invalid phone number in update user request: %s", in.Phone)
		return &pb.UpdateUserResponse{
			Success:      false,
			ErrorMessage: "Invalid phone number",
		}, status.Error(codes.InvalidArgument, "Invalid phone number")
	}

	succ, err := s.db.UpdateUser(in.Username, in.Email, in.Phone, in.Role)
	if err != nil {
		s.logger.Error("Internal error while updating user '%s': %v", in.Username, err)
		return &pb.UpdateUserResponse{
			Success:      false,
			ErrorMessage: "Internal server error while updating user",
		}, err
	}
	if !succ {
		s.logger.Warn("User update failed for user '%s'", in.Username)
		return &pb.UpdateUserResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("User '%s' not found", in.Username),
		}, nil
	}

	s.logger.Info("Successful update of the user '%s'", in.Username)
	return &pb.UpdateUserResponse{
		Success: true,
	}, nil
}

// GetUser retrieves a single user by username.
func (s *AuthServer) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if in.Username == "" {
		s.logger.Warn("Username empty in get user request")
		return &pb.GetUserResponse{
			Success:      false,
			User:         nil,
			ErrorMessage: "Username must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username must be provided and not empty")
	}

	succ, user, err := s.db.GetUser(in.Username)
	if err != nil {
		s.logger.Error("Internal error while retrieving user '%s': %v", in.Username, err)
		return &pb.GetUserResponse{
			Success:      false,
			User:         nil,
			ErrorMessage: "Internal server error while retrieving user",
		}, err
	}
	if !succ {
		s.logger.Warn("User retrieval failed for user '%s'", in.Username)
		return &pb.GetUserResponse{
			Success:      false,
			User:         nil,
			ErrorMessage: fmt.Sprintf("User '%s' not found", in.Username),
		}, nil
	}

	s.logger.Info("Successful retrieval of the user '%s'", in.Username)
	return &pb.GetUserResponse{
		Success: true,
		User:    user,
	}, nil
}

// GetUsers retrieves all users from the database.
func (s *AuthServer) GetUsers(ctx context.Context, in *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
	succ, users, err := s.db.GetUsers()
	if err != nil {
		s.logger.Error("Internal error while retrieving all users: %v", err)
		return &pb.GetUsersResponse{
			Success:      false,
			Users:        nil,
			ErrorMessage: "Internal server error while retrieving all users",
		}, err
	}
	if !succ {
		s.logger.Warn("No users found")
		return &pb.GetUsersResponse{
			Success:      false,
			Users:        nil,
			ErrorMessage: "No users found",
		}, nil
	}

	s.logger.Info("Successful retrieval of all users")
	return &pb.GetUsersResponse{
		Success: true,
		Users:   users,
	}, nil
}

// isValidEmail check the validity of an email address
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// isValidPhone check the validity of a phone number
func isValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}
