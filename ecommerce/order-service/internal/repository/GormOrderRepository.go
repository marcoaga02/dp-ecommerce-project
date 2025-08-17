package repository

import (
	"errors"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/order-service/internal/model"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"
	"gorm.io/gorm"
)

const errOrderID int32 = -1

// GormOrderRepository implements the OrderServiceInterface using GORM as the ORM layer.
type GormOrderRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewGormOrderRepository creates an instance of GormOrderRepository
func NewGormOrderRepository(db *gorm.DB, logger logger.Logger) *GormOrderRepository {
	return &GormOrderRepository{
		db:     db,
		logger: logger,
	}
}

// CreateOrder creates a new order
func (r *GormOrderRepository) CreateOrder(username string, items []*pb.OrderItem) (bool, int32, error) {
	order := model.Order{
		Username: username,
		StatusID: model.StatusProcessing,
	}

	for _, pbItem := range items {
		orderItem := model.OrderItem{
			ProductCode: pbItem.ProductCode,
			ProductName: pbItem.Name,
			UnitPrice:   pbItem.Price,
			Quantity:    pbItem.Quantity,
		}
		order.Items = append(order.Items, orderItem)
	}

	if err := r.db.Create(&order).Error; err != nil {
		r.logger.Error("Error while creating new order for user '%s': %v", username, err)
		return false, errOrderID, err
	}

	r.logger.Info("Successful creation of order with %d items for user '%s'", len(order.Items), username)
	return true, order.ID, nil
}

// GetOrder retrieves the order with the given id
func (r *GormOrderRepository) GetOrder(orderId int32) (bool, *pb.Order, error) {
	var orderModel model.Order
	err := r.db.
		Preload("Items").
		Preload("Status").
		Where("id = ?", orderId).
		First(&orderModel).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("Order with ID '%d' not found", orderId)
			return false, nil, nil
		}
		r.logger.Error("Error retrieving order with ID '%d': %v", orderId, err)
		return false, nil, err
	}

	pbOrder, err := model.ModelOrderToProtoOrder(&orderModel)
	if err != nil {
		r.logger.Error("Failed to convert model order into protobuf order: %v", err)
		return false, nil, err
	}

	r.logger.Info("Order with ID '%d' retrieved successfully", orderId)
	return true, pbOrder, nil
}

// UpdateOrderStatus updates the status of the order with the given id
func (r *GormOrderRepository) UpdateOrderStatus(orderId int32, status pb.Status) (bool, error) {
	statusId := model.ProtoStatusToModelStatus(status)

	res := r.db.
		Model(&model.Order{}).
		Where("id = ?", orderId).
		Update("status_id", statusId)

	if res.Error != nil {
		r.logger.Error("Error while updating status for order with ID '%d': %v", orderId, res.Error)
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		r.logger.Warn("No order with ID '%d' found", orderId)
		return false, nil
	}

	r.logger.Info("Updated to '%d' status id of order with id '%d'", statusId, orderId)
	return true, nil
}

// ListOrdersByUsername retrieves the list of all orders of a given user
func (r *GormOrderRepository) ListOrdersByUsername(username string) (bool, []*pb.Order, error) {
	var ordersModelList []model.Order
	err := r.db.
		Preload("Items").
		Preload("Status").
		Where("username = ?", username).
		Order("id DESC").
		Find(&ordersModelList).Error

	if err != nil {
		r.logger.Error("Error while retrieving orders for user '%s': %v", username, err)
		return false, nil, err
	}

	if len(ordersModelList) == 0 {
		r.logger.Warn("No orders found for user '%s'", username)
		return false, nil, nil
	}

	pbOrders := make([]*pb.Order, 0, len(ordersModelList))
	for _, modelOrder := range ordersModelList {
		pbOrder, err := model.ModelOrderToProtoOrder(&modelOrder)
		if err != nil {
			r.logger.Error("Failed to convert model order into protobuf order: %v", err)
			return false, nil, err
		}
		pbOrders = append(pbOrders, pbOrder)
	}

	r.logger.Info("Retrieved all orders for user '%s'", username)
	return true, pbOrders, nil
}

// ListAllOrders retrieves the list of all orders
func (r *GormOrderRepository) ListAllOrders() (bool, []*pb.Order, error) {
	var ordersModelList []model.Order
	err := r.db.
		Preload("Items").
		Preload("Status").
		Order("id DESC").
		Find(&ordersModelList).Error

	if err != nil {
		r.logger.Error("Error while retrieving all orders: %v", err)
		return false, nil, err
	}

	if len(ordersModelList) == 0 {
		r.logger.Warn("No orders found")
		return false, nil, nil
	}

	pbOrders := make([]*pb.Order, 0, len(ordersModelList))
	for _, modelOrder := range ordersModelList {
		pbOrder, err := model.ModelOrderToProtoOrder(&modelOrder)
		if err != nil {
			r.logger.Error("Failed to convert model order into protobuf order: %v", err)
			return false, nil, err
		}
		pbOrders = append(pbOrders, pbOrder)
	}

	r.logger.Info("Retrieved all orders")
	return true, pbOrders, nil
}
