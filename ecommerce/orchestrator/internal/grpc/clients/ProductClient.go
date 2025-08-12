package clients

import (
	"time"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/manager"
)

// ProductClient is a gRPC client for communicating with the product service
type ProductClient struct {
	serviceName string
	sm *manager.ServiceManager
	logger logger.Logger
	timeout time.Duration
}

func (c *ProductClient) CreateProduct(prod *pb.Product) (bool, error) {

}

func (c *ProductClient) GetProduct(code string) (bool, *pb.Product, error) {

}

func (c *ProductClient) UpdateProduct(code string, prod *pb.Product) (bool, error) {

}

func (c *ProductClient) DeleteProduct(code string) (bool, error) {

}

func (c *ProductClient) ListProducts() (bool, []*pb.Product, error) {

}