package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/grpc/clients"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/http/server"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/manager"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/orchestrator"
)

func main() {
	logLevel := logger.ParseLogLevel("LOG_LEVEL")
	myLogger := logger.NewStdLogger(logLevel, "orchestrator-main")

	grpcPort := GetEnvOrFatal(myLogger, "GRPC_PORT")
	addresses := map[string]string{
		"auth": fmt.Sprintf("%s:%s", GetEnvOrFatal(myLogger, "AUTH_NAME"), grpcPort),
		//"product": fmt.Sprintf("%s:%s", GetEnvOrFatal(myLogger, "PRODUCT_NAME"), grpcPort),
		//"cart": fmt.Sprintf("%s:%s", GetEnvOrFatal(myLogger, "CART_NAME"), grpcPort),
		//"order": fmt.Sprintf("%s:%s", GetEnvOrFatal(myLogger, "ORDER_NAME"), grpcPort),
	}

	serviceManager := manager.NewServiceManager(addresses, logger.NewStdLogger(logLevel, "service-manager"), 15*time.Second, 2*time.Second)
	serviceManager.StartMonitoring()
	defer serviceManager.Stop()

	authClient := clients.NewAuthClient("auth", serviceManager, logger.NewStdLogger(logLevel, "auth-client"), 1*time.Second)
	srv_orch := orchestrator.NewServiceOrchestrator(authClient, logger.NewStdLogger(logLevel, "service-orchestrator"))

	sessionSecret := GetEnvOrFatal(myLogger, "SESSION_SECRET")
	router := gin.Default()

	store := cookie.NewStore([]byte(sessionSecret))
	router.Use(sessions.Sessions("ecommerce_session", store))

	router.LoadHTMLGlob("./orchestrator/templates/*.tmpl")

	webServer := server.NewHTTPWebServer(router, srv_orch, logger.NewStdLogger(logLevel, "web-server-HTTP"))

	router.GET("/", webServer.RootHandler)

	app := router.Group("/app")
	app.Use(webServer.AuthRequired())
	{
		app.GET("/", webServer.IndexHandler)
		app.GET("/change-password", webServer.ChangePasswordGetHandler)
		app.POST("/change-password", webServer.ChangePasswordPostHandler)

		app.GET("/user-profile", webServer.UserProfileGetHandler)
		app.POST("/user-profile", webServer.UserProfilePostHandler)

		app.GET("/logout", webServer.LogoutHandler)
	}

	admin := app.Group("/admin")
	admin.Use(webServer.AuthRequired(), webServer.AdminRequired())
	{
		admin.GET("/users", webServer.UsersGetHandler)
		admin.POST("/users/:username/role", webServer.SetUserRolePostHandler)
	}

	router.GET("/login", webServer.LoginGetHandler)
	router.POST("/login", webServer.LoginPostHandler)

	router.GET("/register", webServer.RegisterGetHandler)
	router.POST("/register", webServer.RegisterPostHandler)

	webServiceAddress := fmt.Sprintf(":%s", GetEnvOrFatal(myLogger, "WEB_SERVICE_PORT"))
	webServer.Run(webServiceAddress)
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
		logger.Fatal("Environment variable '%s' not set", key)
	}
	return val
}
