package repository

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"

// AuthDB defines the interface for authentication and user management operations.
//
// It includes methods to:
//   - Authenticate a user (Login)
//   - Register a new user (Register)
//   - Change a user's password (ChangePassword)
//   - Set a user's role (SetUserRole)
//   - Retrieve a user's role (GetUserRole)
//
// This interface should be implemented by any data store
// responsible for handling user-related operations.

type AuthServiceInterface interface {
	Login(username, password string) (bool, pb.Role, error)
	Register(username, password, email, phone string) (bool, error)
	ChangePassword(username, oldPassword, newPassword string) (bool, error)
	SetUserRole(username string, role pb.Role) (bool, error)
	GetUserRole(username string) (bool, pb.Role, error)
}