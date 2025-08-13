package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/manager"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
)

// AuthClient is a gRPC client for communicating with the authentication service.
//
// Implements the AuthClientInterface
type AuthClient struct {
	serviceName string
	sm          *manager.ServiceManager
	logger      logger.Logger
	timeout     time.Duration
}

// NewAuthClient creates a new instance of AuthClient.
func NewAuthClient(serviceName string, sm *manager.ServiceManager, log logger.Logger, timeout time.Duration) *AuthClient {
	return &AuthClient{
		serviceName: serviceName,
		sm:          sm,
		logger:      log,
		timeout:     timeout,
	}
}

// Login authenticates a user with the given username and password.
func (c *AuthClient) Login(username, password string) (bool, *pb.User, error) {
	if username == "" || password == "" {
		c.logger.Error("Username or Password empty in login request")
		return false, nil, fmt.Errorf("Username and Password must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, nil, err
	}

	req := &pb.LoginRequest{
		Username: username,
		Password: password,
	}

	res, err := client.Login(ctx, req)
	if err != nil {
		c.logger.Error("Login error for user '%s': %v", username, err)
		return false, nil, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during login for user '%s'", username)
		return false, nil, fmt.Errorf("Login failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Invalid login attempt for user '%s': %s", username, res.GetErrorMessage())
		return false, nil, nil
	}

	c.logger.Info("Login succeeded for user '%s'", username)
	return true, res.GetUser(), nil
}

// Register creates a new user account with the provided credentials and contact info.
func (c *AuthClient) Register(username, password, email, phone string) (bool, error) {
	if username == "" || password == "" || email == "" || phone == "" {
		c.logger.Error("Username, password, email or phone empty in register request")
		return false, fmt.Errorf("Username, Password, Email and Phone must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.RegisterRequest{
		Username: username,
		Password: password,
		Email:    email,
		Phone:    phone,
	}

	res, err := client.Register(ctx, req)
	if err != nil {
		c.logger.Error("Registration error for user '%s': %v", username, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during registration for user '%s'", username)
		return false, fmt.Errorf("Registration failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Registration failed for user '%s': %s", username, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Registration succeeded for user '%s'", username)
	return true, nil
}

// ChangePassword updates the password for a given user.
func (c *AuthClient) ChangePassword(username, oldPassword, newPassword string) (bool, error) {
	if username == "" || oldPassword == "" || newPassword == "" {
		c.logger.Error("Username, old password or new password empty in change password request")
		return false, fmt.Errorf("Username, old password and new password must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.ChangePasswordRequest{
		Username:    username,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	res, err := client.ChangePassword(ctx, req)
	if err != nil {
		c.logger.Error("Password change error for user '%s': %v", username, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during password change for user '%s'", username)
		return false, fmt.Errorf("Password change failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Password change failed for user '%s': %s", username, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Password change succeeded for user '%s'", username)
	return true, nil
}

// UpdateUser updates user email, phone or role.
func (c *AuthClient) UpdateUser(username, email, phone string, role pb.Role) (bool, error) {
	if username == "" {
		c.logger.Error("Username empty in update user request")
		return false, fmt.Errorf("Username must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.UpdateUserRequest{
		Username: username,
		Email:    email,
		Phone:    phone,
		Role:     role,
	}

	res, err := client.UpdateUser(ctx, req)
	if err != nil {
		c.logger.Error("Update error for user '%s': %v", username, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during update for user '%s'", username)
		return false, fmt.Errorf("User update failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Update failed for user '%s': %s", username, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Update succeeded for user '%s'", username)
	return true, nil
}

// GetUser retrieves the user information for the specified username.
func (c *AuthClient) GetUser(username string) (bool, *pb.User, error) {
	if username == "" {
		c.logger.Error("Username empty in get user request")
		return false, nil, fmt.Errorf("Username must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, nil, err
	}

	req := &pb.GetUserRequest{
		Username: username,
	}

	res, err := client.GetUser(ctx, req)
	if err != nil {
		c.logger.Error("Error during the retrieval of the user '%s': %v", username, err)
		return false, nil, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during the retrieval of the user '%s'", username)
		return false, nil, fmt.Errorf("User retrieval failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed retrieval of user '%s': %s", username, res.GetErrorMessage())
		return false, nil, nil
	}

	c.logger.Info("Successful retrieval of the user '%s'", username)
	return true, res.GetUser(), nil

}

// GetUsers retrieves all users registered in the system.
func (c *AuthClient) GetUsers() (bool, []*pb.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, nil, err
	}

	res, err := client.GetUsers(ctx, &pb.GetUsersRequest{})
	if err != nil {
		c.logger.Error("Error during the retrieval of all users: %v", err)
		return false, nil, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during the retrieval of all users")
		return false, nil, fmt.Errorf("Users retrieval failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed retrieval of all users: %s", res.GetErrorMessage())
		return false, nil, nil
	}

	c.logger.Info("Successful retrieval of all users")
	return true, res.GetUsers(), nil
}

// getClient retrieves a new AuthenticationServiceClient connected to the service.
//
// Returns:
//   - pb.AuthenticationServiceClient: the gRPC client to communicate with the auth service
//   - error: non-nil if connection setup failed
func (c *AuthClient) getClient() (pb.AuthenticationServiceClient, error) {
	conn, err := c.sm.GetConnWithTimeout(c.serviceName, c.timeout)
	if err != nil {
		return nil, err
	}
	return pb.NewAuthenticationServiceClient(conn), nil
}
