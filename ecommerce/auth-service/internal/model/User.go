package model

import (
	"fmt"

	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
)

// User represents a user account in the system.
//
// Fields:
//   - Username: unique username (PK).
//   - PasswordHash: hashed password for secure authentication.
//   - Email: unique email address associated with the user.
//   - Phone: phone number.
//   - RoleID: foreign key referencing the user's role.
//   - Role: GORM relation to the Role entity.
type User struct {
	Username     string `gorm:"column:username;type:varchar(32);primaryKey"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null"`
	Email        string `gorm:"column:email;type:varchar(64);uniqueIndex;not null"`
	Phone        string `gorm:"column:phone;type:varchar(20);not null"`
	RoleID       int    `gorm:"column:role_id;not null;default:1"` // foreign key
	Role         Role   `gorm:"foreignKey:RoleID"`                 // GORM relationship
}

// ModelUserToProtoUser converts a model.User into a pb.User
func ModelUserToProtoUser(user *User) (*pb.User, error) {
	if user == nil {
		return nil, fmt.Errorf("Input argument is nil")
	}

	role := ModelRoleToProtoRole(user.RoleID)
	if role == RoleUnspecified {
		return nil, fmt.Errorf("Invalid user role id '%d'", user.RoleID)
	}

	return &pb.User{
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
		Role:     role,
	}, nil
}
