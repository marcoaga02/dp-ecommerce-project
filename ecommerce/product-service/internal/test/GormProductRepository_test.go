package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/product-service/internal/model"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/product-service/internal/repository"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"
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
	if err := db.Exec("TRUNCATE TABLE products").Error; err != nil {
		t.Fatalf("Failed to truncate products table: %v", err)
	}
}

func createDefaultProducts(t *testing.T, db *gorm.DB, repo *repository.GormProductRepository) {
	prods := []model.Product{
		{
			Code:        "code1",
			Name:        "large hoodie",
			SizeID:      model.Size_L,
			Color:       "red",
			Description: "large red hoodie",
			Stock:       25,
			Price:       24.99,
		},
		{
			Code:        "code2",
			Name:        "cotton t-shirt",
			SizeID:      model.Size_S,
			Color:       "white",
			Description: "white cotton t-shirt",
			Stock:       75,
			Price:       14.99,
		},
	}

	if err := db.Create(&prods).Error; err != nil {
		t.Fatalf("Failed to create default products with GORM: %v", err)
	}
}

func setupTestProdRepo(t *testing.T) (*repository.GormProductRepository, *gorm.DB) {
	db := setupDockerDB(t)
	cleanDB(t, db)
	prodRepo := repository.NewGormProductRepository(db, logger.NewStdLogger(logger.Info, "gorm-product-repo-test"))
	createDefaultProducts(t, db, prodRepo)
	return prodRepo, db
}

func setupTestEmptyProdRepo(t *testing.T) (*repository.GormProductRepository, *gorm.DB) {
	db := setupDockerDB(t)
	cleanDB(t, db)
	prodRepo := repository.NewGormProductRepository(db, logger.NewStdLogger(logger.Info, "gorm-prod-repo-test"))
	return prodRepo, db
}

func TestCreateProductSuccessful(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Code:        "codenew",
		Name:        "new product",
		Size:        pb.Size_XS,
		Color:       "yellow",
		Description: "New yellow product extra small",
		Stock:       15,
		Price:       12.75,
	}

	succ, err := prodRepo.CreateProduct(&prod)
	if err != nil {
		t.Fatalf("Unexpected error creating product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful creation of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "codenew").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the new product: %v", err)
	}
	if modelProd.Name != "new product" {
		t.Fatalf("Expected name 'new product', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_XS {
		t.Fatalf("Expected size 1 (XS), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "yellow" {
		t.Fatalf("Expected color 'yellow', got %s", modelProd.Color)
	}
	if modelProd.Description != "New yellow product extra small" {
		t.Fatalf("Expected description 'New yellow product extra small', got %s", modelProd.Description)
	}
	if modelProd.Stock != 15 {
		t.Fatalf("Expected stock 15, got %d", modelProd.Stock)
	}
	if modelProd.Price != 12.75 {
		t.Fatalf("Expected price 12.75, got %f", modelProd.Price)
	}
}

func TestCreateProductUsingExistingCode(t *testing.T) {
	prodRepo, _ := setupTestProdRepo(t)

	prod := pb.Product{
		Code:        "code2",
		Name:        "new product",
		Size:        pb.Size_XS,
		Color:       "yellow",
		Description: "New yellow product extra small",
		Stock:       15,
		Price:       12.75,
	}

	succ, err := prodRepo.CreateProduct(&prod)
	if err != nil {
		t.Fatalf("Unexpected error creating product with already existing code: %v", err)
	}
	if succ {
		t.Fatalf("Unexpected success creating product with already existing code:")
	}
}

func TestGetProductUsingExistingCode(t *testing.T) {
	prodRepo, _ := setupTestProdRepo(t)

	succ, prod, err := prodRepo.GetProduct("code2")
	if err != nil {
		t.Fatalf("Unexpected error retrieving product with code 'code2': %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful retrieval of product with code 'code2'")
	}
	if prod == nil {
		t.Fatalf("Unexpected nil product")
	}
	if prod.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", prod.Code)
	}
	if prod.Name != "cotton t-shirt" {
		t.Fatalf("Expected name 'cotton t-shirt', got '%s'", prod.Name)
	}
	if model.ProtoSizeToModelSize(prod.Size) != model.Size_S {
		t.Fatalf("Expected size 2 (S), got %d", model.ProtoSizeToModelSize(prod.Size))
	}
	if prod.Color != "white" {
		t.Fatalf("Expected color 'white', got %s", prod.Color)
	}
	if prod.Description != "white cotton t-shirt" {
		t.Fatalf("Expected description 'white cotton t-shirt', got %s", prod.Description)
	}
	if prod.Stock != 75 {
		t.Fatalf("Expected stock 75, got %d", prod.Stock)
	}
	if prod.Price != 14.99 {
		t.Fatalf("Expected price 14.99, got %f", prod.Price)
	}
}

