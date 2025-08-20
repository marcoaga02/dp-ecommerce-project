package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/order-service/internal/model"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/order-service/internal/repository"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupDockerDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf(
		"root:%s@tcp(%s:%s)/%s",
		getEnvOrFail(t, "DB_PASSWORD"),
		getEnvOrFail(t, "DB_HOST"),
		getEnvOrFail(t, "DB_PORT"),
		getEnvOrFail(t, "DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to Docker DB: %v", err)
	}

	return db
}

func getEnvOrFail(t *testing.T, key string) string {
	val := os.Getenv(key)
	if val == "" {
		t.Fatalf("Environment variable %s not set", key)
	}
	return val
}

func cleanDB(t *testing.T, db *gorm.DB) {
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		t.Fatalf("Failed to disable foreign key checks: %v", err)
	}

	if err := db.Exec("TRUNCATE TABLE order_items").Error; err != nil {
		t.Fatalf("Failed to truncate order_items: %v", err)
	}

	if err := db.Exec("TRUNCATE TABLE orders").Error; err != nil {
		t.Fatalf("failed to truncate orders: %v", err)
	}

	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		t.Fatalf("Failed to enable foreign key checks: %v", err)
	}
}

func createDefaultOrders(t *testing.T, db *gorm.DB, repo *repository.GormOrderRepository) {
	orders := []model.Order{
		{
			Username: "user1",
			StatusID: model.StatusProcessing,
			Items: []model.OrderItem{
				{
					ProductCode: "code1",
					ProductName: "large hoodie",
					UnitPrice:   19.99,
					Quantity:    3,
				},
				{
					ProductCode: "code2",
					ProductName: "small jeans",
					UnitPrice:   25.99,
					Quantity:    7,
				},
			},
		},
		{
			Username: "user2",
			StatusID: model.StatusShipped,
			Items: []model.OrderItem{
				{
					ProductCode: "code3",
					ProductName: "large jeans",
					UnitPrice:   27.99,
					Quantity:    2,
				},
			},
		},
		{
			Username: "user1",
			StatusID: model.StatusDelivered,
			Items: []model.OrderItem{
				{
					ProductCode: "code1",
					ProductName: "large hoodie",
					UnitPrice:   19.99,
					Quantity:    2,
				},
				{
					ProductCode: "code3",
					ProductName: "large jeans",
					UnitPrice:   35.99,
					Quantity:    8,
				},
			},
		},
		{
			Username: "user3",
			StatusID: model.StatusCanceled,
			Items: []model.OrderItem{
				{
					ProductCode: "code1",
					ProductName: "large hoodie",
					UnitPrice:   19.99,
					Quantity:    6,
				},
			},
		},
	}

	for _, o := range orders {
		if err := db.Create(&o).Error; err != nil {
			t.Fatalf("Failed to insert order %+v: %v", o, err)
		}
	}
}

func createTwoDefaultOrders(t *testing.T, db *gorm.DB, repo *repository.GormOrderRepository) {
	orders := []model.Order{
		{
			Username: "user1",
			StatusID: model.StatusProcessing,
			Items: []model.OrderItem{
				{
					ProductCode: "code1",
					ProductName: "large hoodie",
					UnitPrice:   19.99,
					Quantity:    3,
				},
				{
					ProductCode: "code2",
					ProductName: "small jeans",
					UnitPrice:   25.99,
					Quantity:    7,
				},
			},
		},
		{
			Username: "user2",
			StatusID: model.StatusShipped,
			Items: []model.OrderItem{
				{
					ProductCode: "code3",
					ProductName: "large jeans",
					UnitPrice:   27.99,
					Quantity:    2,
				},
			},
		},
	}

	for _, o := range orders {
		if err := db.Create(&o).Error; err != nil {
			t.Fatalf("Failed to insert order %+v: %v", o, err)
		}
	}
}

