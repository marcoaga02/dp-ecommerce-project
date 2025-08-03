package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal"
)

func main() {
	lis, err := net.Listen("tcp", ":50051") // scegli una porta libera
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthenticationServer(s, &internal.AuthServer{})

	log.Println("Auth service listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}