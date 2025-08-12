package repository

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"

type AuthServiceInterface interface {
	Login(username, password string) (bool, *pb.User, error)
	Register(username, password, email, phone string) (bool, error)
	ChangePassword(username, oldPassword, newPassword string) (bool, error)
	UpdateUser(username, email, phone string, role pb.Role) (bool, error)
	GetUser(username string) (bool, *pb.User, error)
	GetUsers() (bool, []*pb.User, error)
}
