package main

import (
	"fmt"
	"net"
	"os"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/order-service/internal"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/order-service/internal/repository"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	logLevel := logger.ParseLogLevel("LOG_LEVEL")
	myLogger := logger.NewStdLogger(logLevel, "order-service/main")

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

	gormOrderRepo := repository.NewGormOrderRepository(db, logger.NewStdLogger(logLevel, "order-service/db"))
	orderServer := internal.NewOrderServer(gormOrderRepo, logger.NewStdLogger(logLevel, "order-service/server"))

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, orderServer)

	// gRPC Health Server
	healthServer := health.NewServer()
	healthServer.SetServingStatus("order.OrderService", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	myLogger.Info("Order service listening on port %s", port)

	if err := grpcServer.Serve(lis); err != nil {
		myLogger.Fatal("gRPC order server failed: %v", err)
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