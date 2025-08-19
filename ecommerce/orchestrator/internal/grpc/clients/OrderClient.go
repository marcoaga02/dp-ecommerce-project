package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/manager"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"
)

const errOrderID int32 = -1

// OrderClient is a gRPC client for communicating with the order service
//
// Implements the OrderClientInterface
type OrderClient struct {
	serviceName string
	sm          *manager.ServiceManager
	logger      logger.Logger
	timeout     time.Duration
}

// NewOrderClient returns a new instance of OrderClient
func NewOrderClient(serviceName string, sm *manager.ServiceManager, log logger.Logger, timeout time.Duration) *OrderClient {
	return &OrderClient{
		serviceName: serviceName,
		sm:          sm,
		logger:      log,
		timeout:     timeout,
	}
}

// CreateOrder creates a new order
func (c *OrderClient) CreateOrder(username string, items []*pb.OrderItem) (bool, int32, error) {
	if username == "" {
		c.logger.Warn("Username empty in create order request")
		return false, errOrderID, fmt.Errorf("Username must be provided and not empty")
	}

	if items == nil || len(items) == 0 {
		c.logger.Warn("Items list empty in create order request")
		return false, errOrderID, fmt.Errorf("At least one item must be provided in the items list")
	}

	for i, item := range items {
		if item.ProductCode == "" || item.Name == "" {
			c.logger.Warn("Item number %d has empty product code or name", i)
			return false, errOrderID, fmt.Errorf("All product codes and names must be provided and not empty")
		}

		if item.Price < 0 {
			c.logger.Warn("Item number %d has negative price", i)
			return false, errOrderID, fmt.Errorf("All product prices must be non-negative")
		}

		if item.Quantity == 0 {
			c.logger.Warn("Item number %d has quantity equal to zero", i)
			return false, errOrderID, fmt.Errorf("All product quantities must be positive")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, errOrderID, err
	}

	req := &pb.CreateOrderRequest{
		Username: username,
		Items:    items,
	}

	var itemsLen int = len(items)
	res, err := client.CreateOrder(ctx, req)
	if err != nil {
		c.logger.Error("Error during the creation of an order with %d items for user '%s': %v", itemsLen, username, err)
		return false, errOrderID, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during creation of an order with %d items for user '%s': %v", itemsLen, username)
		return false, errOrderID, fmt.Errorf("Create order failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed creation of order with %d items for user '%s': %s", itemsLen, username, res.GetErrorMessage())
		return false, errOrderID, nil
	}

	c.logger.Info("Successful creation of order with %d items for user '%s'", itemsLen, username)
	return true, res.GetOrderId(), nil
}

// GetOrder retrieves the order with the given id
func (c *OrderClient) GetOrder(orderId int32) (bool, *pb.Order, error) {
	if orderId <= 0 {
		c.logger.Warn("Order ID less or equal to zero in get order request")
		return false, nil, fmt.Errorf("Order ID must be provided and positive")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, nil, err
	}

	req := &pb.GetOrderRequest{
		OrderId: orderId,
	}

	res, err := client.GetOrder(ctx, req)
	if err != nil {
		c.logger.Error("Error during the retrieval of the order with ID '%d': %v", orderId, err)
		return false, nil, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during retrieval of order with ID '%d'", orderId)
		return false, nil, fmt.Errorf("Order retrieval failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed retrieval of order with ID '%d': %s", orderId, res.GetErrorMessage())
		return false, nil, nil
	}

	c.logger.Info("Successful retrieval of order with ID '%d'", orderId)
	return true, res.GetOrder(), nil
}

// UpdateOrderStatus updates the status of the order with the given id
func (c *OrderClient) UpdateOrderStatus(orderId int32, status pb.Status) (bool, error) {
	if orderId <= 0 {
		c.logger.Warn("Order ID less or equal to zero in update order status request")
		return false, fmt.Errorf("Order ID must be provided and positive")
	}
	if status == pb.Status_UNSPECIFIED {
		c.logger.Warn("Status unspecified in update order status request")
		return false, fmt.Errorf("Status must be provided and not unspecified")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.UpdateOrderStatusRequest{
		OrderId: orderId,
		Status:  status,
	}

	res, err := client.UpdateOrderStatus(ctx, req)
	if err != nil {
		c.logger.Error("Error while updating status to '%s' for order with ID '%d': %v", status.String(), orderId, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error while updating status to '%s' for order with ID '%d'", status.String(), orderId)
		return false, fmt.Errorf("Status update failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed update of the status to '%s' for order with ID '%d': %s", status.String(), orderId, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Successful update of the status to '%s' for order with ID '%d'", status.String(), orderId)
	return true, nil
}

// ListOrdersByUsername retrieves the list of all orders of a given user
func (c *OrderClient) ListOrdersByUsername(username string) (bool, []*pb.Order, error) {
	if username == "" {
		c.logger.Warn("Username empty in list orders by username request")
		return false, nil, fmt.Errorf("Username must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, nil, err
	}

	req := &pb.ListOrdersByUsernameRequest{
		Username: username,
	}

	res, err := client.ListOrdersByUsername(ctx, req)
	if err != nil {
		c.logger.Error("Error during the retrieval of all orders for user '%s': %v", username, err)
		return false, nil, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during the retrieval of orders for user '%s'", username)
		return false, nil, fmt.Errorf("Orders retrieval failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed retrieval of all orders for user '%s': %s", username, res.GetErrorMessage())
		return false, nil, nil
	}

	c.logger.Info("Successful retrieval of all orders for user '%s'", username)
	return true, res.GetOrders(), nil
}

// ListAllOrders retrieves the list of all orders
func (c *OrderClient) ListAllOrders() (bool, []*pb.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, nil, err
	}

	req := &pb.ListAllOrdersRequest{}

	res, err := client.ListAllOrders(ctx, req)
	if err != nil {
		c.logger.Error("Error during the retrieval of all orders: %v", err)
		return false, nil, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during the retrieval of all orders")
		return false, nil, fmt.Errorf("Orders retrieval failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed retrieval of all orders: %s", res.GetErrorMessage())
		return false, nil, nil
	}

	c.logger.Info("Successful retrieval of all orders")
	return true, res.GetOrders(), nil
}

// getClient retrieves a new OrderServiceClient connected to the service.
//
// Returns:
//   - pb.OrderServiceClient: the gRPC client to communicate with the order service
//   - error: non-nil if connection setup failed
func (c *OrderClient) getClient() (pb.OrderServiceClient, error) {
	conn, err := c.sm.GetConnWithTimeout(c.serviceName, c.timeout)
	if err != nil {
		return nil, err
	}
	return pb.NewOrderServiceClient(conn), nil
}
