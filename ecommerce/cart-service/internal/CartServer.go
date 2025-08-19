package internal

import (
	"context"
	"fmt"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/cart-service/internal/interfaces"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/cart"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CartServer implements the cart service gRPC server.
type CartServer struct {
	pb.UnimplementedCartServiceServer
	db     interfaces.CartServiceInterface
	logger logger.Logger
}

// NewCartServer returns a new CartServer instance
func NewCartServer(db interfaces.CartServiceInterface, logger logger.Logger) *CartServer {
	return &CartServer{
		db:     db,
		logger: logger,
	}
}

// AddItem adds a product to the specified user's cart, increasing quantity if already present
func (s *CartServer) AddItem(ctx context.Context, in *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	if in.Username == "" || in.ProductCode == "" {
		s.logger.Warn("Username or product code empty in add item request")
		return &pb.AddItemResponse{
			Success:      false,
			ErrorMessage: "Username and product code must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username and product code must be provided and not empty")
	}

	if in.Quantity == 0 {
		s.logger.Warn("Quantity equal to zero in add item request")
		return &pb.AddItemResponse{
			Success:      false,
			ErrorMessage: "The quantity must be provided and greather than zero",
		}, status.Error(codes.InvalidArgument, "The quantity must be provided and greather than zero")
	}

	succ, err := s.db.AddItem(in.Username, in.ProductCode, in.Quantity)
	if err != nil {
		s.logger.Error("Internal error while adding n=%d products '%s' to the user '%s' cart: %v", in.Quantity, in.ProductCode, in.Username, err)
		return &pb.AddItemResponse{
			Success:      false,
			ErrorMessage: "Internal server error while adding products to the cart",
		}, err
	}
	if !succ {
		s.logger.Warn("Addition of n=%d products '%s' to the user '%s' cart failed", in.Quantity, in.ProductCode, in.Username)
		return &pb.AddItemResponse{
			Success:      false,
			ErrorMessage: "Addition of the product in the cart failed",
		}, nil
	}

	s.logger.Info("Successful addition of n=%d products '%s' to the user '%s' cart", in.Quantity, in.ProductCode, in.Username)
	return &pb.AddItemResponse{
		Success: true,
	}, nil
}

// RemoveItem removes a specific product from the specified user's cart
func (s *CartServer) RemoveItem(ctx context.Context, in *pb.RemoveItemRequest) (*pb.RemoveItemResponse, error) {
	if in.Username == "" || in.ProductCode == "" {
		s.logger.Warn("Username or product code empty in remove item request")
		return &pb.RemoveItemResponse{
			Success:      false,
			ErrorMessage: "Username and product code must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username and product code must be provided and not empty")
	}

	succ, err := s.db.RemoveItem(in.Username, in.ProductCode)
	if err != nil {
		s.logger.Error("Internal error while removing product with code '%s' from the user '%s' cart: %v", in.ProductCode, in.Username, err)
		return &pb.RemoveItemResponse{
			Success:      false,
			ErrorMessage: "Internal server error while removing the product from the cart",
		}, err
	}
	if !succ {
		s.logger.Warn("Failed removal of product with code '%s' from the user '%s' cart", in.ProductCode, in.Username)
		return &pb.RemoveItemResponse{
			Success:      false,
			ErrorMessage: "Failed removal of the product: product not found in the user's cart",
		}, nil
	}

	s.logger.Info("Successful removal of the product with code '%s' from the user '%s' cart", in.ProductCode, in.Username)
	return &pb.RemoveItemResponse{
		Success: true,
	}, nil
}

// UpdateItemQuantity updates the quantity of a specific product in the specified user's cart
func (s *CartServer) UpdateItemQuantity(ctx context.Context, in *pb.UpdateItemQuantityRequest) (*pb.UpdateItemQuantityResponse, error) {
	if in.Username == "" || in.ProductCode == "" {
		s.logger.Warn("Username or product code empty in update item request")
		return &pb.UpdateItemQuantityResponse{
			Success:      false,
			ErrorMessage: "Username and product code must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username and product code must be provided and not empty")
	}

	succ, err := s.db.UpdateItemQuantity(in.Username, in.ProductCode, in.Quantity)
	if err != nil {
		s.logger.Error("Internal error while updating quantity for product '%s' into the cart of user '%s': %v", in.ProductCode, in.Username, err)
		return &pb.UpdateItemQuantityResponse{
			Success:      false,
			ErrorMessage: "Internal server error while updating the quantity of the product into the user's cart",
		}, err
	}
	if !succ {
		s.logger.Warn("No cart item found for user '%s' and product with code '%s'", in.Username, in.ProductCode)
		return &pb.UpdateItemQuantityResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Product with code '%s' not present in the cart", in.ProductCode),
		}, nil
	}

	s.logger.Info("Quantity of products with code '%s' in the cart of the user '%s' successfully updated", in.ProductCode, in.Username)
	return &pb.UpdateItemQuantityResponse{
		Success: true,
	}, nil
}

