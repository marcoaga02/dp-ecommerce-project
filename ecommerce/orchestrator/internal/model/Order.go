package model

import (
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"
)

type Order struct {
	ID          int32
	Username    string
	Status      Status
	Items       []*OrderItem
	TotalAmount float64
}

// ProtoOrderToModelOrder converts a pb.Order into a model.Order
func ProtoOrderToModelOrder(order *pb.Order) *Order {
	if order == nil {
		return nil
	}
	return &Order{
		ID:          order.GetOrderId(),
		Username:    order.GetUsername(),
		Status:      ProtoStatusToModelStatus(order.GetStatus()),
		Items:       ProtoOrderItemsListToModelOrderItemsList(order.GetItems()),
		TotalAmount: order.TotalAmount,
	}
}

// ProtoOrdersListToModelOrdersList converts a []*pb.Order into a []*model.Order
func ProtoOrdersListToModelOrdersList(orders []*pb.Order) []*Order {
	if orders == nil {
		return nil
	}

	var modelOrders []*Order

	for _, order := range orders {
		modelOrders = append(modelOrders, ProtoOrderToModelOrder(order))
	}
	return modelOrders
}
