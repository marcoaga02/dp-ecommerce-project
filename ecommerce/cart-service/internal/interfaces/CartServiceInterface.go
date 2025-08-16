package interfaces

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/cart"

type CartServiceInterface interface {
	// AddItem adds a product to the specified user's cart, increasing quantity if already present
	AddItem(username, prodCode string, quantity uint32) (bool, error)

	// RemoveItem removes a specific product from the specified user's cart
	RemoveItem(username, prodCode string) (bool, error)

	// UpdateItemQuantity updates the quantity of a specific product in the specified user's cart
	UpdateItemQuantity(username, prodCode string, quantity uint32) (bool, error)

	// ListCartItems retrieves all cart items for the specified user
	ListCartItems(username string) (bool, []*pb.CartItem, error)

	// ClearCart removes all products from the specified user's cart
	ClearCart(username string) (bool, error)

	// RemoveProductFromAllCarts removes all cart items related to a given product
	RemoveProductFromAllCarts(prodCode string) (bool, error)
}