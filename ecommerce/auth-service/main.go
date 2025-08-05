package main

import (
	"fmt"
	"net"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"google.golang.org/grpc"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal/repository"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
)

func main() {
	logLevel := ParseLogLevel("LOG_LEVEL")
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

// ParseLogLevel reads the environment variable identified by key and converts its value
// to the corresponding logger.Level.
//
// If the environment variable is not set or contains an unrecognized value, it returns logger.Info as the default.
//
// Parameters:
//   - key: the name of the environment variable to read
//
// Returns:
//   - logger.Level: the corresponding log level constant
func ParseLogLevel(key string) logger.Level {
    level := os.Getenv(key)
    switch level {
    case "debug", "DEBUG":
        return logger.Debug
    case "info", "INFO":
        return logger.Info
    case "warn", "WARN", "warning", "WARNING":
        return logger.Warn
    case "error", "ERROR":
        return logger.Error
    default:
        return logger.Info // default level
    }
}