func setupTestOrderRepo(t *testing.T) (*repository.GormOrderRepository, *gorm.DB) {
	db := setupDockerDB(t)
	cleanDB(t, db)
	orderRepo := repository.NewGormOrderRepository(db, logger.NewStdLogger(logger.Info, "gorm-order-repo-test"))
	createDefaultOrders(t, db, orderRepo)
	return orderRepo, db
}

func setupTestOrderRepoWithTwoDefaultOrders(t *testing.T) (*repository.GormOrderRepository, *gorm.DB) {
	db := setupDockerDB(t)
	cleanDB(t, db)
	orderRepo := repository.NewGormOrderRepository(db, logger.NewStdLogger(logger.Info, "gorm-order-repo-test"))
	createTwoDefaultOrders(t, db, orderRepo)
	return orderRepo, db
}

func setupTestEmptyOrderRepo(t *testing.T) (*repository.GormOrderRepository, *gorm.DB) {
	db := setupDockerDB(t)
	cleanDB(t, db)
	orderRepo := repository.NewGormOrderRepository(db, logger.NewStdLogger(logger.Info, "gorm-order-repo-test"))
	return orderRepo, db
}

func TestCreateOrderSuccessfully(t *testing.T) {
	orderRepo, db := setupTestOrderRepo(t)

	modelItems := []*model.OrderItem{
		{
			ProductCode: "code4",
			ProductName: "large t-shirt",
			UnitPrice:   14.99,
			Quantity:    3,
		},
		{
			ProductCode: "code1",
			ProductName: "large hoodie",
			UnitPrice:   19.99,
			Quantity:    7,
		},
	}

	succ, orderId, err := orderRepo.CreateOrder("userNew", model.ModelOrderItemListToProtoOrderItemList(modelItems))
	if err != nil {
		t.Fatalf("Unexpected error during the creation of an order: %v", err)
	}
	if !succ {
		t.Fatalf("Unsexpected unsuccessful order creation")
	}

	if orderId != 5 {
		t.Fatalf("Expected order id '5', but got '%d'", orderId)
	}

	var order model.Order
	err = db.Preload("Items").Where("id = ?", 5).First(&order).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving order with id 5: %v", err)
	}
	if len(order.Items) != 2 {
		t.Fatalf("Expected 2 items in the order, but got %d", len(order.Items))
	}
	if order.StatusID != model.StatusProcessing {
		t.Fatalf("Expected status with id 1 (processing), got %d", order.StatusID)
	}
	if order.Username != "userNew" {
		t.Fatalf("Expected username 'userNew', got '%s'", order.Username)
	}

	var foundCode1 bool = false
	var foundCode4 bool = false

	for _, item := range order.Items {
		if item.ProductCode == "code1" {
			foundCode1 = true
			if item.ProductName != "large hoodie" {
				t.Fatalf("Expected name 'large hoodie', got '%s'", item.ProductName)
			}
			if item.UnitPrice != 19.99 {
				t.Fatalf("Expected price '19.99', got '%f'", item.UnitPrice)
			}
			if item.Quantity != 7 {
				t.Fatalf("Expected quantity '7', got '%d'", item.Quantity)
			}
		}
		if item.ProductCode == "code4" {
			foundCode4 = true
			if item.ProductName != "large t-shirt" {
				t.Fatalf("Expected name 'large t-shirt', got '%s'", item.ProductName)
			}
			if item.UnitPrice != 14.99 {
				t.Fatalf("Expected price '14.99', got '%f'", item.UnitPrice)
			}
			if item.Quantity != 3 {
				t.Fatalf("Expected quantity '3', got '%d'", item.Quantity)
			}
		}
	}

	if !foundCode1 || !foundCode4 {
		t.Fatalf("Expected both items with product code 'code1' and 'code4' to be found, but that's not the case")
	}
}

