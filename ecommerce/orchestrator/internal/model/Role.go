package model

import (
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
)

type Role string

const (
	RoleUnspecified Role = "unspecified"
	RoleClient      Role = "client"
	RoleAdmin       Role = "administrator"
)

var RoleMapStrToRole = map[string]Role{
	"client":        RoleClient,
	"administrator": RoleAdmin,
}

var RoleMapRoleToStr = map[Role]string{
	RoleClient: "client",
	RoleAdmin:  "administrator",
}

// ProtoRoleToModelRole converts a pb.Role into a model.Role
func ProtoRoleToModelRole(pr pb.Role) Role {
	switch pr {
	case pb.Role_CLIENT:
		return RoleClient
	case pb.Role_ADMIN:
		return RoleAdmin
	default:
		return RoleUnspecified
	}
}

// ModelRoleToProtoRole converts a model.Role into a pb.Role
func ModelRoleToProtoRole(r Role) pb.Role {
	switch r {
	case RoleClient:
		return pb.Role_CLIENT
	case RoleAdmin:
		return pb.Role_ADMIN
	default:
		return pb.Role_UNSPECIFIED
	}
}

func (r Role) String() string {
	return string(r)
}
