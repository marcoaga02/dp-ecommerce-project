package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/cart"

type CartItem struct {
	Username string `gorm:"column:username;type:varchar(32);primaryKey;not null"`
	Code     string `gorm:"column:code;type:varchar(32);primaryKey;not null"`
	Quantity uint32 `gorm:"column:quantity;type:int unsigned;not null"`
}

// ModelItemToProtoItem converts a model.CartItem into a pb.CartItem
func ModelItemToProtoItem(cartItem *CartItem) *pb.CartItem {
	if cartItem == nil {
		return nil
	}

	return &pb.CartItem{
		Username: cartItem.Username,
		ProductCode: cartItem.Code,
		Quantity: cartItem.Quantity,
	}
}