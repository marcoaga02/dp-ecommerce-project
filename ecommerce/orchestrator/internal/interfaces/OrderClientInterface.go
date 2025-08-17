package interfaces

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"

type OrderClientInterface interface {
	// CreateOrder creates a new order
	CreateOrder(username string, items []*pb.OrderItem) (bool, int32, error)

	// GetOrder retrieves the order with the given id
	GetOrder(orderId int32) (bool, *pb.Order, error)

	// UpdateOrderStatus updates the status of the order with the given id
	UpdateOrderStatus(orderId int32, status pb.Status) (bool, error)

	// ListOrdersByUsername retrieves the list of all orders of a given user
	ListOrdersByUsername(username string) (bool, []*pb.Order, error)

	// ListAllOrders retrieves the list of all orders
	ListAllOrders() (bool, []*pb.Order, error)
}