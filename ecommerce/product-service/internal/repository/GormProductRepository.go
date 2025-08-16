package repository

import (
	"errors"
	"fmt"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/product-service/internal/model"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/product"
	"gorm.io/gorm"
)

// GormProductRepository implements the ProductServiceInterface interface using GORM as the ORM layer.
type GormProductRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewGormProductRepository creates a new instance of GormProductRepository
func NewGormProductRepository(db *gorm.DB, logger logger.Logger) *GormProductRepository {
	return &GormProductRepository{
		db:     db,
		logger: logger,
	}
}

// CreateProduct creates a new product in the database
func (r *GormProductRepository) CreateProduct(prod *pb.Product) (bool, error) {
	var existingProd model.Product
	err := r.db.
		Where("code = ?", prod.Code).
		First(&existingProd).Error

	if err == nil {
		r.logger.Warn("Product with code '%s' already exists", prod.Code)
		return false, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		r.logger.Error("Error while checking existing products: %v", err)
		return false, err
	}

	sizeID := model.ProtoSizeToModelSize(prod.Size)
	if sizeID == model.SizeUnspecified {
		r.logger.Error("Invalid size for product with code '%s'", prod.Code)
		return false, fmt.Errorf("Invalid size spectified")
	}

	newProd := model.Product{
		Code:        prod.Code,
		Name:        prod.Name,
		SizeID:      sizeID,
		Color:       prod.Color,
		Description: prod.Description,
		Stock:       prod.Stock,
		Price:       prod.Price,
	}

	if err := r.db.Create(&newProd).Error; err != nil {
		r.logger.Error("Error creating product with code '%s': %v", prod.Code, err)
		return false, err
	}

	r.logger.Info("Product with code '%s' created successfully", prod.Code)
	return true, nil
}

// GetProduct retrieves the product in the database related to the given code
func (r *GormProductRepository) GetProduct(code string) (bool, *pb.Product, error) {
	prodModel, err := r.getProdByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("Product with code '%s' not found", code)
			return false, nil, nil
		}
		r.logger.Error("Error retrieving product with code '%s': %v", code, err)
		return false, nil, err
	}

	prod, err := model.ModelProductToProtoProduct(prodModel)
	if err != nil {
		r.logger.Error("Failed to convert model product into protobuf product: %v", err)
		return false, nil, err
	}

	r.logger.Info("Product with code '%s' retrieved successfully", code)
	return true, prod, nil
}

// UpdateProduct updates the product in the database related to the given code
func (r *GormProductRepository) UpdateProduct(code string, prod *pb.Product) (bool, error) {
	_, err := r.getProdByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("Product with code '%s' not found", code)
			return false, nil
		}
		r.logger.Error("Error retrieving product with code '%s': %v", code, err)
		return false, err
	}

	updates := map[string]interface{}{}

	updates["stock"] = prod.Stock

	if prod.Name != "" {
		updates["name"] = prod.Name
	}

	sizeID := model.ProtoSizeToModelSize(prod.Size)
	if sizeID != model.SizeUnspecified {
		updates["size_id"] = sizeID
	}
	
	if prod.Color != "" {
		updates["color"] = prod.Color
	}

	if prod.Description != "" {
		updates["description"] = prod.Description
	}

	if prod.Price >= 0 {
		updates["price"] = prod.Price
	}

	err = r.db.Model(&model.Product{}).
		Where("code = ?", code).
		Updates(updates).Error

	if err != nil {
		r.logger.Error("Failed to update product with code '%s': %v", code, err)
		return false, err
	}

	r.logger.Info("Product with code '%s' updated successfully", code)
	return true, nil
}

// DeleteProduct deletes the product in the database related to the given code
func (r *GormProductRepository) DeleteProduct(code string) (bool, error) {
	_, err := r.getProdByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("Product with code '%s' not found", code)
			return false, nil
		}
		r.logger.Error("Error retrieving product with code '%s': %v", code, err)
		return false, err
	}

	err = r.db.Where("code = ?", code).
		Delete(&model.Product{}).Error

	if err != nil {
		r.logger.Error("Error deleting product with code '%s': %v", code, err)
		return false, err
	}
	r.logger.Info("Product with code '%s' successfully deleted", code)
	return true, nil
}

