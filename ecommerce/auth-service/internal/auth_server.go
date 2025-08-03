package internal

import (
	"context"

	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
)

type AuthServer struct {
	pb.UnimplementedAuthenticationServer
}

var user = map[string]string{
    "alice": "password123",
    "bob":   "secret",
}

func (s *AuthServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	pass, ok := user[in.Username]
	if !ok || pass != in.Password {
		return &pb.LoginResponse{Success: false}, nil
	}
	return &pb.LoginResponse{Success: true}, nil
}

//func Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
//	
//}