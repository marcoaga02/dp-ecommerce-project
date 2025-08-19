package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"

type Size string

const (
	SizeUnspecified Size = "unspecified"
	Size_XS         Size = "XS"
	Size_S          Size = "S"
	Size_M          Size = "M"
	Size_L          Size = "L"
	Size_XL         Size = "XL"
	Size_XXL        Size = "XXL"
)

var SizeMapStrToSize = map[string]Size{
	"XS":  Size_XS,
	"S":   Size_S,
	"M":   Size_M,
	"L":   Size_L,
	"XL":  Size_XL,
	"XXL": Size_XXL,
}

var SizeMapSizeToStr = map[Size]string{
	Size_XS:  "XS",
	Size_S:   "S",
	Size_M:   "M",
	Size_L:   "L",
	Size_XL:  "XL",
	Size_XXL: "XXL",
}

var AllSizes = []string{
	Size_XS.String(),
	Size_S.String(),
	Size_M.String(),
	Size_L.String(),
	Size_XL.String(),
	Size_XXL.String(),
}

// ProtoSizeToModelSize converts a pb.Size into a model.Size
func ProtoSizeToModelSize(ps pb.Size) Size {
	switch ps {
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

// ModelSizeToProtoSize converts a model.Size into a pb.Size
func ModelSizeToProtoSize(s Size) pb.Size {
	switch s {
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

// String returns the string associated to the model.Size
func (s Size) String() string {
	return string(s)
}
