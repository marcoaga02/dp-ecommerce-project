package test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/cart-service/internal/model"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/cart-service/internal/repository"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
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
	if err := db.Exec("TRUNCATE TABLE cart_items").Error; err != nil {
		t.Fatalf("Failed to truncate cart_items table: %v", err)
	}
}

func createDefaultCartItems(t *testing.T, db *gorm.DB, repo *repository.GormCartRepository) {
	items := []model.CartItem{
		{Username: "user1", Code: "code1", Quantity: 10},
		{Username: "user1", Code: "code2", Quantity: 20},
		{Username: "user2", Code: "code2", Quantity: 30},
		{Username: "user2", Code: "code3", Quantity: 5},
	}

	if err := db.Create(&items).Error; err != nil {
		t.Fatalf("Failed to create default cart items with GORM: %v", err)
	}
}

func setupTestCartRepo(t *testing.T) (*repository.GormCartRepository, *gorm.DB) {
	db := setupDockerDB(t)
	cleanDB(t, db)
	cartRepo := repository.NewGormCartRepository(db, logger.NewStdLogger(logger.Info, "gorm-cart-repo-test"))
	createDefaultCartItems(t, db, cartRepo)
	return cartRepo, db
}

func setupTestEmptyCartRepo(t *testing.T) (*repository.GormCartRepository, *gorm.DB) {
	db := setupDockerDB(t)
	cleanDB(t, db)
	cartRepo := repository.NewGormCartRepository(db, logger.NewStdLogger(logger.Info, "gorm-cart-repo-test"))
	return cartRepo, db
}

func TestAddItemNotExistingWithNewCodeWithCorrectFields(t *testing.T) {
	cartRepo, db := setupTestCartRepo(t)

	succ, err := cartRepo.AddItem("user1", "code4", 10)
	if err != nil {
		t.Fatalf("Unexpected error adding a new product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected failure addin a new product")
	}

	var item model.CartItem
	err = db.Where("username = ? AND code = ?", "user1", "code4").First(&item).Error
	if err != nil {
		t.Fatalf("Cart item not found in database after successful item creation: %v", err)
	}
	if item.Username != "user1" {
		t.Fatalf("Expected username 'user1', got '%s'", item.Username)
	}
	if item.Code != "code4" {
		t.Fatalf("Expected code 'code4', got '%s'", item.Code)
	}
	if item.Quantity != 10 {
		t.Fatalf("Expected quantity 10, got %d", item.Quantity)
	}
}

func TestAddItemNotExistingWithNewUsernameWithCorrectFields(t *testing.T) {
	cartRepo, db := setupTestCartRepo(t)

	succ, err := cartRepo.AddItem("user3", "code1", 15)
	if err != nil {
		t.Fatalf("Unexpected error adding a new product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected failure addin a new product")
	}

	var item model.CartItem
	err = db.Where("username = ? AND code = ?", "user3", "code1").First(&item).Error
	if err != nil {
		t.Fatalf("Cart item not found in database after successful item creation: %v", err)
	}
	if item.Username != "user3" {
		t.Fatalf("Expected username 'user3', got '%s'", item.Username)
	}
	if item.Code != "code1" {
		t.Fatalf("Expected code 'code1', got '%s'", item.Code)
	}
	if item.Quantity != 15 {
		t.Fatalf("Expected quantity 15, got %d", item.Quantity)
	}
}

func TestAddItemNotExistingWithNewUsernameAndCodeWithCorrectFields(t *testing.T) {
	cartRepo, db := setupTestCartRepo(t)

	succ, err := cartRepo.AddItem("user3", "code4", 15)
	if err != nil {
		t.Fatalf("Unexpected error adding a new product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected failure addin a new product")
	}

	var item model.CartItem
	err = db.Where("username = ? AND code = ?", "user3", "code4").First(&item).Error
	if err != nil {
		t.Fatalf("Cart item not found in database after successful item creation: %v", err)
	}
	if item.Username != "user3" {
		t.Fatalf("Expected username 'user3', got '%s'", item.Username)
	}
	if item.Code != "code4" {
		t.Fatalf("Expected code 'code4', got '%s'", item.Code)
	}
	if item.Quantity != 15 {
		t.Fatalf("Expected quantity 15, got %d", item.Quantity)
	}
}

