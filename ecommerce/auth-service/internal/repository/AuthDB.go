package repository

type AuthDB interface {
    Login(username, password string) (bool, error)
    Register(username, password, email, phone string) (bool, error)
	ChangePassword(username, old_password, new_password string) (bool, error)
}