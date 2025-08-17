package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"

type OrderItem struct {
	ProductCode string
	Name        string
	Price       float64
	Quantity    uint32
}

// ProtoOrderItemToModelOrderItem converts a pb.OrderItem into a model.OrderItem
func ProtoOrderItemToModelOrderItem(item *pb.OrderItem) *OrderItem {
	if item == nil {
		return nil
	}
	return &OrderItem{
		ProductCode: item.GetProductCode(),
		Name:        item.GetName(),
		Price:       item.GetPrice(),
		Quantity:    item.GetQuantity(),
	}
}

// ModelOrderItemToProtoOrderItem converts a model.OrderItem into a pb.OrderItem
func ModelOrderItemToProtoOrderItem(item *OrderItem) *pb.OrderItem {
	if item == nil {
		return nil
	}
	return &pb.OrderItem{
		ProductCode: item.ProductCode,
		Name:        item.Name,
		Price:       item.Price,
		Quantity:    item.Quantity,
	}
}

// ProtoOrderItemsListToModelOrderItemsList converts a []*pb.OrderItem into a []*model.OrderItem
func ProtoOrderItemsListToModelOrderItemsList(items []*pb.OrderItem) []*OrderItem {
	if items == nil {
		return nil
	}

	var modelItems []*OrderItem

	for _, item := range items {
		modelItems = append(modelItems, ProtoOrderItemToModelOrderItem(item))
	}
	return modelItems
}

// ModelOrderItemsListToProtoOrderItemsList converts a []*model.OrderItem into a []*pb.OrderItem
func ModelOrderItemsListToProtoOrderItemsList(items []*OrderItem) []*pb.OrderItem {
	if items == nil {
		return nil
	}

	var pbItems []*pb.OrderItem

	for _, item := range items {
		pbItems = append(pbItems, ModelOrderItemToProtoOrderItem(item))
	}
	return pbItems
}