func TestAddItemAlreadyExistingWithCorrectFields(t *testing.T) {
	cartRepo, db := setupTestCartRepo(t)

	succ, err := cartRepo.AddItem("user1", "code1", 17)
	if err != nil {
		t.Fatalf("Unexpected error adding a new product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected failure addin a new product")
	}

	var item model.CartItem
	err = db.Where("username = ? AND code = ?", "user1", "code1").First(&item).Error
	if err != nil {
		t.Fatalf("Cart item not found in database after successful item creation: %v", err)
	}
	if item.Username != "user1" {
		t.Fatalf("Expected username 'user1', got '%s'", item.Username)
	}
	if item.Code != "code1" {
		t.Fatalf("Expected code 'code1', got '%s'", item.Code)
	}
	if item.Quantity != 27 {
		t.Fatalf("Expected quantity 27, got %d", item.Quantity)
	}
}

func TestRemoveItemForAnExistingItem(t *testing.T) {
	cartRepo, db := setupTestCartRepo(t)

	succ, err := cartRepo.RemoveItem("user2", "code2")
	if err != nil {
		t.Fatalf("Unexpected error removing an existing item: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected failure removing an existing item")
	}

	var item model.CartItem
	err = db.Where("username = ? AND code = ?", "user2", "code2").First(&item).Error
	if err == nil {
		t.Fatalf("Expected error for record not found in the database")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Unexpected error retrieving the removed user")
	}
}

func TestRemoveItemForANonExistentItemButExistingUser(t *testing.T) {
	cartRepo, _ := setupTestCartRepo(t)

	succ, err := cartRepo.RemoveItem("user1", "code4")
	if err != nil {
		t.Fatalf("Unexpected error removing a non existing item: %v", err)
	}
	if succ {
		t.Fatalf("Expected failure removing an existing item")
	}
}

func TestRemoveItemForANonExistentItemButExistingCode(t *testing.T) {
	cartRepo, _ := setupTestCartRepo(t)

	succ, err := cartRepo.RemoveItem("user3", "code1")
	if err != nil {
		t.Fatalf("Unexpected error removing a non existing item: %v", err)
	}
	if succ {
		t.Fatalf("Expected failure removing an existing item")
	}
}

func TestRemoveItemForANonExistentItem(t *testing.T) {
	cartRepo, _ := setupTestCartRepo(t)

	succ, err := cartRepo.RemoveItem("user3", "code4")
	if err != nil {
		t.Fatalf("Unexpected error removing a non existing item: %v", err)
	}
	if succ {
		t.Fatalf("Expected failure removing an existing item")
	}
}

func TestUpdateItemQuantityWithAPositiveNumberForAnExistingItem(t *testing.T) {
	cartRepo, db := setupTestCartRepo(t)

	succ, err := cartRepo.UpdateItemQuantity("user2", "code2", 17)
	if err != nil {
		t.Fatalf("Unexpected error updating quantity of an existing item: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected error updating quantity of an existing item")
	}

	var item model.CartItem
	err = db.Where("username = ? AND code = ?", "user2", "code2").First(&item).Error
	if err != nil {
		t.Fatalf("Unxpected error retrieving item in the DB")
	}

	if item.Quantity != 17 {
		t.Fatalf("Expecting quantity 17, but got %d", item.Quantity)
	}
}

func TestUpdateItemQuantityWithAZeroQuantityForAnExistingItem(t *testing.T) {
	cartRepo, db := setupTestCartRepo(t)

	succ, err := cartRepo.UpdateItemQuantity("user2", "code2", 0)
	if err != nil {
		t.Fatalf("Unexpected error updating quantity of an existing item: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected failure updating quantity of an existing item")
	}

	var item model.CartItem
	err = db.Where("username = ? AND code = ?", "user2", "code2").First(&item).Error
	if err == nil {
		t.Fatalf("Expected error for record not found in the database")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Unexpected error retrieving the removed item")
	}
}

func TestUpdateItemQuantityWithAPositiveQuantityForANonExistingItemButExistingUser(t *testing.T) {
	cartRepo, _ := setupTestCartRepo(t)

	succ, err := cartRepo.UpdateItemQuantity("user1", "code4", 30)
	if err != nil {
		t.Fatalf("Unexpected error updating quantity of a non-existing item: %v", err)
	}
	if succ {
		t.Fatalf("Expected failure updating quantity of a non-existing item")
	}
}

