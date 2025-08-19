package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"

type Product struct {
	Code        string
	Name        string
	Size        Size
	Color       string
	Description string
	Stock       uint32
	Price       float64
}

// ProtoProductToModelProduct converts a pb.Product into a model.Product
func ProtoProductToModelProduct(prod *pb.Product) *Product {
	if prod == nil {
		return nil
	}
	return &Product{
		Code:        prod.GetCode(),
		Name:        prod.GetName(),
		Size:        ProtoSizeToModelSize(prod.GetSize()),
		Color:       prod.GetColor(),
		Description: prod.GetDescription(),
		Stock:       prod.GetStock(),
		Price:       prod.GetPrice(),
	}
}

// ModelProductToProtoProduct converts a model.Product into a pb.Product
func ModelProductToProtoProduct(prod *Product) *pb.Product {
	if prod == nil {
		return nil
	}
	return &pb.Product{
		Code:        prod.Code,
		Name:        prod.Name,
		Size:        ModelSizeToProtoSize(prod.Size),
		Color:       prod.Color,
		Description: prod.Description,
		Stock:       prod.Stock,
		Price:       prod.Price,
	}
}

// ProtoProductToModelProductsList converts a []*pb.Product into a []*model.Product
func ProtoProductsListToModelProductsList(prods []*pb.Product) []*Product {
	if prods == nil {
		return nil
	}

	var modelProds []*Product

	for _, prod := range prods {
		modelProds = append(modelProds, ProtoProductToModelProduct(prod))
	}
	return modelProds
}
