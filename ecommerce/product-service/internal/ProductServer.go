package internal

import (
	"context"
	"fmt"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/product-service/pkg/repository"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
	db     repository.ProductServiceInterface
	logger logger.Logger
}

func NewProductServer(db repository.ProductServiceInterface, logger logger.Logger) *ProductServer {
	return &ProductServer{
		db:     db,
		logger: logger,
	}
}

// CreateProduct creates a new product in the database
func (s *ProductServer) CreateProduct(ctx context.Context, in *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	if in.Product.Code == "" || in.Product.Name == "" || in.Product.Color == "" || in.Product.Description == "" {
		s.logger.Warn("Code, name, color or description empty in create product request")
		return &pb.CreateProductResponse{
			Success:      false,
			ErrorMessage: "Code, name, color and description must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Code, name, color and description must be provided and not empty")
	}

	if in.Product.Size == pb.Size_UNSPECIFIED {
		s.logger.Warn("Unspecified size in create product request")
		return &pb.CreateProductResponse{
			Success:      false,
			ErrorMessage: "The product size must be provided and must not be unspecified",
		}, status.Error(codes.InvalidArgument, "The product size must be provided and must not be unspecified")
	}

	if in.Product.Stock < 0 {
		s.logger.Warn("Negative stock values in create product request")
		return &pb.CreateProductResponse{
			Success:      false,
			ErrorMessage: "The product stock must be provided and must be non-negative",
		}, status.Error(codes.InvalidArgument, "The product stock must be provided and must be non-negative")
	}

	succ, err := s.db.CreateProduct(in.Product)
	if err != nil {
		s.logger.Error("Internal error during creation of product with code '%s': %v", in.Product.Code, err)
		return &pb.CreateProductResponse{
			Success:      false,
			ErrorMessage: "Internal server error during the creation of the product",
		}, err
	}
	if !succ {
		s.logger.Warn("Creation of the product with code '%s' failed: a product with the same code already exists", in.Product.Code)
		return &pb.CreateProductResponse{
			Success:      false,
			ErrorMessage: "The product is already existing",
		}, nil
	}

	s.logger.Info("Successful creation of the product with code '%s'", in.Product.Code)
	return &pb.CreateProductResponse{
		Success: true,
	}, nil
}

// GetProduct retrieves the product in the database related to the given code
func (s *ProductServer) GetProduct(ctx context.Context, in *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	if in.Code == "" {
		s.logger.Warn("Code empty in get product request")
		return &pb.GetProductResponse{
			Success:      false,
			Product:      nil,
			ErrorMessage: "Code must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Code must be provided and not empty")
	}

	succ, prod, err := s.db.GetProduct(in.Code)
	if err != nil {
		s.logger.Error("Internal error while retrieving product with code '%s': %v", in.Code, err)
		return &pb.GetProductResponse{
			Success:      false,
			Product:      nil,
			ErrorMessage: "Internal server error while retrieving the product",
		}, err
	}
	if !succ {
		s.logger.Warn("Product retrieval failed for product with code '%s'", in.Code)
		return &pb.GetProductResponse{
			Success:      false,
			Product:      nil,
			ErrorMessage: fmt.Sprintf("Product with code '%s' not found", in.Code),
		}, nil
	}

	s.logger.Info("Successful retrieval of the product with code '%s'", in.Code)
	return &pb.GetProductResponse{
		Success: true,
		Product: prod,
	}, nil
}

// UpdateProduct updates the product in the database related to the given code
func (s *ProductServer) UpdateProduct(ctx context.Context, in *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	if in.Code == "" {
		s.logger.Warn("Code empty in update product request")
		return &pb.UpdateProductResponse{
			Success:      false,
			ErrorMessage: "Code must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Code must be provided and not empty")
	}

	succ, err := s.db.UpdateProduct(in.Code, in.Product)
	if err != nil {
		s.logger.Error("Internal error while updating product with code '%s': %v", in.Code, err)
		return &pb.UpdateProductResponse{
			Success:      false,
			ErrorMessage: "Internal server error while updating the product",
		}, err
	}
	if !succ {
		s.logger.Warn("Product update failed for product with code '%s'", in.Code)
		return &pb.UpdateProductResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Product with code '%s' not found", in.Code),
		}, nil
	}

	s.logger.Info("Successful update of the product with code '%s'", in.Code)
	return &pb.UpdateProductResponse{
		Success: true,
	}, nil
}

// DeleteProduct deletes the product in the database related to the given code
func (s *ProductServer) DeleteProduct(ctx context.Context, in *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	if in.Code == "" {
		s.logger.Warn("Code empty in delete product request")
		return &pb.DeleteProductResponse{
			Success:      false,
			ErrorMessage: "Code must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Code must be provided and not empty")
	}

	succ, err := s.db.DeleteProduct(in.Code)
	if err != nil {
		s.logger.Error("Internal error while deleting product with code '%s': %v", in.Code, err)
		return &pb.DeleteProductResponse{
			Success:      false,
			ErrorMessage: "Internal server error while deleting the product",
		}, err
	}
	if !succ {
		s.logger.Warn("Product delete failed for product with code '%s'", in.Code)
		return &pb.DeleteProductResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Product with code '%s' not found", in.Code),
		}, nil
	}

	s.logger.Info("Successful delete of the product with code '%s'", in.Code)
	return &pb.DeleteProductResponse{
		Success: true,
	}, nil
}

// ListProducts retrieves the list of all products
func (s *ProductServer) ListProducts(ctx context.Context, in *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	succ, products, err := s.db.ListProducts()
	if err != nil {
		s.logger.Error("Internal error while retrieving all products: %v", err)
		return &pb.ListProductsResponse{
			Success:      false,
			Products:     nil,
			ErrorMessage: "Internal server error while retrieving all products",
		}, err
	}
	if !succ {
		s.logger.Warn("Retrieval of all products failed")
		return &pb.ListProductsResponse{
			Success:      false,
			Products:     nil,
			ErrorMessage: "Retrieval of all products failed",
		}, nil
	}

	s.logger.Info("Successful retrieval of all users")
	return &pb.ListProductsResponse{
		Success:  true,
		Products: products,
	}, nil
}
