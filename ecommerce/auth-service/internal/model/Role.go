package model

import pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"

// Role represents a user role within the system (e.g., admin, client).
//
// Fields:
//   - ID: unique identifier for the role (auto-incremented primary key).
//   - Name: unique name of the role (e.g., "admin", "client"); required and must be unique.
type Role struct {
    ID   int    `gorm:"column:id;primaryKey"`
    Name string `gorm:"column:name;type:varchar(20);uniqueIndex;not null"`
}

// Constants representing user roles in the database.
const (
    RoleUnspecified = 0
	RoleClient = 1 // Regular client user
	RoleAdmin  = 2 // Administrative user
)

// ModelRoleToProtoRole converts an int role into a pb.Role
func ModelRoleToProtoRole(role int) pb.Role {
    switch role {
    case RoleClient:
        return pb.Role_CLIENT
    case RoleAdmin:
        return pb.Role_ADMIN
    default:
        return pb.Role_UNSPECIFIED
    }
}

// ProtoRoleToModelRole converts a pb.Role into an int role
func ProtoRoleToModelRole(role pb.Role) int {
    switch role {
    case pb.Role_CLIENT:
        return RoleClient
    case pb.Role_ADMIN:
        return RoleAdmin
    default:
        return RoleUnspecified
    }
}