// ListProducts retrieves the list of all products
func (r *GormProductRepository) ListProducts() (bool, []*pb.Product, error) {
	var prodModels []model.Product
	err := r.db.Preload("Size").Find(&prodModels).Error
	if err != nil {
		r.logger.Error("Error while retrieving all products: %v", err)
		return false, nil, err
	}
	if len(prodModels) == 0 {
		r.logger.Warn("No products found")
		return false, nil, nil
	}

	var products []*pb.Product
	for _, prodModel := range prodModels {
		prodPb, err := model.ModelProductToProtoProduct(&prodModel)
		if err != nil {
			r.logger.Warn("Skipping products with code '%s' due to invalid size mapping: %v", prodModel.Code, err)
			continue
		}
		products = append(products, prodPb)
	}

	r.logger.Info("Retrieved all products successfully")
	return true, products, nil
}

// getProdByCode returns the product from the database, given its code
func (r *GormProductRepository) getProdByCode(code string) (*model.Product, error) {
	var prod model.Product
	err := r.db.Preload("Size").
		Where("code = ?", code).
		First(&prod).Error

	if err != nil {
		return nil, err
	}

	return &prod, nil
}

func (r *GormProductRepository) CreateDefaultProducts() error {
	type DefaultProduct struct {
		Code        string
		Name        string
		Size        pb.Size
		Color       string
		Description string
		Stock       uint32
		Price       float64
	}
	defaultProducts := []DefaultProduct{
		{
			Code:        "TSHIRT-001",
			Name:        "Basic Cotton T-Shirt",
			Size:        pb.Size_M,
			Color:       "White",
			Description: "Comfortable 100% cotton t-shirt for everyday wear",
			Stock:       100,
			Price:       24.99,
		},
		{
			Code:        "TSHIRT-002",
			Name:        "Basic Cotton T-Shirt",
			Size:        pb.Size_XS,
			Color:       "Black",
			Description: "Comfortable 100% cotton t-shirt for everyday wear",
			Stock:       85,
			Price:       29.99,
		},
		{
			Code:        "TSHIRT-003",
			Name:        "Basic Cotton T-Shirt",
			Size:        pb.Size_XL,
			Color:       "Black",
			Description: "Comfortable 100% cotton t-shirt for everyday wear",
			Stock:       85,
			Price:       19.99,
		},
		{
			Code:        "JEANS-001",
			Name:        "Classic Denim Jeans",
			Size:        pb.Size_L,
			Color:       "Blue",
			Description: "Durable straight-cut denim jeans",
			Stock:       50,
			Price:       49.99,
		},
		{
			Code:        "JEANS-002",
			Name:        "Classic Denim Jeans",
			Size:        pb.Size_S,
			Color:       "Blue",
			Description: "Durable straight-cut denim jeans",
			Stock:       50,
			Price:       45.99,
		},
		{
			Code:        "HOODIE-001",
			Name:        "Comfortable Fleece Hoodie",
			Size:        pb.Size_S,
			Color:       "Gray",
			Description: "Hoodie with front pocket",
			Stock:       75,
			Price:       27.75,
		},
		{
			Code:        "HOODIE-002",
			Name:        "Comfortable Fleece Hoodie",
			Size:        pb.Size_XXL,
			Color:       "Gray",
			Description: "Hoodie with front pocket",
			Stock:       45,
			Price:       18.25,
		},
	}

	for _, prod := range defaultProducts {
		succ, err := r.CreateProduct(&pb.Product{
			Code:        prod.Code,
			Name:        prod.Name,
			Size:        prod.Size,
			Color:       prod.Color,
			Description: prod.Description,
			Stock:       prod.Stock,
			Price:       prod.Price,
		})
		if err != nil {
			r.logger.Error("Error creating default product with code'%s': %v", prod.Code, err)
			return err
		}
		if !succ {
			r.logger.Info("Default product '%s' created successfully", prod.Code)
		} else {
			r.logger.Warn("Default product with code '%s' not created: product with same code already exists")
		}
	}
	r.logger.Info("Default products initialization completed")
	return nil
}