func TestGetOrderWithExistingId(t *testing.T) {
	orderRepo, _ := setupTestOrderRepo(t)

	succ, order, err := orderRepo.GetOrder(3)
	if err != nil {
		t.Fatalf("Unexpected error retrieving order with id 3: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful retrieval of order with id 3")
	}
	if order == nil {
		t.Fatalf("Unexpected response with a nil order")
	}
	if order.Username != "user1" {
		t.Fatalf("Expected username 'user1', got '%s'", order.Username)
	}
	if model.ProtoStatusToModelStatus(order.Status) != model.StatusDelivered {
		t.Fatalf("Expected status 'DELIVERED', got '%s'", order.Status.String())
	}
	if order.Items == nil {
		t.Fatalf("Unexpected response with a nil order item list")
	}
	if len(order.Items) != 2 {
		t.Fatalf("Expected items list of length 2, got %d", len(order.Items))
	}

	var foundCode1 bool = false
	var foundCode3 bool = false

	for _, item := range order.Items {
		if item.ProductCode == "code1" {
			foundCode1 = true
			if item.Name != "large hoodie" {
				t.Fatalf("Expected name 'large hoodie', got '%s'", item.Name)
			}
			if item.Price != 19.99 {
				t.Fatalf("Expected price 19.99, got %f", item.Price)
			}
			if item.Quantity != 2 {
				t.Fatalf("Expected quantity 2, got %d", item.Quantity)
			}
		}
		if item.ProductCode == "code3" {
			foundCode3 = true
			if item.Name != "large jeans" {
				t.Fatalf("Expected name 'large jeans', got '%s'", item.Name)
			}
			if item.Price != 35.99 {
				t.Fatalf("Expected price 35.99, got %f", item.Price)
			}
			if item.Quantity != 8 {
				t.Fatalf("Expected quantity 8, got %d", item.Quantity)
			}
		}
	}

	if !foundCode1 || !foundCode3 {
		t.Fatalf("Expected both items with product code 'code1' and 'code3' to be found, but that's not the case")
	}
}

func TestGetOrderWithWrongId(t *testing.T) {
	orderRepo, _ := setupTestOrderRepo(t)

	succ, order, err := orderRepo.GetOrder(5)
	if err != nil {
		t.Fatalf("Unexpected error retrieving order using wrong id: %v", err)
	}
	if succ {
		t.Fatalf("Unexpected success retrieving order with wrong id")
	}
	if order != nil {
		t.Fatalf("Expected nil order, got a non-nil one")
	}
}

func TestUpdateOrderStatusUsingCorrectId(t *testing.T) {
	orderRepo, db := setupTestOrderRepo(t)

	succ, err := orderRepo.UpdateOrderStatus(2, model.StatusDelivered)
	if err != nil {
		t.Fatalf("Unexpected error while updating status of order with ID 2: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful status update for order with ID 2")
	}

	var order model.Order
	err = db.Where("id = ?", 2).First(&order).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving order with ID 2")
	}
	if order.StatusID != model.StatusDelivered {
		t.Fatalf("Expected status 3 (delivered), got %d", order.StatusID)
	}
}

func TestUpdateOrderStatusUsingNonExistingOrderId(t *testing.T) {
	orderRepo, _ := setupTestOrderRepo(t)

	succ, err := orderRepo.UpdateOrderStatus(5, model.StatusShipped)
	if err != nil {
		t.Fatalf("Unexpected error while trying to update status of a non-existing order: %v", err)
	}
	if succ {
		t.Fatalf("Unexpected success while trying to update status of a non-existing order")
	}
}

