package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"

type Status struct {
	ID   int    `gorm:"column:id;type:int;primaryKey"`
	Name string `gorm:"column:name;type:varchar(20);uniqueIndex;not null"`
}

const (
	StatusUnspecified = 0
	StatusProcessing = 1
	StatusShipped = 2
	StatusDelivered = 3
	StatusCanceled = 4
)

// ModelStatusToProtoStatus converts an int status into pb.Status
func ModelStatusToProtoStatus(status int) pb.Status {
	switch status {
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

// ProtoStatusToModelStatus converts a pb.Status into an int status
func ProtoStatusToModelStatus(status pb.Status) int {
	switch status {
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

