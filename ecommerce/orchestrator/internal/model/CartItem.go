package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/cart"

type CartItem struct {
	ProdCode string
	Quantity uint32
}

// ProtoCartItemToModelCartItem converts a pb.CartItem into a model.CartItem
func ProtoCartItemToModelCartItem(item *pb.CartItem) *CartItem {
	if item == nil {
		return nil
	}

	return &CartItem{
		ProdCode: item.GetProductCode(),
		Quantity: item.GetQuantity(),
	}
}

func ProtoCartItemsListToModelCartItemsList(items []*pb.CartItem) []*CartItem {
	if items == nil {
		return nil
	}

	var modelItems []*CartItem
	for _, item := range items {
		modelItems = append(modelItems, ProtoCartItemToModelCartItem(item))
	}
	return modelItems
}