func TestListOrdersByUsernameUsingAnExistingUsername(t *testing.T) {
	orderRepo, _ := setupTestOrderRepo(t)

	succ, orders, err := orderRepo.ListOrdersByUsername("user1")
	if err != nil {
		t.Fatalf("Unexpected error retrieving orders for user 'user1': %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful retrieval of orders for user 'user1'")
	}
	if orders == nil || len(orders) == 0 {
		t.Fatalf("Unexpected retrieval of an empty list of orders; it should contains 2 elements")
	}
	if len(orders) != 2 {
		t.Fatalf("Expected list of 2 orders, got a list of %d orders", len(orders))
	}

	var foundOrderId1 bool = false
	var foundOrderId3 bool = false

	for _, order := range orders {
		if order.OrderId == 1 {
			foundOrderId1 = true
			if order.Username != "user1" {
				t.Fatalf("Expected to find username 'user1', got '%s'", order.Username)
			}
			if model.ProtoStatusToModelStatus(order.Status) != model.StatusProcessing {
				t.Fatalf("Expected status 'PROCESSING', got '%s'", order.Status.String())
			}
			if order.Items == nil {
				t.Fatalf("Expected list of 2 items, got nil list")
			}
			if len(order.Items) != 2 {
				t.Fatalf("Expected list of 2 items, got '%d' items", len(order.Items))
			}

			var foundCode1 bool = false
			var foundCode2 bool = false

			for _, item := range order.Items {
				if item.ProductCode == "code1" {
					foundCode1 = true
					if item.Name != "large hoodie" {
						t.Fatalf("Expected name 'large hoodie', got '%s'", item.Name)
					}
					if item.Price != 19.99 {
						t.Fatalf("Expected price 19.99, got %f", item.Price)
					}
					if item.Quantity != 3 {
						t.Fatalf("Expected quantity 3, got %d", item.Quantity)
					}
				}
				if item.ProductCode == "code2" {
					foundCode2 = true
					if item.Name != "small jeans" {
						t.Fatalf("Expected name 'small jeans', got '%s'", item.Name)
					}
					if item.Price != 25.99 {
						t.Fatalf("Expected price 25.99, got %f", item.Price)
					}
					if item.Quantity != 7 {
						t.Fatalf("Expected quantity 7, got %d", item.Quantity)
					}
				}
			}
			if !foundCode1 || !foundCode2 {
				t.Fatalf("Expected list of items code1 and code2, but got a different list")
			}
		}
		if order.OrderId == 3 {
			foundOrderId3 = true
			if order.Username != "user1" {
				t.Fatalf("Expected to find username 'user1', got '%s'", order.Username)
			}
			if model.ProtoStatusToModelStatus(order.Status) != model.StatusDelivered {
				t.Fatalf("Expected status 'DELIVERED', got '%s'", order.Status.String())
			}
			if order.Items == nil {
				t.Fatalf("Expected list of 2 items, got nil list")
			}
			if len(order.Items) != 2 {
				t.Fatalf("Expected list of 2 items, got '%d' items", len(order.Items))
			}

			var foundCode1 bool = false
			var foundCode3 bool = false

			for _, item := range order.Items {
				if item.ProductCode == "code1" {
					foundCode1 = true
					if item.Name != "large hoodie" {
						t.Fatalf("Expected name 'large hoodie', got '%s'", item.Name)
					}
					if item.Price != 19.99 {
						t.Fatalf("Expected price 19.99, got %f", item.Price)
					}
					if item.Quantity != 2 {
						t.Fatalf("Expected quantity 2, got %d", item.Quantity)
					}
				}
				if item.ProductCode == "code3" {
					foundCode3 = true
					if item.Name != "large jeans" {
						t.Fatalf("Expected name 'large jeans', got '%s'", item.Name)
					}
					if item.Price != 35.99 {
						t.Fatalf("Expected price 35.99, got %f", item.Price)
					}
					if item.Quantity != 8 {
						t.Fatalf("Expected quantity 8, got %d", item.Quantity)
					}
				}
			}
			if !foundCode1 || !foundCode3 {
				t.Fatalf("Expected list of items code1 and code3, but got a different list")
			}
		}
	}
	if !foundOrderId1 || !foundOrderId3 {
		t.Fatalf("Expected list of orders with ID's 1 and 3, but retrieved a different list")
	}
}

