package internal

import (
	"context"
	"fmt"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/order-service/internal/interfaces"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const errOrderID int32 = -1

// OrderServer implements the order service gRPC server.
type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	db     interfaces.OrderServiceInterface
	logger logger.Logger
}

// NewOrderServer return a new OrderServer instance
func NewOrderServer(db interfaces.OrderServiceInterface, logger logger.Logger) *OrderServer {
	return &OrderServer{
		db:     db,
		logger: logger,
	}
}

// CreateOrder creates a new order
func (s *OrderServer) CreateOrder(ctx context.Context, in *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if in.Username == "" {
		s.logger.Warn("Username empty in create order request")
		return &pb.CreateOrderResponse{
			Success:      false,
			OrderId:      errOrderID,
			ErrorMessage: "Username must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username must be provided and not empty")
	}

	if in.Items == nil || len(in.Items) == 0 {
		s.logger.Warn("Items list empty in create order request")
		return &pb.CreateOrderResponse{
			Success:      false,
			OrderId:      errOrderID,
			ErrorMessage: "At least one item must be provided in the items list",
		}, status.Error(codes.InvalidArgument, "At least one item must be provided in the items list")
	}

	for i, item := range in.Items {
		if item.ProductCode == "" || item.Name == "" {
			s.logger.Warn("Item number %d has empty product code or name", i)
			return &pb.CreateOrderResponse{
				Success:      false,
				OrderId:      errOrderID,
				ErrorMessage: fmt.Sprintf("Item number %d has empty product code or name", i),
			}, status.Error(codes.InvalidArgument, "All product codes and names must be provided and not empty")
		}

		if item.Price < 0 {
			s.logger.Warn("Item number %d has negative price", i)
			return &pb.CreateOrderResponse{
				Success:      false,
				OrderId:      errOrderID,
				ErrorMessage: fmt.Sprintf("Item number %d has negative price", i),
			}, status.Error(codes.InvalidArgument, "All product prices must be non-negative")
		}

		if item.Quantity == 0 {
			s.logger.Warn("Item number %d has quantity equal to zero", i)
			return &pb.CreateOrderResponse{
				Success:      false,
				OrderId:      errOrderID,
				ErrorMessage: fmt.Sprintf("Item number %d has quantity equal to zero", i),
			}, status.Error(codes.InvalidArgument, "All product quantities must be positive")
		}
	}

	itemsNumber := len(in.Items)
	succ, orderId, err := s.db.CreateOrder(in.Username, in.Items)
	if err != nil {
		s.logger.Error("Internal error while creating order of %d products for user '%s': %v", itemsNumber, in.Username, err)
		return &pb.CreateOrderResponse{
			Success:      false,
			OrderId:      errOrderID,
			ErrorMessage: "Internal server error while creating the order",
		}, err
	}
	if !succ {
		s.logger.Warn("Creation of %d products order for user '%s' failed", itemsNumber, in.Username)
		return &pb.CreateOrderResponse{
			Success:      false,
			OrderId:      errOrderID,
			ErrorMessage: "Creation of the order failed",
		}, nil
	}

	s.logger.Info("Successful creation of the %d products order for user '%s'", itemsNumber)
	return &pb.CreateOrderResponse{
		Success: true,
		OrderId: orderId,
	}, nil
}

// GetOrder retrieves the order with the given id
func (s *OrderServer) GetOrder(ctx context.Context, in *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	if in.OrderId <= 0 {
		s.logger.Warn("Order ID less or equal to zero in get order request")
		return &pb.GetOrderResponse{
			Success:      false,
			Order:        nil,
			ErrorMessage: "Order ID must be positive",
		}, status.Error(codes.InvalidArgument, "Order ID must be positive")
	}

	succ, order, err := s.db.GetOrder(in.OrderId)
	if err != nil {
		s.logger.Error("Internal error while retrieving order with ID '%d': %v", in.OrderId, err)
		return &pb.GetOrderResponse{
			Success:      false,
			Order:        nil,
			ErrorMessage: "Internal server error while retrieving the order",
		}, err
	}
	if !succ {
		s.logger.Warn("Failed retrieval of order with ID '%d'", in.OrderId)
		return &pb.GetOrderResponse{
			Success:      false,
			Order:        nil,
			ErrorMessage: fmt.Sprintf("Order with ID '%d' not found", in.OrderId),
		}, nil
	}

	s.logger.Info("Successful retrieval of order with ID '%d'", in.OrderId)
	return &pb.GetOrderResponse{
		Success: true,
		Order:   order,
	}, nil
}

