package interfaces

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"

type AuthClientInterface interface {
	// Login authenticates a user with the given username and password.
	Login(username, password string) (bool, *pb.User, error)

	// Register creates a new user account with the provided credentials and contact info.
	Register(username, password, email, phone string) (bool, error)

	// ChangePassword updates the password for a given user.
	ChangePassword(username, oldPassword, newPassword string) (bool, error)

	// UpdateUser updates user email, phone or role.
	UpdateUser(username, email, phone string, role pb.Role) (bool, error)

	// GetUser retrieves the user information for the specified username.
	GetUser(username string) (bool, *pb.User, error)

	// GetUsers retrieves all users registered in the system.
	GetUsers() (bool, []*pb.User, error)
}
