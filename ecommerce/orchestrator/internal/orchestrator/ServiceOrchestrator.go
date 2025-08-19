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
	cartClient    interfaces.CartClientInterface
	orderClient   interfaces.OrderClientInterface

	logger logger.Logger
}

// NewServiceOrchestrator creates an instance of the struct ServiceOrchestrator
func NewServiceOrchestrator(
	authClient interfaces.AuthClientInterface,
	productClient interfaces.ProductClientInterface,
	cartClient interfaces.CartClientInterface,
	orderClient interfaces.OrderClientInterface,
	log logger.Logger,
) *ServiceOrchestrator {

	return &ServiceOrchestrator{
		authClient:    authClient,
		productClient: productClient,
		cartClient:    cartClient,
		orderClient:   orderClient,
		logger:        log,
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

// UpdateUser attempts to update email, phone and / or role of the user associated to the username
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
	if !succ || user == nil {
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
	if !succ || users == nil || len(users) == 0 {
		so.logger.Warn("No users returned")
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of all users")
	return true, model.ProtoUsersListToModelUsersList(users), nil
}

/*
* PRODUCT SECTION
 */

// CreateProduct attempts to create a new product with the given input characteristics
func (so *ServiceOrchestrator) CreateProduct(code, name string, size model.Size, color, description string, stock uint32, price float64) (bool, error) {
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
	if !succ || prod == nil {
		so.logger.Warn("Retrieval failed of product with code '%s'", code)
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of product with code '%s'", code)
	return true, model.ProtoProductToModelProduct(prod), nil
}

// UpdateProduct attempts to update name, size, color, description, stock and / or price of the product related to the given code
func (so *ServiceOrchestrator) UpdateProduct(code, name string, size model.Size, color, description string, stock uint32, price float64) (bool, error) {
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

// DeleteProduct attempts to delete the product with the given code
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
	if !succ || prods == nil || len(prods) == 0 {
		so.logger.Warn("No products returned")
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of all products")
	return true, model.ProtoProductsListToModelProductsList(prods), nil
}

/*
* CART SECTION
 */

// AddProductToCart attempts to add a certain quantity of a product to a user's cart
func (so *ServiceOrchestrator) AddProductToCart(username, prodCode string, quantity uint32) (bool, error) {
	succ, err := so.cartClient.AddItem(username, prodCode, quantity)
	if err != nil {
		so.logger.Error("Error while adding product '%s' to the cart of user '%s': %v", prodCode, username, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Failed addition of product '%s' to the cart of user '%s'", prodCode, username)
		return false, nil
	}

	so.logger.Info("Successful addition of n=%d products '%s' to the cart of user '%s'", quantity, prodCode, username)
	return true, nil
}

// RemoveProductFromCart attempts to fully remove a product from a user's cart
func (so *ServiceOrchestrator) RemoveProductFromCart(username, prodCode string) (bool, error) {
	succ, err := so.cartClient.RemoveItem(username, prodCode)
	if err != nil {
		so.logger.Error("Error while removing product '%s' form the cart of user '%s': %v", prodCode, username, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Failed removal of product '%s' from the cart of user '%s'", prodCode, username)
		return false, nil
	}

	so.logger.Info("Successful removal of product '%s' from the cart of user '%s'", prodCode, username)
	return true, nil
}

// UpdateQuantityOfProductIntoCart attempts to update the quantity of a product into the user's cart
func (so *ServiceOrchestrator) UpdateQuantityOfProductIntoCart(username, prodCode string, quantity uint32) (bool, error) {
	succ, err := so.cartClient.UpdateItemQuantity(username, prodCode, quantity)
	if err != nil {
		so.logger.Error("Error while updating quantity of product '%s' into the cart of user '%s': %v", prodCode, username, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Failed update of quantity for product '%s' into the cart of user '%s'", prodCode, username)
		return false, nil
	}

	so.logger.Info("Successful update at n=%d of the quantity for product '%s' into the cart of user '%s'", quantity, prodCode, username)
	return true, nil
}

// GetListProductsIntoCart attempts to retrieves the list of all products, with their quantities, from the cart of a user
func (so *ServiceOrchestrator) GetListOfProductsIntoCart(username string) (bool, []*model.CartItem, error) {
	succ, items, err := so.cartClient.ListCartItems(username)
	if err != nil {
		so.logger.Error("Error while retrieving products into the cart of user '%s': %v", username, err)
		return false, nil, nil
	}
	if !succ || items == nil || len(items) == 0 {
		so.logger.Warn("No cart items returned for user '%s'", username)
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of products from cart of user '%s'", username)
	return true, model.ProtoCartItemsListToModelCartItemsList(items), nil
}

// RemoveAllProductsFromCart attempts to remova all the products from the cart of a user
func (so *ServiceOrchestrator) RemoveAllProductsFromCart(username string) (bool, error) {
	succ, err := so.cartClient.ClearCart(username)
	if err != nil {
		so.logger.Error("Error while removing all product from the cart of user '%s': %v", username, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Failed removal of all products from cart of user '%s'", username)
		return false, nil
	}

	so.logger.Info("Successful removal of all products from cart of user '%s'", username)
	return true, nil
}

// RemoveProductFromAllCarts removes all cart items related to a given product
func (so *ServiceOrchestrator) RemoveProductFromAllCarts(code string) (bool, error) {
	succ, err := so.cartClient.RemoveProductFromAllCarts(code)
	if err != nil {
		so.logger.Error("Error while removing all cart items for product with code '%s': %v", code, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Failed removal of all cart items for product with code '%s'", code)
		return false, nil
	}

	so.logger.Info("Successful removal of all cart items for product with code '%s'", code)
	return true, nil
}

// GetItemFromCart retrieves a cart item from the cart of a user
func (so *ServiceOrchestrator) GetItemFromCart(username, code string) (bool, *model.CartItem, error) {
	succ, items, err := so.GetListOfProductsIntoCart(username)
	if err != nil { // error already logged in the other orchestrator function
		return false, nil, err
	}
	if !succ { // warning already logged in the other orchestrator function
		return false, nil, nil
	}

	for _, item := range items {
		if item.ProdCode == code {
			so.logger.Info("Product with code '%s' found into the cart of user '%s'", code, username)
			return true, item, nil
		}
	}

	so.logger.Error("Product with code '%s' not found into the cart of user '%s'", code, username)
	return false, nil, nil
}

/*
* ORDER SECTION
 */

// CreateOrder creates a new order
func (so *ServiceOrchestrator) CreateOrder(username string, items []*model.OrderItem) (bool, int32, error) {
	const errOrderID int32 = -1
	succ, orderId, err := so.orderClient.CreateOrder(username, model.ModelOrderItemsListToProtoOrderItemsList(items))
	if err != nil {
		so.logger.Error("Error during creation of order for user '%s' :%v", username, err)
		return false, errOrderID, err
	}
	if !succ {
		so.logger.Warn("Order creation failed for user '%s'", username)
		return false, errOrderID, nil
	}

	so.logger.Info("Successful order creation for user '%s'", username)
	return true, orderId, nil
}

// GetOrder retrieves the order with the given id
func (so *ServiceOrchestrator) GetOrder(orderId int32) (bool, *model.Order, error) {
	succ, order, err := so.orderClient.GetOrder(orderId)
	if err != nil {
		so.logger.Error("Retrieval error for order with ID '%d': %v", order, err)
		return false, nil, err
	}
	if !succ || order == nil {
		so.logger.Warn("Retrieval failed for order with ID '%d'", orderId)
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of order with ID '%d'", orderId)
	return true, model.ProtoOrderToModelOrder(order), nil
}

// UpdateOrderStatus updates the status of the order with the given id
func (so *ServiceOrchestrator) UpdateOrderStatus(orderId int32, status model.Status) (bool, error) {
	succ, err := so.orderClient.UpdateOrderStatus(orderId, model.ModelStatusToProtoStatus(status))
	if err != nil {
		so.logger.Error("Update status error for order with ID '%d': %v", orderId, err)
		return false, err
	}
	if !succ {
		so.logger.Warn("Status update failed for order with ID '%d'", orderId)
		return false, err
	}

	so.logger.Info("Successful update of the status to '%s' for order with ID '%d'", status.String(), orderId)
	return true, nil
}

// GetOrdersListByUsername retrieves the list of all orders of a given user
func (so *ServiceOrchestrator) GetOrdersListByUsername(username string) (bool, []*model.Order, error) {
	succ, orders, err := so.orderClient.ListOrdersByUsername(username)
	if err != nil {
		so.logger.Error("Retrieval error for orders of user '%s': %v", username, err)
		return false, nil, err
	}
	if !succ || orders == nil || len(orders) == 0 {
		so.logger.Warn("Failed retrieval of orders for user '%s'", username)
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of products for user '%s'", username)
	return true, model.ProtoOrdersListToModelOrdersList(orders), nil
}

// GetAllOrdersList retrieves the list of all orders
func (so *ServiceOrchestrator) GetAllOrdersList() (bool, []*model.Order, error) {
	succ, orders, err := so.orderClient.ListAllOrders()
	if err != nil {
		so.logger.Error("Error during the retrieval of all orders")
		return false, nil, err
	}
	if !succ || orders == nil || len(orders) == 0 {
		so.logger.Warn("Failed retrieval of all orders")
		return false, nil, nil
	}

	so.logger.Info("Successful retrieval of all orders")
	return true, model.ProtoOrdersListToModelOrdersList(orders), nil
}