func TestGetProductUsingNonExistingCode(t *testing.T) {
	prodRepo, _ := setupTestProdRepo(t)

	succ, prod, err := prodRepo.GetProduct("codeUnk")
	if err != nil {
		t.Fatalf("Unexpected error retrieving a product with wrong code: %v", err)
	}
	if succ {
		t.Fatalf("Unexpected successful retrieving of a product with wrong code")
	}
	if prod != nil {
		t.Fatalf("Expected nil product, got a non-nil one")
	}
}

func TestUpdateProductToUpdateOnlyName(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Name:  "New Name",
		Stock: 75,    // the stock must always be specified, since a zero stock is a valid value
		Price: 14.99, // the price must always be specified, since a zero price is a valid value
	}

	succ, err := prodRepo.UpdateProduct("code2", &prod)
	if err != nil {
		t.Fatalf("Unexpected error updating name of product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful update for the name of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "code2").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the product: %v", err)
	}
	if modelProd.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", modelProd.Code)
	}
	if modelProd.Name != "New Name" {
		t.Fatalf("Expected updated name 'New Name', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_S {
		t.Fatalf("Expected same size 2 (S), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "white" {
		t.Fatalf("Expected color 'white', got %s", modelProd.Color)
	}
	if modelProd.Description != "white cotton t-shirt" {
		t.Fatalf("Expected description 'white cotton t-shirt', got %s", modelProd.Description)
	}
	if modelProd.Stock != 75 {
		t.Fatalf("Expected stock 75, got %d", modelProd.Stock)
	}
	if modelProd.Price != 14.99 {
		t.Fatalf("Expected price 14.99, got %f", modelProd.Price)
	}
}

func TestUpdateProductToUpdateOnlySize(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Size:  pb.Size_XL,
		Stock: 75,    // the stock must always be specified, since a zero stock is a valid value
		Price: 14.99, // the price must always be specified, since a zero price is a valid value
	}

	succ, err := prodRepo.UpdateProduct("code2", &prod)
	if err != nil {
		t.Fatalf("Unexpected error updating size of product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful update for the size of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "code2").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the product: %v", err)
	}
	if modelProd.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", modelProd.Code)
	}
	if modelProd.Name != "cotton t-shirt" {
		t.Fatalf("Expected same name 'cotton t-shirt', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_XL {
		t.Fatalf("Expected new size 5 (XL), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "white" {
		t.Fatalf("Expected color 'white', got %s", modelProd.Color)
	}
	if modelProd.Description != "white cotton t-shirt" {
		t.Fatalf("Expected description 'white cotton t-shirt', got %s", modelProd.Description)
	}
	if modelProd.Stock != 75 {
		t.Fatalf("Expected stock 75, got %d", modelProd.Stock)
	}
	if modelProd.Price != 14.99 {
		t.Fatalf("Expected price 14.99, got %f", modelProd.Price)
	}
}

func TestUpdateProductToUpdateOnlyColor(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Color: "new color",
		Stock: 75,    // the stock must always be specified, since a zero stock is a valid value
		Price: 14.99, // the price must always be specified, since a zero price is a valid value
	}

	succ, err := prodRepo.UpdateProduct("code2", &prod)
	if err != nil {
		t.Fatalf("Unexpected error updating color of product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful update for the color of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "code2").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the product: %v", err)
	}
	if modelProd.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", modelProd.Code)
	}
	if modelProd.Name != "cotton t-shirt" {
		t.Fatalf("Expected same name 'cotton t-shirt', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_S {
		t.Fatalf("Expected same size 2 (S), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "new color" {
		t.Fatalf("Expected new color 'new color', got %s", modelProd.Color)
	}
	if modelProd.Description != "white cotton t-shirt" {
		t.Fatalf("Expected description 'white cotton t-shirt', got %s", modelProd.Description)
	}
	if modelProd.Stock != 75 {
		t.Fatalf("Expected stock 75, got %d", modelProd.Stock)
	}
	if modelProd.Price != 14.99 {
		t.Fatalf("Expected price 14.99, got %f", modelProd.Price)
	}
}

func TestUpdateProductToUpdateOnlyDescription(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Description: "new description",
		Stock:       75,    // the stock must always be specified, since a zero stock is a valid value
		Price:       14.99, // the price must always be specified, since a zero price is a valid value
	}

	succ, err := prodRepo.UpdateProduct("code2", &prod)
	if err != nil {
		t.Fatalf("Unexpected error updating description of product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful update for the description of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "code2").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the product: %v", err)
	}
	if modelProd.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", modelProd.Code)
	}
	if modelProd.Name != "cotton t-shirt" {
		t.Fatalf("Expected same name 'cotton t-shirt', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_S {
		t.Fatalf("Expected same size 2 (S), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "white" {
		t.Fatalf("Expected same color 'white', got %s", modelProd.Color)
	}
	if modelProd.Description != "new description" {
		t.Fatalf("Expected new description 'new description', got %s", modelProd.Description)
	}
	if modelProd.Stock != 75 {
		t.Fatalf("Expected stock 75, got %d", modelProd.Stock)
	}
	if modelProd.Price != 14.99 {
		t.Fatalf("Expected price 14.99, got %f", modelProd.Price)
	}
}

func TestUpdateProductToUpdateOnlyStockWithZeroValue(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Stock: 0,     // the stock must always be specified, since a zero stock is a valid value
		Price: 14.99, // the price must always be specified, since a zero price is a valid value
	}

	succ, err := prodRepo.UpdateProduct("code2", &prod)
	if err != nil {
		t.Fatalf("Unexpected error updating stock of product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful update for the stock of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "code2").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the product: %v", err)
	}
	if modelProd.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", modelProd.Code)
	}
	if modelProd.Name != "cotton t-shirt" {
		t.Fatalf("Expected same name 'cotton t-shirt', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_S {
		t.Fatalf("Expected same size 2 (S), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "white" {
		t.Fatalf("Expected same color 'white', got %s", modelProd.Color)
	}
	if modelProd.Description != "white cotton t-shirt" {
		t.Fatalf("Expected same description 'white cotton t-shirt', got %s", modelProd.Description)
	}
	if modelProd.Stock != 0 {
		t.Fatalf("Expected new stock 0, got %d", modelProd.Stock)
	}
	if modelProd.Price != 14.99 {
		t.Fatalf("Expected price 14.99, got %f", modelProd.Price)
	}
}

func TestUpdateProductToUpdateOnlyStockWithPositiveValue(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Stock: 150,   // the stock must always be specified, since a zero stock is a valid value
		Price: 14.99, // the price must always be specified, since a zero price is a valid value
	}

	succ, err := prodRepo.UpdateProduct("code2", &prod)
	if err != nil {
		t.Fatalf("Unexpected error updating stock of product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful update for the stock of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "code2").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the product: %v", err)
	}
	if modelProd.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", modelProd.Code)
	}
	if modelProd.Name != "cotton t-shirt" {
		t.Fatalf("Expected same name 'cotton t-shirt', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_S {
		t.Fatalf("Expected same size 2 (S), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "white" {
		t.Fatalf("Expected same color 'white', got %s", modelProd.Color)
	}
	if modelProd.Description != "white cotton t-shirt" {
		t.Fatalf("Expected same description 'white cotton t-shirt', got %s", modelProd.Description)
	}
	if modelProd.Stock != 150 {
		t.Fatalf("Expected new stock 150, got %d", modelProd.Stock)
	}
	if modelProd.Price != 14.99 {
		t.Fatalf("Expected price 14.99, got %f", modelProd.Price)
	}
}

func TestUpdateProductToUpdateOnlyPriceWithZeroValue(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Stock: 75, // the stock must always be specified, since a zero stock is a valid value
		Price: 0,  // the price must always be specified, since a zero price is a valid value
	}

	succ, err := prodRepo.UpdateProduct("code2", &prod)
	if err != nil {
		t.Fatalf("Unexpected error updating price of product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful update for the price of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "code2").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the product: %v", err)
	}
	if modelProd.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", modelProd.Code)
	}
	if modelProd.Name != "cotton t-shirt" {
		t.Fatalf("Expected same name 'cotton t-shirt', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_S {
		t.Fatalf("Expected same size 2 (S), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "white" {
		t.Fatalf("Expected same color 'white', got %s", modelProd.Color)
	}
	if modelProd.Description != "white cotton t-shirt" {
		t.Fatalf("Expected same description 'white cotton t-shirt', got %s", modelProd.Description)
	}
	if modelProd.Stock != 75 {
		t.Fatalf("Expected same stock 75, got %d", modelProd.Stock)
	}
	if modelProd.Price != 0 {
		t.Fatalf("Expected new price 0, got %f", modelProd.Price)
	}
}

func TestUpdateProductToUpdateOnlyPriceWithPositiveValue(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Stock: 75,    // the stock must always be specified, since a zero stock is a valid value
		Price: 72.25, // the price must always be specified, since a zero price is a valid value
	}

	succ, err := prodRepo.UpdateProduct("code2", &prod)
	if err != nil {
		t.Fatalf("Unexpected error updating price of product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful update for the price of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "code2").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the product: %v", err)
	}
	if modelProd.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", modelProd.Code)
	}
	if modelProd.Name != "cotton t-shirt" {
		t.Fatalf("Expected same name 'cotton t-shirt', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_S {
		t.Fatalf("Expected same size 2 (S), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "white" {
		t.Fatalf("Expected same color 'white', got %s", modelProd.Color)
	}
	if modelProd.Description != "white cotton t-shirt" {
		t.Fatalf("Expected same description 'white cotton t-shirt', got %s", modelProd.Description)
	}
	if modelProd.Stock != 75 {
		t.Fatalf("Expected same stock 75, got %d", modelProd.Stock)
	}
	if modelProd.Price != 72.25 {
		t.Fatalf("Expected new price 72.25, got %f", modelProd.Price)
	}
}

func TestUpdateProductWithoutAnyChangeRequired(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Stock: 75,    // the stock must always be specified, since a zero stock is a valid value
		Price: 14.99, // the price must always be specified, since a zero price is a valid value
	}

	succ, err := prodRepo.UpdateProduct("code2", &prod)
	if err != nil {
		t.Fatalf("Unexpected error updating no fields of product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful update for no fields of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "code2").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the product: %v", err)
	}
	if modelProd.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", modelProd.Code)
	}
	if modelProd.Name != "cotton t-shirt" {
		t.Fatalf("Expected same name 'cotton t-shirt', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_S {
		t.Fatalf("Expected same size 2 (S), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "white" {
		t.Fatalf("Expected same color 'white', got %s", modelProd.Color)
	}
	if modelProd.Description != "white cotton t-shirt" {
		t.Fatalf("Expected same description 'white cotton t-shirt', got %s", modelProd.Description)
	}
	if modelProd.Stock != 75 {
		t.Fatalf("Expected same stock 75, got %d", modelProd.Stock)
	}
	if modelProd.Price != 14.99 {
		t.Fatalf("Expected same price 14.99, got %f", modelProd.Price)
	}
}

func TestUpdateProductToUpdateAllParams(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	prod := pb.Product{
		Name:        "new name",
		Size:        model.Size_M,
		Color:       "new color",
		Description: "new description",
		Stock:       27,    // the stock must always be specified, since a zero stock is a valid value
		Price:       48.13, // the price must always be specified, since a zero price is a valid value
	}

	succ, err := prodRepo.UpdateProduct("code2", &prod)
	if err != nil {
		t.Fatalf("Unexpected error updating all fields of product: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful update for all fields of product")
	}

	var modelProd model.Product
	err = db.Where("code = ?", "code2").First(&modelProd).Error
	if err != nil {
		t.Fatalf("Unexpected error retrieving the product: %v", err)
	}
	if modelProd.Code != "code2" {
		t.Fatalf("Expected code 'code2', got %s", modelProd.Code)
	}
	if modelProd.Name != "new name" {
		t.Fatalf("Expected new name 'new name', got '%s'", modelProd.Name)
	}
	if modelProd.SizeID != model.Size_M {
		t.Fatalf("Expected new size 3 (M), got %d", modelProd.SizeID)
	}
	if modelProd.Color != "new color" {
		t.Fatalf("Expected new color 'new color', got %s", modelProd.Color)
	}
	if modelProd.Description != "new description" {
		t.Fatalf("Expected new description 'new description', got %s", modelProd.Description)
	}
	if modelProd.Stock != 27 {
		t.Fatalf("Expected new stock 27, got %d", modelProd.Stock)
	}
	if modelProd.Price != 48.13 {
		t.Fatalf("Expected new price 48.13, got %f", modelProd.Price)
	}
}

func TestDeleteProductUsingExistingCode(t *testing.T) {
	prodRepo, db := setupTestProdRepo(t)

	succ, err := prodRepo.DeleteProduct("code2")
	if err != nil {
		t.Fatalf("Unexpected error deleting product using correct code: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected usuccessful delete using correct code")
	}

	var prod model.Product
	err = db.Where("code = ?", "code2").First(&prod).Error
	if err == nil {
		t.Fatalf("Expected record not found error, got nil")
	}
}

func TestDeleteProductUsingNonExistingCode(t *testing.T) {
	prodRepo, _ := setupTestProdRepo(t)

	succ, err := prodRepo.DeleteProduct("codeUnk")
	if err != nil {
		t.Fatalf("Unexpected error deleting product using non-existing code: %v", err)
	}
	if succ {
		t.Fatalf("Unexpected successful delete using non-existing code")
	}
}

func TestListProductsWithNonEmptyProductList(t *testing.T) {
	prodRepo, _ := setupTestProdRepo(t)

	succ, prods, err := prodRepo.ListProducts()
	if err != nil {
		t.Fatalf("Unexpected error retrieving the list of products: %v", err)
	}
	if !succ {
		t.Fatalf("Unexpected unsuccessful retrieval of the list of products")
	}
	if prods == nil || len(prods) == 0 {
		t.Fatalf("Expected list of two products, got an empty one")
	}
	if len(prods) != 2 {
		t.Fatalf("Expected list of two products, got list of %d products", len(prods))
	}

	var foundCode1 bool = false
	var foundCode2 bool = false

	for _, prod := range prods {
		if prod.Code == "code1" {
			foundCode1 = true
			if prod.Name != "large hoodie" {
				t.Fatalf("Expected name 'large hoodie', got '%s'", prod.Name)
			}
			if model.ProtoSizeToModelSize(prod.Size) != model.Size_L {
				t.Fatalf("Expected size L, got size %s", prod.Size.String())
			}
			if prod.Color != "red" {
				t.Fatalf("Expected color 'red', got color '%s'", prod.Color)
			}
			if prod.Description != "large red hoodie" {
				t.Fatalf("Expected description 'large red hoodie', got '%s'", prod.Description)
			}
			if prod.Stock != 25 {
				t.Fatalf("Expected stock 25, got %d", prod.Stock)
			}
			if prod.Price != 24.99 {
				t.Fatalf("Expected price 24.99, got %f", prod.Price)
			}
		}
		if prod.Code == "code2" {
			foundCode2 = true
			if prod.Name != "cotton t-shirt" {
				t.Fatalf("Expected name 'cotton t-shirt', got '%s'", prod.Name)
			}
			if model.ProtoSizeToModelSize(prod.Size) != model.Size_S {
				t.Fatalf("Expected size S, got size %s", prod.Size.String())
			}
			if prod.Color != "white" {
				t.Fatalf("Expected color 'white', got color '%s'", prod.Color)
			}
			if prod.Description != "white cotton t-shirt" {
				t.Fatalf("Expected description 'white cotton t-shirt', got '%s'", prod.Description)
			}
			if prod.Stock != 75 {
				t.Fatalf("Expected stock 75, got %d", prod.Stock)
			}
			if prod.Price != 14.99 {
				t.Fatalf("Expected price 14.99, got %f", prod.Price)
			}
		}
	}

	if !foundCode1 || !foundCode2 {
		t.Fatalf("Expected list of two products with codes 'code1' and 'code2', got a different list")
	}
}

func TestListProductsWithEmptyProductList(t *testing.T) {
	prodRepo, _ := setupTestEmptyProdRepo(t)

	succ, prods, err := prodRepo.ListProducts()
	if err != nil {
		t.Fatalf("Unexpected error retrieving the empty list of products: %v", err)
	}
	if succ {
		t.Fatalf("Unexpected success retrieving the empty list of products")
	}
	if prods != nil {
		t.Fatalf("Expected nil list of products, got a non-nil list")
	}
}