func TestUpdateItemQuantityWithAPositiveQuantityForANonExistingItemButExistingCode(t *testing.T) {
	cartRepo, _ := setupTestCartRepo(t)

	succ, err := cartRepo.UpdateItemQuantity("user3", "code1", 30)
	if err != nil {
		t.Fatalf("Unexpected error updating quantity of a non-existing item: %v", err)
	}
	if succ {
		t.Fatalf("Expected failure updating quantity of a non-existing item")
	}
}

func TestListCartItemsForAnExistingUsername(t *testing.T) {
	cartRepo, _ := setupTestCartRepo(t)

	succ, items, err := cartRepo.ListCartItems("user1")
	if err != nil {
		t.Fatalf("Unexpected error retrieving items for existing user: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected failure retrieving items for existing user")
	}
	if items == nil {
		t.Fatalf("Unexpected nil item list for existing user")
	}
	if len(items) != 2 {
		t.Fatalf("Expected list of two items")
	}

	var foundCode1 bool = false
	var foundCode2 bool = false

	for _, item := range items {
		if item.ProductCode == "code1" {
			foundCode1 = true
			if item.Username != "user1" {
				t.Fatalf("Expected username 'user1', got %s", item.Username)
			}
			if item.Quantity != 10 {
				t.Fatalf("Expected quantity 10, got %d", item.Quantity)
			}
		}
		if item.ProductCode == "code2" {
			foundCode2 = true
			if item.Username != "user1" {
				t.Fatalf("Expected username 'user1', got %s", item.Username)
			}
			if item.Quantity != 20 {
				t.Fatalf("Expected quantity 20, got %d", item.Quantity)
			}
		}
	}

	if !foundCode1 || !foundCode2 {
		t.Fatalf("Expected to find 'code1' and 'code2', but that's not the case")
	}
}

func TestListCartItemsForNonExistingUser(t *testing.T) {
	cartRepo, _ := setupTestCartRepo(t)

	succ, items, err := cartRepo.ListCartItems("user3")
	if err != nil {
		t.Fatalf("Unexpected error retrieving items for a non-existing user: %v", err)
	}
	if succ {
		t.Fatalf("Expected failure retrieving items for a non-existing user")
	}
	if items != nil {
		t.Fatalf("Expected nil list, got a non-nil one")
	}
}

func TestClearCartForExistingUser(t *testing.T) {
	cartRepo, db := setupTestCartRepo(t)

	succ, err := cartRepo.ClearCart("user2")
	if err != nil {
		t.Fatalf("Unexpected error clearing cart of an existing user: %v", err)
	}
	if !succ {
		t.Fatalf("Unxpected failure clearing cart of an existing user")
	}

	var item model.CartItem
	err = db.Where("username = ?", "user2").First(&item).Error
	if err == nil {
		t.Fatalf("Expected error for record not found")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Expected no records, but found one or got unexpected error: %v", err)
	}
}

func TestClearCartForNonExistingUser(t *testing.T) {
	cartRepo, _ := setupTestCartRepo(t)

	succ, err := cartRepo.ClearCart("user3")
	if err != nil {
		t.Fatalf("Unexpected error clearing cart of a non-existing user: %v", err)
	}
	if succ {
		t.Fatalf("Unxpected success clearing cart of a mon-existing user")
	}
}

func TestRemoveProductFromAllCartsForAnExistingProduct(t *testing.T) {
	cartRepo, db := setupTestCartRepo(t)

	succ, err := cartRepo.RemoveProductFromAllCarts("code2")
	if err != nil {
		t.Fatalf("Unexpected error removing an existing product from all carts: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected failure removing an existing product from all carts")
	}

	var item model.CartItem
	err = db.Where("code = ?", "code2").First(&item).Error
	if err == nil {
		t.Fatalf("Expected error for record not found")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Expected no records, but found one or got unexpected error: %v", err)
	}
}

func TestRemoveProductFromAllCartsForNonExistingProduct(t *testing.T) {
	cartRepo, _ := setupTestCartRepo(t)

	succ, err := cartRepo.RemoveProductFromAllCarts("code4")
	if err != nil {
		t.Fatalf("Unexpected error removing a non-existing product from all carts: %v", err)
	}
	if succ {
		t.Fatalf("Unexpected success removing a non-existing product from all carts")
	}
}
