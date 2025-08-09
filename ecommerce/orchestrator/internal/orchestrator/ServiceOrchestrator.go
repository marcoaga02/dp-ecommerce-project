package orchestrator

import (
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/grpc/clients"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/model"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
)

// ServiceOrchestrator
type ServiceOrchestrator struct {
	authClient clients.AuthClient
    // productClient clients.productClient
    // cartClient clients.cartClient
    // orderClient clients.orderClient

	logger logger.Logger
}

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
//
// Parameters:
//   - username: the user's login identifier
//   - password: the user's password
//
// Returns:
//   - bool: true if login was successful, false otherwise
//   - model.Role: the role of the user if login succeeds; RoleUnspecified otherwise
//   - error: non-nil if a server or internal error occurred during login
func (so *ServiceOrchestrator) Login(username, password string) (bool, model.Role, error) {
	succ, protoRole, err := so.authClient.Login(username, password)
	if err != nil {
		so.logger.Error("Login error for user '%s': %v", username, err)
		return false, model.RoleUnspecified, err
	}
	if !succ {
		so.logger.Warn("Login failed for user '%s'", username)
		return false, model.RoleUnspecified, nil
	}

	role := model.ProtoRoleToRole(protoRole)
	so.logger.Info("Successful login for user '%s' with role '%s'", username, role)
	return true, role, nil
}


// Register attempts to create a new user account with given credentials and contact info.
//
// Parameters:
//   - username: desired username
//   - password: desired password
//   - email: user's email address
//   - phone: user's phone number
//
// Returns:
//   - bool: true if registration was successful, false otherwise
//   - error: non-nil if a server or internal error occurred during registration
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
//
// Parameters:
//   - username: user's login identifier
//   - oldPassword: current password
//   - newPassword: new desired password
//
// Returns:
//   - bool: true if password change succeeded, false otherwise
//   - error: non-nil if a server or internal error occurred during password update
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


// SetUserRole attempts to set the role of a specific user.
//
// Parameters:
//   - username: user's login identifier
//   - role: desired role to assign
//
// Returns:
//   - bool: true if role setting succeeded, false otherwise
//   - error: non-nil if a server or internal error occurred during role assignment
func (so *ServiceOrchestrator) SetUserRole(username string, role model.Role) (bool, error) {
	pbRole := model.RoleToProtoRole(role)
	succ, err := so.authClient.SetUserRole(username, pbRole)
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


// GetUserRole retrieves the role assigned to a specific user.
//
// Parameters:
//   - username: user's login identifier
//
// Returns:
//   - bool: true if retrieval succeeded, false otherwise
//   - model.Role: the role assigned to the user if retrieval succeeded; RoleUnspecified otherwise
//   - error: non-nil if a server or internal error occurred during role retrieval
func (so *ServiceOrchestrator) GetUserRole(username string) (bool, model.Role, error) {
	succ, pbRole, err := so.authClient.GetUserRole(username)
	if err != nil {
		so.logger.Error("Role retrieval error for user '%s': %v", username, err)
		return false, model.RoleUnspecified, err
	}
	if !succ {
		so.logger.Warn("Role retrieval failed for user '%s'", username)
		return false, model.RoleUnspecified, nil
	}

	role := model.ProtoRoleToRole(pbRole)
	so.logger.Info("Successful role retrieval for user '%s'", username)
	return true, role, nil
}

