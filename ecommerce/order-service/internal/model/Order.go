package model

import (
	"fmt"

	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"
)

type Order struct {
	ID       int32       `gorm:"column:id;primaryKey;autoIncrement"`
	Username string      `gorm:"column:username;type:varchar(32);not null"`
	StatusID int         `gorm:"column:status_id;not null;default:1"` // foreign key
	Status   Status      `gorm:"foreignKey:StatusID;references:ID"`   // GORM relationship
	Items    []OrderItem `gorm:"foreignKey:OrderID;references:ID"`
}

// ModelOrderToProtoOrder converts a model.Order into a pb.Order
func ModelOrderToProtoOrder(order *Order) (*pb.Order, error) {
	if order == nil {
		return nil, fmt.Errorf("Input argument is nil")
	}

	status := ModelStatusToProtoStatus(order.StatusID)
	if status == StatusUnspecified {
		return nil, fmt.Errorf("Invalid order status id '%d'", order.StatusID)
	}

	pbOrder := pb.Order{
		OrderId: order.ID,
		Username: order.Username,
		Status: status,
	}
	
	pbOrder.Items = make([]*pb.OrderItem, 0, len(order.Items))

	var totalAmount float64 = 0

	for _, modelItem := range order.Items {
		pbItem := ModelOrderItemToProtoOrderItem(&modelItem)

		pbOrder.Items = append(pbOrder.Items, pbItem)
		totalAmount = totalAmount + pbItem.TotalPrice
	}
	
	pbOrder.TotalAmount = totalAmount
	return &pbOrder, nil
}
