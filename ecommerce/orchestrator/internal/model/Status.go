package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"

type Status string

const (
	StatusUnspecified Status = "unspecified"
	StatusProcessing  Status = "processing"
	StatusShipped     Status = "shipped"
	StatusDelivered   Status = "delivered"
	StatusCanceled    Status = "canceled"
)

var StatusMapStrToStatus = map[string]Status{
	"unspecified": StatusUnspecified,
	"processing":  StatusProcessing,
	"shipped":     StatusShipped,
	"delivered":   StatusDelivered,
	"canceled":    StatusCanceled,
}

var StatusMapStatusToStr = map[Status]string{
	StatusUnspecified: "unspecified",
	StatusProcessing:  "processing",
	StatusShipped:     "shipped",
	StatusDelivered:   "delivered",
	StatusCanceled:    "canceled",
}

// ProtoStatusToModelStatus converts a pb.Status into a model.Status
func ProtoStatusToModelStatus(ps pb.Status) Status {
	switch ps {
	case pb.Status_PROCESSING:
		return StatusProcessing
	case pb.Status_SHIPPED:
		return StatusShipped
	case pb.Status_DELIVERED:
		return StatusDelivered
	case pb.Status_CANCELED:
		return StatusCanceled
	default:
		return StatusUnspecified
	}
}

// ModelStatusToProtoStatus converts a model.Status into a pb.Status
func ModelStatusToProtoStatus(ps Status) pb.Status {
	switch ps {
	case StatusProcessing:
		return pb.Status_PROCESSING
	case StatusShipped:
		return pb.Status_SHIPPED
	case StatusDelivered:
		return pb.Status_DELIVERED
	case StatusCanceled:
		return pb.Status_CANCELED
	default:
		return pb.Status_UNSPECIFIED
	}
}

func (r Status) String() string {
	return string(r)
}