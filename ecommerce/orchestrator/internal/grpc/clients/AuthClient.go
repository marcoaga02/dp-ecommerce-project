package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/manager"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
)


// AuthClient is a gRPC client for communicating with the authentication service.
type AuthClient struct {
	serviceName string
	sm     *manager.ServiceManager
	logger logger.Logger
	timeout     time.Duration
}


// NewAuthClient creates a new instance of AuthClient.
//
// Parameters:
//   - serviceName: the service name
//   - sm: the service manager
//   - log: the logger
//   - timeout: maximum duration for RPC calls (e.g., 1 * time.Second)
//
// Returns:
//   - *AuthClient: pointer to the initialized AuthClient.
func NewAuthClient(serviceName string, sm *manager.ServiceManager, log logger.Logger, timeout time.Duration) *AuthClient {
	return &AuthClient{
		serviceName: serviceName,
		sm: sm,
		logger: log,
		timeout: timeout,
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

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

	client, err := c.getClient()
	if err != nil {
		return false, pb.Role_UNSPECIFIED, err
	}

	req := &pb.LoginRequest{
		Username: username,
		Password: password,
	}

	res, err := client.Login(ctx, req)
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

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		return false, err
	}

	req := &pb.RegisterRequest{
		Username: username,
		Password: password,
		Email: email,
		Phone: phone,
	}

	res, err := client.Register(ctx, req)
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
//   - oldPassword: current password (non-empty).
//   - newPassword: new desired password (non-empty).
//
// Returns:
//   - bool: true if password change succeeded, false otherwise.
//   - error: non-nil if an internal error or invalid input occurs.
func (c *AuthClient) ChangePassword(username, oldPassword, newPassword string) (bool, error) {
	if username == "" || oldPassword == "" || newPassword =="" {
		c.logger.Error("Username, old password or new password empty in change password request")
		return false, fmt.Errorf("Username, old password and new password must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		return false, err
	}

	req := &pb.ChangePasswordRequest{
		Username: username,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	res, err := client.ChangePassword(ctx, req)
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

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		return false, err
	}

	req := &pb.SetUserRoleRequest{
		Username: username,
		Role: role,
	}

	res, err := client.SetUserRole(ctx, req)
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

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		return false, pb.Role_UNSPECIFIED, err
	}

	req := &pb.GetUserRoleRequest{
		Username: username,
	}

	res, err := client.GetUserRole(ctx, req)
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

// getClient retrieves a new AuthenticationClient connected to the service.
//
// Returns:
//   - pb.AuthenticationClient: the gRPC client to communicate with the auth service
//   - error: non-nil if connection setup failed
func (c *AuthClient) getClient() (pb.AuthenticationClient, error) {
	conn, err := c.sm.GetConnWithTimeout(c.serviceName, c.timeout)
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return nil, err
	}
	return pb.NewAuthenticationClient(conn), nil
}
