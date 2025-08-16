package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"

type OrderItem struct {
	ProductCode string  `gorm:"column:product_code;type:varchar(32);primaryKey"`
	ProductName string  `gorm:"column:product_name;type:varchar(255);not null"`
	UnitPrice   float64 `gorm:"column:unit_price;type:double(10,2) unsigned;not null"`
	Quantity    uint32  `gorm:"column:quantity;type:int unsigned;not null"`
	OrderID     int32   `gorm:"column:order_id;primaryKey"`
	Order       Order   `gorm:"foreignKey:OrderID;references:ID"`
}

// ModelOrderItemToProtoOrderItem converts a model.OrderItem into a pb.OrderItem
func ModelOrderItemToProtoOrderItem(item *OrderItem) *pb.OrderItem {
	if item == nil {
		return nil
	}

	var totalPrice float64 = item.UnitPrice * float64(item.Quantity)

	return &pb.OrderItem{
		ProductCode: item.ProductCode,
		Name: item.ProductName,
		Price: item.UnitPrice,
		Quantity: item.Quantity,
		TotalPrice: totalPrice,
	}
}
