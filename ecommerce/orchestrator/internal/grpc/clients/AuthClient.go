package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
	"google.golang.org/grpc"
)


// AuthClient is a gRPC client for communicating with the authentication service.
type AuthClient struct {
	client pb.AuthenticationClient
	logger logger.Logger
}


// NewAuthClient creates a new instance of AuthClient.
//
// Parameters:
//   - conn: an established gRPC ClientConn to the authentication service.
//
// Returns:
//   - *AuthClient: pointer to the initialized AuthClient.
func NewAuthClient(conn *grpc.ClientConn, log logger.Logger) *AuthClient {
    return &AuthClient{
        client: pb.NewAuthenticationClient(conn),
		logger: log,
    }
}


// Login authenticates a user with the given username and password.
//
// Parameters:
//   - username: user's login name (non-empty).
//   - password: user's password (non-empty).
//
// Returns:
//   - bool: true if login was successful, false otherwise.
//   - pb.Role: role assigned to the user (pb.Role_UNSPECIFIED if login fails).
//   - error: non-nil if an internal error or invalid input occurs.
func (c *AuthClient) Login(username, password string) (bool, pb.Role, error) {
	if username == "" || password == "" {
		c.logger.Error("Username or Password empty in login request")
		return false, pb.Role_UNSPECIFIED, fmt.Errorf("Username and Password must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

	req := &pb.LoginRequest{
		Username: username,
		Password: password,
	}

	res, err := c.client.Login(ctx, req)
	if err != nil {
		c.logger.Error("Login failed for user '%s': %v", username, err)
		return false, pb.Role_UNSPECIFIED, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during login for user '%s'", username)
    	return false, pb.Role_UNSPECIFIED, fmt.Errorf("Login failed: received nil response without error")
	}
	if !res.Success {
		err_mes := res.GetErrorMessage()
		c.logger.Warn("Invalid login attempt for user '%s': %s", username, err_mes)
    	return false, pb.Role_UNSPECIFIED, fmt.Errorf(err_mes)
	}

	c.logger.Info("Login succeeded for user '%s'", username)
	return true, res.GetRole(), nil
}


// Register creates a new user account with the provided credentials and contact info.
//
// Parameters:
//   - username: desired username (non-empty).
//   - password: password (non-empty).
//   - email: user's email address (non-empty).
//   - phone: user's phone number (non-empty).
//
// Returns:
//   - bool: true if registration succeeded, false otherwise.
//   - error: non-nil if an internal error or invalid input occurs.
func (c *AuthClient) Register(username, password, email, phone string) (bool, error) {
	if username == "" || password == "" || email =="" || phone == ""{
		c.logger.Error("Username, password, email or phone empty in register request")
		return false, fmt.Errorf("Username, Password, Email and Phone must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.RegisterRequest{
		Username: username,
		Password: password,
		Email: email,
		Phone: phone,
	}

	res, err := c.client.Register(ctx, req)
	if err != nil {
		c.logger.Error("Registration failed for user '%s': %v", username, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during registration for user '%s'", username)
    	return false, fmt.Errorf("Registration failed: received nil response without error")
	}
	if !res.Success {
		err_mes := res.GetErrorMessage()
		c.logger.Warn("Registration failed for user '%s': %s", username, err_mes)
		return false, fmt.Errorf(err_mes)
	}

	c.logger.Info("Registration succeeded for user '%s'", username)
	return true, nil
}


// ChangePassword updates the password for a given user.
//
// Parameters:
//   - username: user's login name (non-empty).
//   - old_password: current password (non-empty).
//   - new_password: new desired password (non-empty).
//
// Returns:
//   - bool: true if password change succeeded, false otherwise.
//   - error: non-nil if an internal error or invalid input occurs.
func (c *AuthClient) ChangePassword(username, old_password, new_password string) (bool, error) {
	if username == "" || old_password == "" || new_password =="" {
		c.logger.Error("Username, old password or new password empty in change password request")
		return false, fmt.Errorf("Username, old password and new password must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.ChangePasswordRequest{
		Username: username,
		OldPassword: old_password,
		NewPassword: new_password,
	}

	res, err := c.client.ChangePassword(ctx, req)
	if err != nil {
		c.logger.Error("Password change failed for user '%s': %v", username, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during password change for user '%s'", username)
    	return false, fmt.Errorf("Password change failed: received nil response without error")
	}
	if !res.Success {
		err_mes := res.GetErrorMessage()
		c.logger.Warn("Password change failed for user '%s': %s", username, err_mes)
		return false, fmt.Errorf(err_mes)
	}

	c.logger.Info("Password change succeeded for user '%s'", username)
	return true, nil
}


// SetUserRole sets the role of a specific user.
//
// Parameters:
//   - username: target user's login name (non-empty).
//   - role: desired role to assign (must not be pb.Role_UNSPECIFIED).
//
// Returns:
//   - bool: true if role was successfully set, false otherwise.
//   - error: non-nil if an internal error or invalid input occurs.
func (c *AuthClient) SetUserRole(username string, role pb.Role) (bool, error) {
	if username == "" || role == pb.Role_UNSPECIFIED {
		c.logger.Warn("Username empty or unspecified Role in set user role request")
		return false, fmt.Errorf("Username must be provided and not empty and the role must not be Unspecified")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.SetUserRoleRequest{
		Username: username,
		Role: role,
	}

	res, err := c.client.SetUserRole(ctx, req)
	if err != nil {
		c.logger.Error("Role setting failed for user '%s': %v", username, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during role setting for user '%s'", username)
    	return false, fmt.Errorf("Role setting failed: received nil response without error")
	}
	if !res.Success {
		err_mes := res.GetErrorMessage()
		c.logger.Warn("Role setting failed for user '%s': %s", username, err_mes)
		return false, fmt.Errorf(err_mes)
	}

	c.logger.Info("Role setting succeeded for user '%s'", username)
	return true, nil
}


// GetUserRole retrieves the role assigned to a given user.
//
// Parameters:
//   - username: target user's login name (non-empty).
//
// Returns:
//   - bool: true if retrieval succeeded, false otherwise.
//   - pb.Role: role assigned to the user (pb.Role_UNSPECIFIED if retrieval fails).
//   - error: non-nil if an internal error or invalid input occurs.
func (c *AuthClient) GetUserRole(username string) (bool, pb.Role, error) {
	if username == "" {
		c.logger.Warn("Username empty in get user role request")
		return false, pb.Role_UNSPECIFIED, fmt.Errorf("Username must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.GetUserRoleRequest{
		Username: username,
	}

	res, err := c.client.GetUserRole(ctx, req)
	if err != nil {
		c.logger.Error("Role retrieval failed for user '%s': %v", username, err)
		return false, pb.Role_UNSPECIFIED, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during role retrieval for user '%s'", username)
    	return false, pb.Role_UNSPECIFIED, fmt.Errorf("Role retrieval failed: received nil response without error")
	}
	if !res.Success {
		err_mes := res.GetErrorMessage()
		c.logger.Warn("Role retrieval failed for user '%s': %s", username, err_mes)
		return false, pb.Role_UNSPECIFIED, fmt.Errorf(err_mes)
	}

	c.logger.Info("Role retrieval succeeded for user '%s'", username)
	return true, res.GetRole(), nil
}