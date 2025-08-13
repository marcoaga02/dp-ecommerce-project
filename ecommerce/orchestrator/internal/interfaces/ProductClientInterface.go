package interfaces

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"

type ProductClientInterface interface {
	// CreateProduct creates a new product.
	CreateProduct(code, name string, size pb.Size, color, description string, stock int32, price float64) (bool, error)

	// GetProduct retrieves the product related to the given code.
	GetProduct(code string) (bool, *pb.Product, error)

	// UpdateProduct updates the product related to the given code.
	UpdateProduct(code, name string, size pb.Size, color, description string, stock int32, price float64) (bool, error)

	// DeleteProduct deletes the product related to the given code.
	DeleteProduct(code string) (bool, error)

	// ListProducts retrieves the list of all products
	ListProducts() (bool, []*pb.Product, error)
}
