package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"

type Size struct {
	ID   int    `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name;type:varchar(20);uniqueIndex;not null"`
}

// Constants representing sizes in the database
const (
	SizeUnspecified = 0
	Size_XS         = 1
	Size_S          = 2
	Size_M          = 3
	Size_L          = 4
	Size_XL         = 5
	Size_XXL        = 6
)

// ModelSizeToProtoSize converts an int size into a pb.Size
func ModelSizeToProtoSize(size int) pb.Size {
	switch size {
	case Size_XS:
		return pb.Size_XS
	case Size_S:
		return pb.Size_S
	case Size_M:
		return pb.Size_M
	case Size_L:
		return pb.Size_L
	case Size_XL:
		return pb.Size_XL
	case Size_XXL:
		return pb.Size_XXL
	default:
		return pb.Size_UNSPECIFIED
	}
}

// ProtoSizeToModelSize converts a pb.Size into an int size
func ProtoSizeToModelSize(size pb.Size) int {
	switch size {
	case pb.Size_XS:
		return Size_XS
	case pb.Size_S:
		return Size_S
	case pb.Size_M:
		return Size_M
	case pb.Size_L:
		return Size_L
	case pb.Size_XL:
		return Size_XL
	case pb.Size_XXL:
		return Size_XXL
	default:
		return SizeUnspecified
	}
}