// UpdateOrderStatus updates the status of the order with the given id
func (s *OrderServer) UpdateOrderStatus(ctx context.Context, in *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	res, err := s.GetOrder(ctx, &pb.GetOrderRequest{
		OrderId: in.OrderId,
	})
	if err != nil { // error already logged in the GetOrder method
		return &pb.UpdateOrderStatusResponse{
			Success:      false,
			ErrorMessage: "Internal error while retrieving order",
		}, err
	}
	if !res.GetSuccess() || res.GetOrder() == nil {
		return &pb.UpdateOrderStatusResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Order with ID '%d' not found", in.OrderId),
		}, nil
	}

	currentStatus := res.GetOrder().GetStatus()

	if currentStatus == in.Status {
		s.logger.Info("Order ID '%d' already has status '%s'", in.OrderId, in.Status.String())
		return &pb.UpdateOrderStatusResponse{
			Success: true,
		}, nil
	}

	if !isStatusChangeAllowed(currentStatus, in.Status) {
		s.logger.Warn("Invalid status transition from '%s' to '%s' for order ID '%d'", currentStatus.String(), in.Status.String(), in.OrderId)
		return &pb.UpdateOrderStatusResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Cannot change status from '%s' to '%s'", currentStatus.String(), in.Status.String()),
		}, nil
	}

	succ, err := s.db.UpdateOrderStatus(in.OrderId, in.Status)
	if err != nil {
		s.logger.Error("Internal error while updating status for order with ID '%d': %v", in.OrderId, err)
		return &pb.UpdateOrderStatusResponse{
			Success:      false,
			ErrorMessage: "Internal server error while updating the status for the order",
		}, err
	}
	if !succ {
		s.logger.Warn("No order found with ID '%d'", in.OrderId)
		return &pb.UpdateOrderStatusResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Order with ID '%d' not found", in.OrderId),
		}, nil
	}

	s.logger.Info("Status of order with ID '%d' successfully changed to '%s'", in.OrderId, in.Status.String())
	return &pb.UpdateOrderStatusResponse{
		Success: true,
	}, nil
}

// ListOrdersByUsername retrieves the list of all orders of a given user
func (s *OrderServer) ListOrdersByUsername(ctx context.Context, in *pb.ListOrdersByUsernameRequest) (*pb.ListOrdersByUsernameResponse, error) {
	if in.Username == "" {
		s.logger.Warn("Username empty in list orders by username request")
		return &pb.ListOrdersByUsernameResponse{
			Success:      false,
			Orders:       nil,
			ErrorMessage: "Username must be provided and not empty",
		}, status.Error(codes.InvalidArgument, "Username must be provided and not empty")
	}

	succ, orders, err := s.db.ListOrdersByUsername(in.Username)
	if err != nil {
		s.logger.Error("Internal error while retrieving all orders for user '%s': %v", in.Username, err)
		return &pb.ListOrdersByUsernameResponse{
			Success:      false,
			Orders:       nil,
			ErrorMessage: "Internal server error while retrieving all orders",
		}, nil
	}
	if !succ {
		s.logger.Warn("Failed retrieval of all orders for user '%s'", in.Username)
		return &pb.ListOrdersByUsernameResponse{
			Success:      false,
			Orders:       nil,
			ErrorMessage: "No orders found for the user",
		}, nil
	}

	s.logger.Info("Successful retrieval of all orders for user '%s'", in.Username)
	return &pb.ListOrdersByUsernameResponse{
		Success: true,
		Orders:  orders,
	}, nil
}

// ListAllOrders retrieves the list of all orders
func (s *OrderServer) ListAllOrders(ctx context.Context, in *pb.ListAllOrdersRequest) (*pb.ListAllOrdersResponse, error) {
	succ, orders, err := s.db.ListAllOrders()
	if err != nil {
		s.logger.Error("Internal error while retrieving all orders: %v", err)
		return &pb.ListAllOrdersResponse{
			Success:      false,
			Orders:       nil,
			ErrorMessage: "Internal server error while retrieving all the orders",
		}, err
	}
	if !succ {
		s.logger.Warn("Failed retrieval of all the orders")
		return &pb.ListAllOrdersResponse{
			Success:      false,
			Orders:       nil,
			ErrorMessage: "No orders found",
		}, nil
	}

	s.logger.Info("Successful retrieval of all orders")
	return &pb.ListAllOrdersResponse{
		Success: true,
		Orders:  orders,
	}, nil
}

// isStatusChangeAllowed checks if the transition from one status to another is allowed or not
func isStatusChangeAllowed(current, next pb.Status) bool {
	validTransitions := map[pb.Status][]pb.Status{
		pb.Status_UNSPECIFIED: {},
		pb.Status_PROCESSING:  {pb.Status_SHIPPED, pb.Status_CANCELED},
		pb.Status_SHIPPED:     {pb.Status_DELIVERED, pb.Status_CANCELED},
		pb.Status_DELIVERED:   {},
		pb.Status_CANCELED:    {},
	}

	for _, s := range validTransitions[current] {
		if s == next {
			return true
		}
	}
	return false
}
