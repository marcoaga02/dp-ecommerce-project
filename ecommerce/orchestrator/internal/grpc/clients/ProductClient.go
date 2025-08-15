package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/manager"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"
)

// ProductClient is a gRPC client for communicating with the product service
//
// Implements the ProductClientInterface
type ProductClient struct {
	serviceName string
	sm          *manager.ServiceManager
	logger      logger.Logger
	timeout     time.Duration
}

// NewProductClient return a new instance of ProductClient
func NewProductClient(serviceName string, sm *manager.ServiceManager, log logger.Logger, timeout time.Duration) *ProductClient {
	return &ProductClient{
		serviceName: serviceName,
		sm:          sm,
		logger:      log,
		timeout:     timeout,
	}
}

// CreateProduct creates a new product.
func (c *ProductClient) CreateProduct(code, name string, size pb.Size, color, description string, stock uint32, price float64) (bool, error) {
	if code == "" || name == "" || color == "" || description == "" {
		c.logger.Warn("Code, name, color or description empty in create product request")
		return false, fmt.Errorf("Code, name, color and description must be provided and not empty")
	}

	if size == pb.Size_UNSPECIFIED {
		c.logger.Warn("Unspecified size in create product request")
		return false, fmt.Errorf("The product size must be provided and must not be unspecified")
	}

	if stock < 0 {
		c.logger.Warn("Negative stock values in create product request")
		return false, fmt.Errorf("The product stock must be provided and must be non-negative")
	}

	if price < 0 {
		c.logger.Warn("Negative price in create product request")
		return false, fmt.Errorf("The product price must be provided and must be non-negative")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.CreateProductRequest{
		Product: &pb.Product{
			Code:        code,
			Name:        name,
			Size:        size,
			Color:       color,
			Description: description,
			Stock:       stock,
			Price:       price,
		},
	}

	res, err := client.CreateProduct(ctx, req)
	if err != nil {
		c.logger.Error("Create product error for product with code '%s': %v", code, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during creation of product with code '%s'", code)
		return false, fmt.Errorf("Create product failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed creation of product with code '%s': %s", code, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Creation succeeded of product with code '%s'", code)
	return true, nil
}

// GetProduct retrieves the product related to the given code.
func (c *ProductClient) GetProduct(code string) (bool, *pb.Product, error) {
	if code == "" {
		c.logger.Error("Code empty in create product request")
		return false, nil, fmt.Errorf("Code must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, nil, err
	}

	req := &pb.GetProductRequest{
		Code: code,
	}

	res, err := client.GetProduct(ctx, req)
	if err != nil {
		c.logger.Error("Error during the retrieval of the product with code '%s': %v", code, err)
		return false, nil, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during retrieval of product with code '%s'", code)
		return false, nil, fmt.Errorf("Product retrieval failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Error("Failed retrieval of product with code '%s': %s", code, res.GetErrorMessage())
		return false, nil, nil
	}

	c.logger.Info("Successful retrieval of product with code '%s'", code)
	return true, res.GetProduct(), nil
}

// UpdateProduct updates the product related to the given code.
func (c *ProductClient) UpdateProduct(code, name string, size pb.Size, color, description string, stock uint32, price float64) (bool, error) {
	if code == "" {
		c.logger.Error("Code empty in create product request")
		return false, fmt.Errorf("Code must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.UpdateProductRequest{
		Code: code,
		Product: &pb.Product{
			Code:        code,
			Name:        name,
			Size:        size,
			Color:       color,
			Description: description,
			Stock:       stock,
			Price:       price,
		},
	}

	res, err := client.UpdateProduct(ctx, req)
	if err != nil {
		c.logger.Error("Update error for product with code '%s': %v", code, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during update for product with code '%s'", code)
		return false, fmt.Errorf("Product update failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Update failed for product with code '%s': %s", code, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Update succeeded for product with code '%s'", code)
	return true, nil
}

// DeleteProduct deletes the product related to the given code.
func (c *ProductClient) DeleteProduct(code string) (bool, error) {
	if code == "" {
		c.logger.Error("Code empty in create product request")
		return false, fmt.Errorf("Code must be provided and not empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, err
	}

	req := &pb.DeleteProductRequest{
		Code: code,
	}

	res, err := client.DeleteProduct(ctx, req)
	if err != nil {
		c.logger.Error("Delete error for product with code '%s': %v", code, err)
		return false, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during delete of product with code '%s'", code)
		return false, fmt.Errorf("Delete product failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Delete failed for product with code '%s': %s", code, res.GetErrorMessage())
		return false, nil
	}

	c.logger.Info("Product with code '%s' successfully deleted", code)
	return true, nil
}

// ListProducts retrieves the list of all products
func (c *ProductClient) ListProducts() (bool, []*pb.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	client, err := c.getClient()
	if err != nil {
		c.logger.Error("Failed to get connection to service '%s': %v", c.serviceName, err)
		return false, nil, err
	}

	res, err := client.ListProducts(ctx, &pb.ListProductsRequest{})
	if err != nil {
		c.logger.Error("Error during the retrieval of all products: %v", err)
		return false, nil, err
	}
	if res == nil {
		c.logger.Error("Received nil response without error during the retrieval of all products")
		return false, nil, fmt.Errorf("Products retrieval failed: received nil response without error")
	}
	if !res.GetSuccess() {
		c.logger.Warn("Failed retrieval of all products: %s", res.GetErrorMessage())
		return false, nil, nil
	}

	c.logger.Info("Successful retrieval of all products")
	return true, res.GetProducts(), nil
}

// getClient retrieves a new ProductServiceClient connected to the service.
//
// Returns:
//   - pb.ProductServiceClient: the gRPC client to communicate with the product service
//   - error: non-nil if connection setup failed
func (c *ProductClient) getClient() (pb.ProductServiceClient, error) {
	conn, err := c.sm.GetConnWithTimeout(c.serviceName, c.timeout)
	if err != nil {
		return nil, err
	}
	return pb.NewProductServiceClient(conn), nil
}