func TestListOrdersByUsernameUsingANonExistingUsername(t *testing.T) {
	orderRepo, _ := setupTestOrderRepo(t)

	succ, orders, err := orderRepo.ListOrdersByUsername("userUnk")
	if err != nil {
		t.Fatalf("Unexpected error retrieving orders for a non-existing user: %v", err)
	}
	if succ {
		t.Fatalf("Unexpeced success retrieving orders for a non-existing user")
	}
	if orders != nil {
		t.Fatalf("Unexpected non-nil list of orders")
	}
}

func TestListAllOrdersWithTwoOrders(t *testing.T) {
	orderRepo, _ := setupTestOrderRepoWithTwoDefaultOrders(t)

	succ, orders, err := orderRepo.ListAllOrders()
	if err != nil {
		t.Fatalf("Unexpected error retriving list of all orders: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful retrieval of the list of orders")
	}
	if orders == nil || len(orders) == 0 {
		t.Fatalf("Expected list of two orders, got an empty one")
	}

	var foundOrderId1 bool = false
	var foundOrderId2 bool = false

	for _, order := range orders {
		if order.OrderId == 1 {
			foundOrderId1 = true
			if order.Items == nil {
				t.Fatalf("Expected list of 2 items, got a nil list")
			}
			if len(order.Items) != 2 {
				t.Fatalf("Expected list of 2 items, got a liust of %d items", len(order.Items))
			}

			var foundCode1 bool = false
			var foundCode2 bool = false
			for _, item := range order.Items {
				if item.ProductCode == "code1" {
					foundCode1 = true
					if item.Name != "large hoodie" {
						t.Fatalf("Expected name 'large hoodie', got %s", item.Name)
					}
					if item.Price != 19.99 {
						t.Fatalf("Expected price 19.99, got %f", item.Price)
					}
					if item.Quantity != 3 {
						t.Fatalf("Expected quantity 3, got %d", item.Quantity)
					}
				}
				if item.ProductCode == "code2" {
					foundCode2 = true
					if item.Name != "small jeans" {
						t.Fatalf("Expected name 'small jeans', got %s", item.Name)
					}
					if item.Price != 25.99 {
						t.Fatalf("Expected price 25.99, got %f", item.Price)
					}
					if item.Quantity != 7 {
						t.Fatalf("Expected quantity 7, got %d", item.Quantity)
					}
				}
			}

			if !foundCode1 || !foundCode2 {
				t.Fatalf("Expected list containing products with codes 'code1' and 'code2', got a different list")
			}
		}
		if order.OrderId == 2 {
			foundOrderId2 = true
			if order.Items == nil {
				t.Fatalf("Expected list of 1 item, got a nil list")
			}
			if len(order.Items) != 1 {
				t.Fatalf("Expected list of 1 item, got a liust of %d items", len(order.Items))
			}

			var foundCode3 bool = false
			item := order.Items[0]
			if item.ProductCode == "code3" {
				foundCode3 = true
				if item.Name != "large jeans" {
					t.Fatalf("Expected name 'large jeans', got %s", item.Name)
				}
				if item.Price != 27.99 {
					t.Fatalf("Expected price 27.99, got %f", item.Price)
				}
				if item.Quantity != 2 {
					t.Fatalf("Expected quantity 2, got %d", item.Quantity)
				}
			}

			if !foundCode3 {
				t.Fatalf("Expected list containing product with code 'code3', got a different list")
			}
		}
	}

	if !foundOrderId1 || !foundOrderId2 {
		t.Fatalf("Expected list containing orders with IDs 1 and 2, got a different list")
	}
}

func TestListAllOrdersWithNoOrders(t *testing.T) {
	orderRepo, _ := setupTestEmptyOrderRepo(t)

	succ, orders, err := orderRepo.ListAllOrders()
	if err != nil {
		t.Fatalf("Unexpected error retrieving the empty list of orders")
	}
	if succ {
		t.Fatalf("Unexpected success retrieving the empty list of orders")
	}
	if orders != nil {
		t.Fatalf("Unexpected non-nil list of orders")
	}
}