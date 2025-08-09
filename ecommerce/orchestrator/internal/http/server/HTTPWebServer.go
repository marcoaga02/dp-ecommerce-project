package server

import (
	"net/http"

    "github.com/gin-contrib/sessions"
    //"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/orchestrator"
)

type HTTPWebServer struct {
	router     *gin.Engine
    orchestrator *orchestrator.ServiceOrchestrator
	logger     logger.Logger
}

func NewHTTPWebServer(router *gin.Engine, orchestrator *orchestrator.ServiceOrchestrator, logger logger.Logger) *HTTPWebServer {
	return &HTTPWebServer{
		router:     router,
        orchestrator: orchestrator,
		logger:     logger,
	}
}

func (s *HTTPWebServer) Run(addr string) error {
	s.logger.Info("Starting HTTP server on '%s'", addr)
	return s.router.Run(addr)
}

func (_ *HTTPWebServer) AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        user := session.Get("username")
        if user == nil {
            c.Redirect(http.StatusFound, "/login")
            c.Abort()
            return
        }
        c.Next()
    }
}

func (s *HTTPWebServer) IndexHandler(c *gin.Context) {
    session := sessions.Default(c)
    username := session.Get("username")
    role := session.Get("role")
    c.HTML(http.StatusOK, "index.html", gin.H{
        "username": username,
        "role": role,
    })
}

func (s *HTTPWebServer) LoginGetHandler(c *gin.Context) {
    session := sessions.Default(c)
    if session.Get("username") != nil {
        c.Redirect(http.StatusFound, "/")
        return
    }
    c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (s *HTTPWebServer) LoginPostHandler(c *gin.Context) {
    session := sessions.Default(c)

    username := c.PostForm("username")
	password := c.PostForm("password")

    succ, role, err := s.orchestrator.Login(username, password)
    if err != nil {
        s.logger.Error("Login error: %v", err)
		c.HTML(http.StatusOK, "login.html", gin.H{
            "error": "Internal server error: try again later",
            "username": username,
        })
		return
    }
    if !succ {
        s.logger.Warn("Login failed: Invalid username or password")
        c.HTML(http.StatusOK, "login.html", gin.H{
            "error": "Invalid username or password",
            "username": username,
        })
		return
	}

    session.Set("username", username)
    session.Set("role", role.String())
    err = session.Save()
    if err != nil {
        s.logger.Error("Session save error: %v", err)
        c.HTML(http.StatusInternalServerError, "login.html", gin.H{
            "error": "Internal server error: try again later",
        })
        return
    }

    c.HTML(http.StatusOK, "/", gin.H{
        "username": username,
        "role": role.String(),
    })
}

func (s *HTTPWebServer) RegisterHandler(c *gin.Context) {

}

func (s *HTTPWebServer) ChangePasswordHandler(c *gin.Context) {
    
}

func (s *HTTPWebServer) SetUserRoleHandler(c *gin.Context) {
    
}

func (s *HTTPWebServer) GetUserRoleHandler(c *gin.Context) {
    
}



/*func (srv *HTTPWebServer) handleLogin(c *gin.Context) {
	// Per leggere JSON dalla request:
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		srv.logger.Warn("Invalid login request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	success, role, err := srv.authClient.Login(loginReq.Username, loginReq.Password)
	if err != nil || !success {
		srv.logger.Warn("Login failed for user " + loginReq.Username + ": " + err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Risposta di successo (esempio)
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"role":    role.String(), // supponendo che pb.Role abbia String()
	})
}*/