// ListCartItems retrieves all cart items for the specified user
func (s *CartServer) ListCartItems(ctx context.Context, in *pb.ListCartItemsRequest) (*pb.ListCartItemsResponse, error) {
	if in.Username == "" {
		s.logger.Warn("Username empty in list user items request")
		return &pb.ListCartItemsResponse{
			Success:      false,
			ErrorMessage: "Username must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username must be provided and not empty")
	}

	succ, items, err := s.db.ListCartItems(in.Username)
	if err != nil {
		s.logger.Error("Internal error while retrieving all items in the cart of the user '%s': %v", in.Username, err)
		return &pb.ListCartItemsResponse{
			Success:      false,
			Items:        nil,
			ErrorMessage: "Internal server error while retrieving all products from the cart",
		}, err
	}
	if !succ {
		s.logger.Warn("Failed retrieval of all products in the cart of user '%s'", in.Username)
		return &pb.ListCartItemsResponse{
			Success:      false,
			Items:        nil,
			ErrorMessage: "No products found in the user's cart",
		}, nil
	}

	s.logger.Info("Successful retrieval of all products in the cart of user '%s'", in.Username)
	return &pb.ListCartItemsResponse{
		Success: true,
		Items:   items,
	}, nil
}

// ClearCart removes all products from the specified user's cart
func (s *CartServer) ClearCart(ctx context.Context, in *pb.ClearCartRequest) (*pb.ClearCartResponse, error) {
	if in.Username == "" {
		s.logger.Warn("Username empty in clear cart request")
		return &pb.ClearCartResponse{
			Success:      false,
			ErrorMessage: "Username must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username must be provided and not empty")
	}

	succ, err := s.db.ClearCart(in.Username)
	if err != nil {
		s.logger.Error("Internal error while clearing the cart of the user '%s': %v", in.Username, err)
		return &pb.ClearCartResponse{
			Success:      false,
			ErrorMessage: "Internal server error while clearing the cart",
		}, err
	}
	if !succ {
		s.logger.Warn("Failed removal of all products from the cart of the user '%s'", in.Username)
		return &pb.ClearCartResponse{
			Success:      false,
			ErrorMessage: "User's cart already empty",
		}, nil
	}

	s.logger.Info("Successful removal of all the product from the user '%s' cart", in.Username)
	return &pb.ClearCartResponse{
		Success: true,
	}, nil
}

// RemoveProductFromAllCarts removes all cart items related to a given product
func (s *CartServer) RemoveProductFromAllCarts(ctx context.Context, in *pb.RemoveProductFromAllCartsRequest) (*pb.RemoveProductFromAllCartsResponse, error) {
	if in.ProductCode == "" {
		s.logger.Warn("Product code empty in remove product from all carts request")
		return &pb.RemoveProductFromAllCartsResponse{
			Success:      false,
			ErrorMessage: "Product code must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Product code must be provided and not empty")
	}

	succ, err := s.db.RemoveProductFromAllCarts(in.ProductCode)
	if err != nil {
		s.logger.Error("Internal error while removing the product with code '%s' from all carts: %v", in.ProductCode, err)
		return &pb.RemoveProductFromAllCartsResponse{
			Success:      false,
			ErrorMessage: "Internal server error while removing the product from all carts",
		}, err
	}
	if !succ {
		s.logger.Warn("Failed removal of the product with code '%s' from all carts", in.ProductCode)
		return &pb.RemoveProductFromAllCartsResponse{
			Success:      false,
			ErrorMessage: "No carts found containing the product",
		}, nil
	}

	s.logger.Info("Suceìcessful removal of product with code '%s' from all carts", in.ProductCode)
	return &pb.RemoveProductFromAllCartsResponse{
		Success: true,
	}, nil
}
