package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"

type User struct {
	Username string
	Email    string
	Phone    string
	Role     Role
}

// ProtoUserToModelUser converts a pb.User into a model.User
func ProtoUserToModelUser(user *pb.User) *User {
	if user == nil {
		return nil
	}
	return &User{
		Username: user.GetUsername(),
		Email:    user.GetEmail(),
		Phone:    user.GetPhone(),
		Role:     ProtoRoleToModelRole(user.GetRole()),
	}
}

// ModelUserToProtoUser converts a model.User into a pb.User
func ModelUserToProtoUser(user *User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
		Role:     ModelRoleToProtoRole(user.Role),
	}
}

// ProtoUsersListToModelUsersList converts a []*pb.User into a []*model.User
func ProtoUsersListToModelUsersList(users []*pb.User) []*User {
	if users == nil {
		return nil
	}

	var modelUsers []*User

	for _, user := range users {
		modelUsers = append(modelUsers, ProtoUserToModelUser(user))
	}
	return modelUsers
}
