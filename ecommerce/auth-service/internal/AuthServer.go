package internal

import (
	"context"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal/repository"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer implements the authentication service gRPC server.
//
// It handles user authentication, registration, password management,
// and role management by interacting with the database layer and logging events.
type AuthServer struct {
	pb.UnimplementedAuthenticationServer
	db repository.AuthDB
	logger logger.Logger
}


// NewAuthServer creates a new instance of AuthServer.
//
// Parameters:
//   - db: an implementation of repository.AuthDB for database operations.
//   - logger: a logger.Logger instance for logging events.
//
// Returns:
//   - *AuthServer: a pointer to the initialized AuthServer.
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
//   - LoginResponse:
//       - Success: true if authentication succeeded, false otherwise.
//       - Role: the user's role (pb.Role_UNSPECIFIED if authentication fails).
//       - ErrorMessage: description of error or failure reason.
//   - error: non-nil only if an internal error occurred during the login process or there are invalid input parameters.
func (s *AuthServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	if in.Username == "" || in.Password == "" {
		s.logger.Warn("Username or password empty in login request")
		return &pb.LoginResponse{
			Success: false,
			Role: pb.Role_UNSPECIFIED,
			ErrorMessage: "Username and password must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username and password must be provided and not empty")
	}

	success, role, err := s.db.Login(in.Username, in.Password)

	if err != nil {
		s.logger.Error("Internal error during login for user '%s': %v", in.Username, err)
		return &pb.LoginResponse{
			Success: false,
			Role: pb.Role_UNSPECIFIED,
			ErrorMessage: "Internal server error during the login",
		}, err
	}

	if !success {
		s.logger.Warn("Invalid login attempt for user '%s'", in.Username)
		return &pb.LoginResponse{
			Success: false,
			Role: pb.Role_UNSPECIFIED,
			ErrorMessage: "Invalid username or password",
		}, nil
	}

	s.logger.Info("Successful login for the user '%s'", in.Username)
	return &pb.LoginResponse{
		Success: true,
		Role: role,
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
//   - error if an internal failure occurs or there are invalid input parameters.
func (s *AuthServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if in.Username == "" || in.Password == "" || in.Email == "" || in.Phone == ""{
		s.logger.Warn("Username, password, email or phone empty in register request")
		return &pb.RegisterResponse{
			Success: false,
			ErrorMessage: "Username, password, email and phone must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username, password, email and phone must be provided and not empty")
	}

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
//   - error if an internal failure occurs or there are invalid input parameters.
func (s *AuthServer) ChangePassword(ctx context.Context, in *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	if in.Username == "" || in.OldPassword == "" || in.NewPassword == ""{
		s.logger.Warn("Username, old password or new password empty in change password request")
		return &pb.ChangePasswordResponse{
			Success: false,
			ErrorMessage: "Username, old password and new password must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username, old password and new password must be provided and not empty")
	}

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


// SetUserRole sets the role of the specified user.
//
// Parameters:
//   - ctx: the context for the request.
//   - in: the SetUserRoleRequest containing the username and the new role to assign.
//
// Returns:
//   - SetUserRoleResponse indicating success or failure.
//   - error if an internal failure occurs during the operation or there are invalid input parameters.
//
// Notes:
//   - If the user already has the specified role, the response indicates no change (i.e. fails) without an error.
func (s *AuthServer) SetUserRole(ctx context.Context, in *pb.SetUserRoleRequest) (*pb.SetUserRoleResponse, error) {
	if in.Username == "" || in.Role == pb.Role_UNSPECIFIED {
		s.logger.Warn("Username empty or unspecified Role in set user role request")
		return &pb.SetUserRoleResponse{
			Success: false,
			ErrorMessage: "Username must be provided and not empty and the role must not be Unspecified",
		}, status.Error(codes.InvalidArgument, "Username must be provided and not empty and the role must not be Unspecified")
	}

	success, err := s.db.SetUserRole(in.Username, in.Role)

	if err != nil {
		s.logger.Error("Internal error while setting the role for user '%s': %v", in.Username, err)
		return &pb.SetUserRoleResponse{
			Success: false,
			ErrorMessage: "Internal server error while setting the role",
		}, err
	}

	if !success {
		s.logger.Warn("Role setting failed: the user '%s' already has this role", in.Username)
		return &pb.SetUserRoleResponse{
			Success: false,
			ErrorMessage: "Role unchanged: user already has this role.",
		}, nil
	}

	s.logger.Info("Successful role setting for user '%s'", in.Username)
	return &pb.SetUserRoleResponse{
		Success: true,
	}, nil
}


// GetUserRole retrieves the role of the specified user.
//
// Parameters:
//   - ctx: the context for the request.
//   - in: the GetUserRoleRequest containing the username.
//
// Returns:
//   - GetUserRoleResponse with the user's role and success status.
//   - error if an internal failure occurs during the operation.
func (s *AuthServer) GetUserRole(ctx context.Context, in *pb.GetUserRoleRequest) (*pb.GetUserRoleResponse, error) {
	if in.Username == "" {
		s.logger.Warn("Username empty in get user role request")
		return &pb.GetUserRoleResponse{
			Success: false,
			Role: pb.Role_UNSPECIFIED,
			ErrorMessage: "Username must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username must be provided and not empty")
	}

	success, role, err := s.db.GetUserRole(in.Username)

	if err != nil {
		s.logger.Error("Internal error while retrieving the role for user '%s': %v", in.Username, err)
		return &pb.GetUserRoleResponse{
			Success: false,
			Role: pb.Role_UNSPECIFIED,
			ErrorMessage: "Internal server error while retrieving the role",
		}, err
	}

	if !success {
		s.logger.Warn("Role retrieval failed for the user '%s'", in.Username)
		return &pb.GetUserRoleResponse{
			Success: false,
			Role: pb.Role_UNSPECIFIED,
			ErrorMessage: "Unable to retrieve the user's role",
		}, nil
	}

	s.logger.Info("Successful role retrieval for user '%s'", in.Username)
	return &pb.GetUserRoleResponse{
		Success: true,
		Role: role,
	}, nil
}