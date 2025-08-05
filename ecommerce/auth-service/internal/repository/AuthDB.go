package repository

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"

type AuthDB interface {
    Login(username, password string) (bool, pb.Role, error)
    Register(username, password, email, phone string) (bool, error)
	ChangePassword(username, old_password, new_password string) (bool, error)
    SetUserRole(username string, role pb.Role) (bool, error)
    GetUserRole(username string) (bool, pb.Role, error)
}