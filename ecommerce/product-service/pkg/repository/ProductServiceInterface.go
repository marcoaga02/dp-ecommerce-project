package repository

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"

type ProductServiceInterface interface {
	CreateProduct(prod *pb.Product) (bool, error)
	GetProduct(code string) (bool, *pb.Product, error)
	UpdateProduct(code string, prod *pb.Product) (bool, error)
	DeleteProduct(code string) (bool, error)
	ListProducts() (bool, []*pb.Product, error)
}