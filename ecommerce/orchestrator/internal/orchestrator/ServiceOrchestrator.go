package orchestrator

import (
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/grpc/clients"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/model"
)

// ServiceOrchestrator
type ServiceOrchestrator struct {
	authClient clients.AuthClient
	// productClient clients.productClient
	// cartClient clients.cartClient
	// orderClient clients.orderClient

	logger logger.Logger
}

// NewServiceOrchestrator creates an instance of the struct ServiceOrchestrator
func NewServiceOrchestrator(authClient clients.AuthClient, log logger.Logger) *ServiceOrchestrator {
	return &ServiceOrchestrator{
		authClient: authClient,
		// productClient: productClient
		// cartClient: cartClient
		// orderClient: orderClient

		logger: log,
	}
}

// Login attempts to authenticate a user with the provided username and password.
func (so *ServiceOrchestrator) Login(username, password string) (bool, model.Role, error) {
	succ, user, err := so.authClient.Login(username, password)
	if err != nil {
		so.logger.Error("Login error for user '%s': %v", username, err)
		return false, model.RoleUnspecified, err
	}
	if !succ {
		so.logger.Warn("Login failed for user '%s'", username)
		return false, model.RoleUnspecified, nil
	}

	role := model.ProtoRoleToModelRole(user.GetRole())
	so.logger.Info("Successful login for user '%s' with role '%s'", username, role)
	return true, role, nil
}

// Register attempts to create a new user account with given credentials and contact info.
func (so *ServiceOrchestrator) Register(username, password, email, phone string) (bool, error) {
	succ, err := so.authClient.Register(username, password, email, phone)
	if err != nil {
		so.logger.Error("Registration error for user '%s': %v", username, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Registration failed for user '%s'", username)
		return false, nil
	}

	so.logger.Info("Successful registration for user '%s'", username)
	return true, nil
}

// ChangePassword attempts to update a user's password.
func (so *ServiceOrchestrator) ChangePassword(username, oldPassword, newPassword string) (bool, error) {
	succ, err := so.authClient.ChangePassword(username, oldPassword, newPassword)
	if err != nil {
		so.logger.Error("Password change error for user '%s': %v", username, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Password change failed for user '%s'", username)
		return false, nil
	}

	so.logger.Info("Successful password change for user '%s'", username)
	return true, nil
}

// UpdateUser attemps to update email, phone and / or role of the user associated to the username
func (so *ServiceOrchestrator) UpdateUser(username, email, phone string, role model.Role) (bool, error) {
	succ, err := so.authClient.UpdateUser(username, email, phone, model.ModelRoleToProtoRole(role))
	if err != nil {
		so.logger.Error("Update error for user '%s': %v", username, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Update failed for user '%s'", username)
		return false, nil
	}

	so.logger.Info("Successful update for user '%s'", username)
	return true, nil
}

// SetUserRole attempts to set the role of a specific user.
func (so *ServiceOrchestrator) SetUserRole(username string, role model.Role) (bool, error) {
	succ, err := so.UpdateUser(username, "", "", role)
	if err != nil {
		so.logger.Error("Role setting error for user '%s': %v", username, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Role setting failed for user '%s'", username)
		return false, nil
	}

	so.logger.Info("Successful role setting for user '%s'", username)
	return true, nil
}

// GetUserRole retrieves all the user information related to the specific username
func (so *ServiceOrchestrator) GetUser(username string) (bool, *model.User, error) {
	succ, user, err := so.authClient.GetUser(username)
	if err != nil {
		so.logger.Error("Retrieval error of user '%s': %v", username, err)
		return false, nil, err
	}
	if !succ {
		so.logger.Warn("Retrieval failed of user '%s'", username)
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of user '%s'", username)
	return true, model.ProtoUserToModelUser(user), nil
}

// GetUserRole retrieves all the information of all users
func (so *ServiceOrchestrator) GetUsers() (bool, []*model.User, error) {
	succ, users, err := so.authClient.GetUsers()
	if err != nil {
		so.logger.Error("Retrieval error of all users: %v", err)
		return false, nil, err
	}
	if !succ {
		so.logger.Warn("Retrieval of all users failed")
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of all users")
	return true, model.ProtoUsersListToModelUsersList(users), nil
}
