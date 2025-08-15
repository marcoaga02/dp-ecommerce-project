package model

import (
	"fmt"

	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"
)

type Product struct {
	Code        string  `gorm:"column:code;type:varchar(32);primaryKey"`
	Name        string  `gorm:"column:name;type:varchar(255);not null"`
	SizeID      int     `gorm:"column:size_id;not null"` // foreign key
	Size        Size    `gorm:"foreignKey:SizeID"`       // GORM relationship
	Color       string  `gorm:"column:color;type:varchar(30);not null"`
	Description string  `gorm:"column:description;type:varchar(255);not null"`
	Stock       uint32  `gorm:"column:stock;type:int unsigned;not null"`
	Price       float64 `gorm:"column:price;type:double(10,2) unsigned;not null"`
}

// ModelProductToProtoProduct converts a model.Product into a pb.Product
func ModelProductToProtoProduct(prod *Product) (*pb.Product, error) {
	if prod == nil {
		return nil, fmt.Errorf("Input argument is nil")
	}
	
	size := ModelSizeToProtoSize(prod.SizeID)
	if size == SizeUnspecified {
		return nil, fmt.Errorf("Invalid product size id '%d'", prod.SizeID)
	}

	return &pb.Product{
		Code:        prod.Code,
		Name:        prod.Name,
		Size:        size,
		Color:       prod.Color,
		Description: prod.Description,
		Stock:       prod.Stock,
		Price:       prod.Price,
	}, nil
}
