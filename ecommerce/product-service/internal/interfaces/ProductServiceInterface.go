package interfaces

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"

type ProductServiceInterface interface {
	// CreateProduct creates a new product
	CreateProduct(prod *pb.Product) (bool, error)

	// GetProduct retrieves the product related to the given code
	GetProduct(code string) (bool, *pb.Product, error)

	// UpdateProduct updates the product related to the given code
	UpdateProduct(code string, prod *pb.Product) (bool, error)

	// DeleteProduct deletes the product related to the given code.
	DeleteProduct(code string) (bool, error)

	// ListProducts retrieves the list of all products
	ListProducts() (bool, []*pb.Product, error)
}