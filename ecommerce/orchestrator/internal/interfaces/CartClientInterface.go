package interfaces

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/cart"

type CartClientInterface interface {
	// AddItem attempts to increase the quantity of a product into the user's cart
	AddItem(username, code string, quantity uint32) (bool, error)

	// RemoveItem attempts ro remove a product from the user's cart
	RemoveItem(username, code string) (bool, error)

	// UpdateItemQuantity attempts to update the quantity of a product into the user's cart
	UpdateItemQuantity(username, code string, quantity uint32) (bool, error)

	// ListCartItems retrievs the list of items into the user's cart
	ListCartItems(username string) (bool, []*pb.CartItem, error)

	// ClearCart attempts to remove all the products from the user's carts
	ClearCart(username string) (bool, error)

	// RemoveProductFromAllCarts removes all cart items related to a given product
	RemoveProductFromAllCarts(prodCode string) (bool, error)
}