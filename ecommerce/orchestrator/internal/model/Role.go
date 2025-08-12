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
