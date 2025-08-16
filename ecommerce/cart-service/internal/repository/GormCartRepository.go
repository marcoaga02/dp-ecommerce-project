package repository

import (
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/cart-service/internal/model"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/cart"
	"gorm.io/gorm"
)

// GormCartRepository implements the CartServiceInterface interface using GORM as the ORM layer.
type GormCartRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewGormCartRepository creates a new instance of GormCartRepository.
func NewGormCartRepository(db *gorm.DB, logger logger.Logger) *GormCartRepository {
	return &GormCartRepository{
		db:     db,
		logger: logger,
	}
}

// AddItem adds a product to the specified user's cart, increasing quantity if already present
func (r *GormCartRepository) AddItem(username, prodCode string, quantity uint32) (bool, error) {
	var existingItem model.CartItem
	err := r.db.
		Where("username = ? AND code = ?", username, prodCode).
		First(&existingItem).Error

	if err == nil { // cart item already exists => add the quantity
		totQuantity := existingItem.Quantity + quantity

		errAdd := r.db.
			Model(&model.CartItem{}).
			Where("username = ? AND code = ?", username, prodCode).
			Update("quantity", totQuantity).Error
		if errAdd != nil {
			r.logger.Error("Failed to add n=%d product with code '%s' to the cart of the user '%s': %v", quantity, prodCode, username, errAdd)
			return false, errAdd
		}

		r.logger.Info("Successful add of n=%d product with code '%s' to the cart of the user '%s'", quantity, prodCode, username)
		return true, nil
	}
	if err == gorm.ErrRecordNotFound { // cart item does not exist => create it with the right quantity
		newCartItem := model.CartItem{
			Username: username,
			Code:     prodCode,
			Quantity: quantity,
		}

		if err := r.db.Create(&newCartItem).Error; err != nil {
			r.logger.Error("Error creating cart item for user '%s', product with code '%s' and quantity=%d: %v", username, prodCode, quantity, err)
			return false, err
		}
		r.logger.Info("Successful creation of cart item for user '%s', product with code '%s' and quantity=%d", username, prodCode, quantity)
		return true, nil
	}

	r.logger.Error("Error retrieving cart item for user '%s' and product with code '%s': %v", username, prodCode, err)
	return false, err
}

// RemoveItem removes a specific product from the specified user's cart
func (r *GormCartRepository) RemoveItem(username, prodCode string) (bool, error) {
	res := r.db.
		Where("username = ? AND code = ?", username, prodCode).
		Delete(&model.CartItem{})

	if res.Error != nil {
		r.logger.Error("Failed to remove product '%s' from user '%s' cart: %v", prodCode, username, res.Error)
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		r.logger.Warn("No cart item found for product '%s' and user '%s'", prodCode, username)
		return false, nil
	}

	r.logger.Info("Successfully removed product '%s' from user '%s' cart", prodCode, username)
	return true, nil
}

// UpdateItemQuantity updates the quantity of a specific product in the specified user's cart
func (r *GormCartRepository) UpdateItemQuantity(username, prodCode string, quantity uint32) (bool, error) {
	if quantity == 0 {
		return r.RemoveItem(username, prodCode)
	}

	res := r.db.Model(&model.CartItem{}).
		Where("username = ? and code = ?", username, prodCode).
		Update("quantity", quantity)

	if res.Error != nil {
		r.logger.Error("Error while updating quantity for pruduct '%s' of user '%s' cart: %v", prodCode, username, res.Error)
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		r.logger.Warn("No cart item found for user '%s' and product with code '%s'", username, prodCode)
		return false, nil
	}

	r.logger.Info("Updated to '%d' the quantity of products with code '%s' in user '%s' cart", quantity, prodCode, username)
	return true, nil
}

// ListCartItems retrieves all cart items for the specified user
func (r *GormCartRepository) ListCartItems(username string) (bool, []*pb.CartItem, error) {
	var cartItemModels []model.CartItem
	err := r.db.
		Where("username = ?", username).
		Find(&cartItemModels).Error

	if err != nil {
		r.logger.Error("Error while retrieving all cart items for user '%s': %v", username, err)
		return false, nil, err
	}
	if len(cartItemModels) == 0 { // no products found
		return false, nil, nil
	}

	var cartItems []*pb.CartItem
	for _, cartItem := range cartItemModels {
		cartItems = append(cartItems, model.ModelItemToProtoItem(&cartItem))
	}

	r.logger.Info("Retrieved all cart items for user '%s'", username)
	return true, cartItems, nil
}

// ClearCart removes all products from the specified user's cart
func (r *GormCartRepository) ClearCart(username string) (bool, error) {
	res := r.db.
		Where("username = ?", username).
		Delete(&model.CartItem{})

	if res.Error != nil {
		r.logger.Error("Error clearing cart for user '%s': %v", username, res.Error)
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		r.logger.Warn("No items found in cart for user '%s'", username)
		return false, nil
	}

	r.logger.Info("Cleared cart for user '%s'", username)
	return true, nil
}

func (r *GormCartRepository) RemoveProductFromAllCarts(prodCode string) (bool, error) {
	res := r.db.
		Where("code = ?", prodCode).
		Delete(&model.CartItem{})

	if res.Error != nil {
		r.logger.Error("Error removing product with code '%s' from all carts: %v", prodCode, res.Error)
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		r.logger.Warn("No items found in cart for product with code '%s'", prodCode)
		return false, nil
	}

	r.logger.Info("Removed all items related to product with code '%s'", prodCode)
	return true, nil
}
