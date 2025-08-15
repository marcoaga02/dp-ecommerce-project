package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/manager"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/cart"
)

// CartClient is a gRPC client for communicating with the cart service
//
// Implements the CartClientInterface
type CartClient struct {
	serviceName string
	sm          *manager.ServiceManager
	logger      logger.Logger
	timeout     time.Duration
}

// NewCartClient returns a new instance of CartClient
func NewCartClient(serviceName string, sm *manager.ServiceManager, log logger.Logger, timeout time.Duration) *CartClient {
	return &CartClient{
		serviceName: serviceName,
		sm:          sm,
		logger:      log,
		timeout:     timeout,
	}
}

// AddItem attempts to increase the quantity of a product into the user's cart
func (c *CartClient) AddItem(username, code string, quantity uint32) (bool, error) {
	if username == "" || code == "" {
		c.logger.Warn("Username or code empty in add item request")
		return false, fmt.Errorf("Username and code must be provided and not empty")
	}

	if quantity == 0 {
		c.logger.Warn("Quantity equal to zero in add item request")
		return false, fmt.Errorf("Quantity must be provided and greather than zero")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.AddItemRequest{
		Username:    username,
		ProductCode: code,
		Quantity:    quantity,
	}

	res, err := client.AddItem(ctx, req)
	if err != nil {
		c.logger.Error("Add item request error for user '%s', product '%s' and quantity=%d: %v", username, code, quantity, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during addition of n=%d products with code '%s' to the user '%s' cart", quantity, code, username)
		return false, fmt.Errorf("Add item failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Addition of n=%d products '%s' to the user '%s' cart failed: %s", quantity, code, username, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Successful addition of n=%d products '%s' to the user '%s' cart", quantity, code, username)
	return true, nil
}

// RemoveItem attempts ro remove a product from the user's cart
func (c *CartClient) RemoveItem(username, code string) (bool, error) {
	if username == "" || code == "" {
		c.logger.Warn("Username or code empty in remove item request")
		return false, fmt.Errorf("Username and code must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.RemoveItemRequest{
		Username:    username,
		ProductCode: code,
	}

	res, err := client.RemoveItem(ctx, req)
	if err != nil {
		c.logger.Error("Remove item error for user '%s' and product '%s': %v", username, code, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error while removing product with code '%s' from the user '%s' cart", code, username)
		return false, fmt.Errorf("Remove item failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed removal of product '%s' from the user '%s' cart: %s", code, username, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Product with code '%s' successfully removed from the cart of user '%s'", code, username)
	return true, nil
}

// UpdateItemQuantity attempts to update the quantity of a product into the user's cart
func (c *CartClient) UpdateItemQuantity(username, code string, quantity uint32) (bool, error) {
	if username == "" || code == "" {
		c.logger.Warn("Username or code empty in update item quantity request")
		return false, fmt.Errorf("Username and code must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.UpdateItemQuantityRequest{
		Username:    username,
		ProductCode: code,
		Quantity:    quantity,
	}

	res, err := client.UpdateItemQuantity(ctx, req)
	if err != nil {
		c.logger.Error("Update error for quantity of products '%s' into the cart of user '%s': %v", code, username, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error while updating quantity of product with code '%s' into the cart of user '%s'", code, username)
		return false, fmt.Errorf("Update item quantity failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Update quantity failed for product '%s' into the cart of user '%s': %s", code, username, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Successful update of the quantity of products '%s' into the cart of user '%s'", code, username)
	return true, nil
}

// ListCartItems retrievs the list of items into the user's cart
func (c *CartClient) ListCartItems(username string) (bool, []*pb.CartItem, error) {
	if username == "" {
		c.logger.Warn("Username empty in list cart items request")
		return false, nil, fmt.Errorf("Username must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, nil, err
	}

	req := &pb.ListCartItemsRequest{
		Username: username,
	}

	res, err := client.ListCartItems(ctx, req)
	if err != nil {
		c.logger.Error("Error while retrieving all cart items for user '%s': %v", username, err)
		return false, nil, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error while retrieving all cart items for user '%s'", username)
		return false, nil, fmt.Errorf("Cart items retrieval failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed retrieval of all cart items for user '%s': %s", username, res.GetErrorMessage())
		return false, nil, nil
	}

	c.logger.Info("Successful retrieval of all cart items for user '%s'", username)
	return true, res.GetItems(), nil
}

// ClearCart attempts to remove all the products from the user's carts
func (c *CartClient) ClearCart(username string) (bool, error) {
	if username == "" {
		c.logger.Warn("Username empty in clear cart request")
		return false, fmt.Errorf("Username must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.ClearCartRequest{
		Username: username,
	}

	res, err := client.ClearCart(ctx, req)
	if err != nil {
		c.logger.Error("Error while removing all cart items of user '%s': %v", username, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error while removing all cart items of user '%s'", username)
		return false, fmt.Errorf("Cart items removal failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed removal of all cart items for user '%s': %s", username, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Successful removal of all cart items for user '%s'", username)
	return true, nil
}

// getClient retrieves a new CartServiceClient connected to the service.
//
// Returns:
//   - pb.CartServiceClient: the gRPC client to communicate with the cart service
//   - error: non-nil if connection setup failed
func (c *CartClient) getClient() (pb.CartServiceClient, error) {
	conn, err := c.sm.GetConnWithTimeout(c.serviceName, c.timeout)
	if err != nil {
		return nil, err
	}
	return pb.NewCartServiceClient(conn), nil
}
