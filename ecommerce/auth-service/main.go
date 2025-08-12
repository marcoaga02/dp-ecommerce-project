package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal/repository"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	logLevel := logger.ParseLogLevel("LOG_LEVEL")
	myLogger := logger.NewStdLogger(logLevel, "auth-service")

	port := GetEnvOrFatal(myLogger, "GRPC_PORT")

	dsn := fmt.Sprintf(
		"root:%s@tcp(%s:%s)/%s?parseTime=true",
		GetEnvOrFatal(myLogger, "DB_PASSWORD"),
		GetEnvOrFatal(myLogger, "DB_HOST"),
		GetEnvOrFatal(myLogger, "DB_PORT"),
		GetEnvOrFatal(myLogger, "DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		myLogger.Fatal("Failed to connect to DB: %v", err)
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		myLogger.Fatal("Failed to listen on port %s: %v", port, err)
	}

	gormAuthRepo := repository.NewGormAuthRepository(db, logger.NewStdLogger(logLevel, "auth-service/db"))
	if err := gormAuthRepo.CreateDefaultUsers(); err != nil {
		myLogger.Fatal("Failed to create default users: %v", err)
	}

	authServer := internal.NewAuthServer(gormAuthRepo, logger.NewStdLogger(logLevel, "auth-service/server"))

	grpcServer := grpc.NewServer()
	pb.RegisterAuthenticationServer(grpcServer, authServer)

	// gRPC Health Server
	healthServer := health.NewServer()
	healthServer.SetServingStatus("auth.Authentication", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	myLogger.Info("Auth service listening on port %s", port)

	if err := grpcServer.Serve(lis); err != nil {
		myLogger.Fatal("Failed to serve gRPC server: %v", err)
	}
}

// GetEnvOrFatal retrieves the value of the environment variable identified by key.
// If the environment variable is not set (empty string), it logs a fatal error and terminates the program.
//
// Parameters:
//   - logger: the logger instance used to log fatal errors
//   - key: the environment variable key to retrieve
//
// Returns:
//   - string: the value of the environment variable
func GetEnvOrFatal(logger logger.Logger, key string) string {
	val := os.Getenv(key)
	if val == "" {
		logger.Fatal("Environment variable %s not set", key)
	}
	return val
}
