package orchestrator

import (
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/interfaces"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/model"
)

// ServiceOrchestrator
type ServiceOrchestrator struct {
	authClient    interfaces.AuthClientInterface
	productClient interfaces.ProductClientInterface
	// cartClient clients.cartClient
	// orderClient clients.orderClient

	logger logger.Logger
}

// NewServiceOrchestrator creates an instance of the struct ServiceOrchestrator
func NewServiceOrchestrator(authClient interfaces.AuthClientInterface,
	productClient interfaces.ProductClientInterface,
	log logger.Logger) *ServiceOrchestrator {

	return &ServiceOrchestrator{
		authClient:    authClient,
		productClient: productClient,
		// cartClient: cartClient
		// orderClient: orderClient

		logger: log,
	}
}

/*
* AUTHENTICATION SECTION
 */

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

// GetUser retrieves all the user information related to the specific username
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

// GetUsers retrieves all the information of all users
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

/*
* PRODUCT SECTION
 */

// CreateProduct attemps to create a new product with the given input characteristics
func (so *ServiceOrchestrator) CreateProduct(code, name string, size model.Size, color, description string, stock int32, price float64) (bool, error) {
	succ, err := so.productClient.CreateProduct(code, name, model.ModelSizeToProtoSize(size), color, description, stock, price)
	if err != nil {
		so.logger.Error("Creation error for product with code '%s': %v", code, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Creation failed for product with code '%s', code")
		return false, nil
	}

	so.logger.Info("Successful creation of product with code '%s'", code)
	return true, nil
}

// GetProduct retrieves all the product information realated to the product having the specified code
func (so *ServiceOrchestrator) GetProduct(code string) (bool, *model.Product, error) {
	succ, prod, err := so.productClient.GetProduct(code)
	if err != nil {
		so.logger.Error("Retrieval error of product with code '%s': %v", code, err)
		return false, nil, err
	}
	if !succ {
		so.logger.Warn("Retrieval failed of product with code '%s'", code)
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of product with code '%s'", code)
	return true, model.ProtoProductToModelProduct(prod), nil
}

// UpdateProduct attemps to update name, size, color, description, stock and / or price of the product related to the given code
func (so *ServiceOrchestrator) UpdateProduct(code, name string, size model.Size, color, description string, stock int32, price float64) (bool, error) {
	succ, err := so.productClient.UpdateProduct(code, name, model.ModelSizeToProtoSize(size), color, description, stock, price)
	if err != nil {
		so.logger.Error("Update error for product with code '%s': %v", code, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Update failed for product with code '%s'", code)
		return false, nil
	}

	so.logger.Info("Successful update for product with code '%s'", code)
	return true, nil
}

// DeleteProduct attemps to delete the product with the given code
func (so *ServiceOrchestrator) DeleteProduct(code string) (bool, error) {
	succ, err := so.productClient.DeleteProduct(code)
	if err != nil {
		so.logger.Error("Delete error for product with code '%s': %v", code, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Delete failed for product with code '%s'", code)
		return false, nil
	}

	so.logger.Info("Successful delete for product with code '%s'", code)
	return true, nil
}

// GetProductLists retrieves all the information of all products
func (so *ServiceOrchestrator) GetProductLists() (bool, []*model.Product, error) {
	succ, prods, err := so.productClient.ListProducts()
	if err != nil {
		so.logger.Error("Retrieval error of all products: %v", err)
		return false, nil, err
	}
	if !succ {
		so.logger.Warn("Retrieval of all products failed")
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of all products")
	return true, model.ProtoProductToModelProductsList(prods), nil
}
