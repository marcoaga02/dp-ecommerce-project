package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/model"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/orchestrator/internal/orchestrator"
)

type HTTPWebServer struct {
	router       *gin.Engine
	orchestrator *orchestrator.ServiceOrchestrator
	logger       logger.Logger
}

const interalServerErrorMsg = "Internal server error. Please try again later."

func NewHTTPWebServer(router *gin.Engine, orchestrator *orchestrator.ServiceOrchestrator, logger logger.Logger) *HTTPWebServer {
	return &HTTPWebServer{
		router:       router,
		orchestrator: orchestrator,
		logger:       logger,
	}
}

func (s *HTTPWebServer) Run(addr string) error {
	s.logger.Info("Starting HTTP server on '%s'", addr)
	return s.router.Run(addr)
}

func (s *HTTPWebServer) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("username")
		s.logger.Info("Authentication required session user: %v", user)
		if user == nil {
			s.setErrorMessage(c, "Please log in to access this page")
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *HTTPWebServer) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, ok := s.getUsernameFromSessionOrRedirect(c)
		if !ok {
			return // method getUsernameFromSessionOrRedirect already did the redirect
		}

		succ, user, err := s.orchestrator.GetUser(username)
		if err != nil {
			s.logger.Error("User retrieval error in AdminRequired: %v", err)
			s.setErrorMessage(c, interalServerErrorMsg)
			c.Redirect(http.StatusFound, "/app/")
			c.Abort()
			return
		}
		if !succ {
			s.logger.Warn("User retrieval failed in AdminRequired: User not found")
			s.setErrorMessage(c, "User not found")
			c.Redirect(http.StatusFound, "/app/")
			c.Abort()
			return
		}

		if user.Role != model.RoleAdmin {
			s.setErrorMessage(c, "Access forbidden: Admins only")
			c.Redirect(http.StatusFound, "/app/")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *HTTPWebServer) RootHandler(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("username") != nil {
		c.Redirect(http.StatusFound, "/app/")
		return
	}
	c.Redirect(http.StatusFound, "/login")
}

func (s *HTTPWebServer) IndexHandler(c *gin.Context) {
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	role, err := s.getUserRole(username)
	roleStr, ok := model.RoleMapRoleToStr[role]
	if err != nil || !ok {
		s.logger.Error("Error fetching user role or unknown role: %v", err)
		c.HTML(http.StatusInternalServerError, "index.tmpl", gin.H{
			"Title":      "Welcome - Ecommerce",
			"Error":      interalServerErrorMsg,
			"IsLoggedIn": isLoggedIn,
		})
		return
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"Title":      "Welcome - Ecommerce",
		"Username":   username,
		"Role":       roleStr,
		"IsLoggedIn": isLoggedIn,
		"Flash":      s.getFlashMessage(c),
		"Error":      s.getErrorMessage(c),
	})
}

/*
* USER HANDLERS
 */

func (s *HTTPWebServer) LoginGetHandler(c *gin.Context) {
	isLoggedIn := false
	session := sessions.Default(c)
	if session.Get("username") != nil {
		c.Redirect(http.StatusFound, "/app/")
		return
	}
	c.HTML(http.StatusOK, "login.tmpl", gin.H{
		"Title":      "Login - Ecommerce",
		"Flash":      s.getFlashMessage(c),
		"Error":      s.getErrorMessage(c),
		"IsLoggedIn": isLoggedIn,
	})
}

func (s *HTTPWebServer) LoginPostHandler(c *gin.Context) {
	session := sessions.Default(c)

	username := c.PostForm("username")
	password := c.PostForm("password")

	succ, role, err := s.orchestrator.Login(username, password)
	if err != nil {
		s.logger.Error("Login error: %v", err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	if !succ {
		s.logger.Warn("Login failed: Invalid username or password")
		s.setErrorMessage(c, "Invalid username or password")
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	session.Set("username", username)
	err = session.Save()
	if err != nil {
		s.logger.Error("Session save error: %v", err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	var flash_role_msg string
	if role.String() == model.RoleAdmin.String() {
		flash_role_msg = "You are an administrator!"
	} else {
		flash_role_msg = "You are a client!"
	}

	s.setFlashMessage(c, fmt.Sprintf("Login successfull! %s", flash_role_msg))
	c.Redirect(http.StatusSeeOther, "/app/")
}

func (s *HTTPWebServer) RegisterGetHandler(c *gin.Context) {
	isLoggedIn := false
	session := sessions.Default(c)
	if session.Get("username") != nil {
		c.Redirect(http.StatusFound, "/app/")
		return
	}
	c.HTML(http.StatusOK, "register.tmpl", gin.H{
		"Title":      "Register - Ecommerce",
		"Flash":      s.getFlashMessage(c),
		"Error":      s.getErrorMessage(c),
		"IsLoggedIn": isLoggedIn,
	})
}

func (s *HTTPWebServer) RegisterPostHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	email := c.PostForm("email")
	phone := c.PostForm("phone")

	succ, err := s.orchestrator.Register(username, password, email, phone)
	if err != nil {
		s.logger.Error("Registration error: %v", err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/register")
		return
	}
	if !succ {
		s.logger.Warn("Registration failed: Username or email already used")
		s.setErrorMessage(c, "Username or email already used")
		c.Redirect(http.StatusSeeOther, "/register")
		return
	}

	s.setFlashMessage(c, "Registration successful. You can now log in!")
	c.Redirect(http.StatusSeeOther, "/login")
}

func (s *HTTPWebServer) ChangePasswordGetHandler(c *gin.Context) {
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	c.HTML(http.StatusOK, "changePassword.tmpl", gin.H{
		"Title":      "Change Password - Ecommerce",
		"Username":   username,
		"Flash":      s.getFlashMessage(c),
		"Error":      s.getErrorMessage(c),
		"IsLoggedIn": isLoggedIn,
	})
}

func (s *HTTPWebServer) ChangePasswordPostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}
	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")

	succ, err := s.orchestrator.ChangePassword(username, currentPassword, newPassword)
	if err != nil {
		s.logger.Error("Password change error: %v", err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/user/password")
		return
	}
	if !succ {
		s.logger.Warn("Change password failed: Invalid current password")
		s.setErrorMessage(c, "Invalid current password")
		c.Redirect(http.StatusSeeOther, "/app/user/password")
		return
	}

	s.setFlashMessage(c, "Password changed successfully!")
	c.Redirect(http.StatusSeeOther, "/app/")
}

func (s *HTTPWebServer) UserProfileGetHandler(c *gin.Context) {
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}
	succ, user, err := s.orchestrator.GetUser(username)
	if err != nil {
		s.logger.Error("User retrieval error: %v", err)
		c.HTML(http.StatusInternalServerError, "profile.tmpl", gin.H{
			"Title":      "User Profile - Ecommerce",
			"Error":      interalServerErrorMsg,
			"IsLoggedIn": isLoggedIn,
		})
		return
	}
	if !succ {
		s.logger.Warn("User retrieval failed: User not found")
		c.HTML(http.StatusNotFound, "profile.tmpl", gin.H{ // ← profile.tmpl + 404!
			"Title":      "User Profile - Ecommerce",
			"Error":      "User not found",
			"IsLoggedIn": isLoggedIn,
		})
		return
	}

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"Title": "User Profile - Ecommerce",
		"User": gin.H{
			"Username": username,
			"Email":    user.Email,
			"Phone":    user.Phone,
			"Role":     user.Role.String(),
		},
		"Flash":      s.getFlashMessage(c),
		"Error":      s.getErrorMessage(c),
		"IsLoggedIn": isLoggedIn,
	})
}

func (s *HTTPWebServer) UserProfilePostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}
	email := c.PostForm("email")
	phone := c.PostForm("phone")

	succ, err := s.orchestrator.UpdateUser(username, email, phone, model.RoleUnspecified)

	if err != nil {
		s.logger.Error("User update error: %v", err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/user/profile")
		return
	}
	if !succ {
		s.logger.Warn("User update failed")
		s.setErrorMessage(c, "User update failed")
		c.Redirect(http.StatusSeeOther, "/app/user/profile")
		return
	}
	s.setFlashMessage(c, "User updated successfully!")
	c.Redirect(http.StatusSeeOther, "/app/user/profile")
}

func (s *HTTPWebServer) UsersGetHandler(c *gin.Context) {
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	succ, users, err := s.orchestrator.GetUsers()
	if err != nil {
		s.logger.Error("Error retrieving users: %v", err)
		c.HTML(http.StatusInternalServerError, "adminUsers.tmpl", gin.H{
			"Title":      "Manage Users - Ecommerce",
			"Username":   username,
			"Error":      interalServerErrorMsg,
			"IsLoggedIn": isLoggedIn,
		})
		return
	}
	if !succ {
		s.logger.Warn("Users retrieval failed")
		c.HTML(http.StatusInternalServerError, "adminUsers.tmpl", gin.H{
			"Title":      "Manage Users - Ecommerce",
			"Username":   username,
			"Error":      "Users retrieval failed",
			"IsLoggedIn": isLoggedIn,
		})
		return
	}

	var usersForTemplate []gin.H
	for _, u := range users {
		if u.Username != username {
			usersForTemplate = append(usersForTemplate, gin.H{
				"Username": u.Username,
				"Email":    u.Email,
				"Phone":    u.Phone,
				"Role":     u.Role.String(),
			})
		}
	}
	c.HTML(http.StatusOK, "adminUsers.tmpl", gin.H{
		"Title":      "Manage Users - Ecommerce",
		"Users":      usersForTemplate,
		"Username":   username,
		"Flash":      s.getFlashMessage(c),
		"Error":      s.getErrorMessage(c),
		"IsLoggedIn": isLoggedIn,
	})
}

func (s *HTTPWebServer) SetUserRolePostHandler(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		s.logger.Warn("Missing username parameter in set user role handler")
		s.setErrorMessage(c, "Missing username parameter")
		c.Redirect(http.StatusSeeOther, "/app/admin/users")
		return
	}

	newRole := c.PostForm("role")

	var role model.Role
	role, ok := model.RoleMapStrToRole[newRole]
	if !ok {
		role = model.RoleUnspecified
	}

	if role == model.RoleUnspecified {
		s.logger.Warn("Invalid role specified in set user role request")
		s.setErrorMessage(c, "Invalid role specified")
		c.Redirect(http.StatusSeeOther, "/app/admin/users")
		return
	}

	succ, err := s.orchestrator.SetUserRole(username, role)
	if err != nil {
		s.logger.Error("Error updating role for user %s: %v", username, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/admin/users")
		return
	}
	if !succ {
		s.logger.Warn("Role update failed for user %s", username)
		s.setErrorMessage(c, "Role setting failed")
		c.Redirect(http.StatusSeeOther, "/app/admin/users")
		return
	}

	s.setFlashMessage(c, fmt.Sprintf("Role updated successfully for user '%s'", username))
	c.Redirect(http.StatusSeeOther, "/app/admin/users")
}

func (s *HTTPWebServer) LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		s.logger.Error("Session save error on logout: %v", err)
	}
	s.setFlashMessage(c, "You have been logged out!")
	c.Redirect(http.StatusFound, "/login")
}

func (s *HTTPWebServer) getUsernameFromSessionOrRedirect(c *gin.Context) (string, bool) {
	session := sessions.Default(c)
	username, ok := session.Get("username").(string)
	if !ok || username == "" {
		session.Clear()
		_ = session.Save()
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return "", false
	}
	return username, true
}

func (s *HTTPWebServer) getUserRole(username string) (model.Role, error) {
	succ, user, err := s.orchestrator.GetUser(username)
	if err != nil {
		return model.RoleUnspecified, err
	}
	if !succ {
		return model.RoleUnspecified, fmt.Errorf("user not found")
	}
	return user.Role, nil
}

/*
* PRODUCT HANDLERS
 */

func (s *HTTPWebServer) ListProductsGetHandler(c *gin.Context) {
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	role, err := s.getUserRole(username)
	roleStr, ok := model.RoleMapRoleToStr[role]
	if err != nil || !ok {
		s.logger.Error("Error fetching user role or unknown role: %v", err)
		c.HTML(http.StatusInternalServerError, "productsList.tmpl", gin.H{
			"Title":      "Products List - Ecommerce",
			"Username":   username,
			"Error":      interalServerErrorMsg,
			"IsLoggedIn": isLoggedIn,
		})
		return
	}

	succ, prods, err := s.orchestrator.GetProductLists()
	if err != nil {
		s.logger.Error("Error retrieving products: %v", err)
		c.HTML(http.StatusInternalServerError, "productsList.tmpl", gin.H{
			"Title":      "Products List - Ecommerce",
			"Username":   username,
			"Role":       roleStr,
			"Error":      interalServerErrorMsg,
			"IsLoggedIn": isLoggedIn,
		})
		return
	}
	if !succ {
		s.logger.Warn("Products retrieval failed")
		c.HTML(http.StatusInternalServerError, "productsList.tmpl", gin.H{
			"Title":      "Products List - Ecommerce",
			"Username":   username,
			"Role":       roleStr,
			"Error":      "Products retrieval failed",
			"IsLoggedIn": isLoggedIn,
		})
		return
	}

	var prodsForTemplate []gin.H
	for _, p := range prods {
		prodsForTemplate = append(prodsForTemplate, gin.H{
			"Code":        p.Code,
			"Name":        p.Name,
			"Size":        p.Size,
			"Color":       p.Color,
			"Description": p.Description,
			"Stock":       p.Stock,
			"Price":       p.Price,
		})
	}
	c.HTML(http.StatusOK, "productsList.tmpl", gin.H{
		"Title":      "Products List - Ecommerce",
		"Products":   prodsForTemplate,
		"Username":   username,
		"Role":       roleStr,
		"Flash":      s.getFlashMessage(c),
		"Error":      s.getErrorMessage(c),
		"IsLoggedIn": isLoggedIn,
	})
}

func (s *HTTPWebServer) NewProductGetHandler(c *gin.Context) {
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	c.HTML(http.StatusOK, "createProduct.tmpl", gin.H{
		"Title":      "Products List - Ecommerce",
		"Sizes":      model.AllSizes,
		"Username":   username,
		"Flash":      s.getFlashMessage(c),
		"Error":      s.getErrorMessage(c),
		"IsLoggedIn": isLoggedIn,
	})
}

func (s *HTTPWebServer) NewProductPostHandler(c *gin.Context) {
	code := c.PostForm("code")
	name := c.PostForm("name")
	size := c.PostForm("size")
	color := c.PostForm("color")
	description := c.PostForm("description")
	stock := c.PostForm("stock")
	price := c.PostForm("price")

	sizeModel, ok := model.SizeMapStrToSize[size]
	if !ok || sizeModel == model.SizeUnspecified {
		s.logger.Warn("Invalid size specified in new product handler")
		s.setErrorMessage(c, "Invalid size specified")
		c.Redirect(http.StatusSeeOther, "/app/product/new")
		return
	}

	stockInt64, err := strconv.ParseInt(stock, 10, 32)
	if err != nil {
		s.logger.Warn("Invalid stock value in new product handler")
		s.setErrorMessage(c, "Invalid stock value")
		c.Redirect(http.StatusSeeOther, "/app/product/new")
		return
	}
	stockInt32 := int32(stockInt64)

	priceFloat64, err := strconv.ParseFloat(price, 64)
	if err != nil {
		s.logger.Warn("Invalid price value in new product handler")
		s.setErrorMessage(c, "Invalid price value")
		c.Redirect(http.StatusSeeOther, "/app/product/new")
		return
	}

	succ, err := s.orchestrator.CreateProduct(code, name, sizeModel, color, description, stockInt32, priceFloat64)
	if err != nil {
		s.logger.Error("Error creating the new product with code '%s': %v", code, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/new")
		return
	}
	if !succ {
		s.logger.Warn("Failed creation for product with code '%s'", code)
		s.setErrorMessage(c, fmt.Sprintf("Product creation failed: already exists a product with the same code '%s'", code))
		c.Redirect(http.StatusSeeOther, "/app/product/new")
		return
	}

	s.setFlashMessage(c, fmt.Sprintf("Product with code '%s' successfully created", code))
	c.Redirect(http.StatusSeeOther, "/app/product/")
}

func (s *HTTPWebServer) EditProductGetHandler(c *gin.Context) {
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	code := c.Param("code")
	if code == "" {
		s.logger.Warn("Missing code parameter in edit product handler")
		s.setErrorMessage(c, "Missing product code")
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	succ, prod, err := s.orchestrator.GetProduct(code)
	if err != nil {
		s.logger.Error("Error during retrieval of product with code '%s': %v", code, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}
	if !succ {
		s.logger.Warn("Product with code '%s' not found")
		s.setErrorMessage(c, "Product not found")
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	c.HTML(http.StatusOK, "editProduct.tmpl", gin.H{
		"Title":    "Edit Product - Ecommerce",
		"Sizes":    model.AllSizes,
		"Username": username,
		"Product": gin.H{
			"Code":        prod.Code,
			"Name":        prod.Name,
			"Size":        prod.Size.String(),
			"Color":       prod.Color,
			"Description": prod.Description,
			"Stock":       prod.Stock,
			"Price":       prod.Price,
		},
		"Flash":      s.getFlashMessage(c),
		"Error":      s.getErrorMessage(c),
		"IsLoggedIn": isLoggedIn,
	})
}

func (s *HTTPWebServer) EditProductPostHandler(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		s.logger.Warn("Missing code parameter in edit product handler")
		s.setErrorMessage(c, "Missing product code")
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}
	name := c.PostForm("name")
	size := c.PostForm("size")
	color := c.PostForm("color")
	description := c.PostForm("description")
	stock := c.PostForm("stock")
	price := c.PostForm("price")

	sizeModel, ok := model.SizeMapStrToSize[size]
	if !ok || sizeModel == model.SizeUnspecified {
		s.logger.Warn("Invalid size specified in edit product handler")
		s.setErrorMessage(c, "Invalid size specified")
		c.Redirect(http.StatusSeeOther, "/app/product/"+code+"/edit")
		return
	}

	stockInt64, err := strconv.ParseInt(stock, 10, 32)
	if err != nil {
		s.logger.Warn("Invalid stock value in edit product handler")
		s.setErrorMessage(c, "Invalid stock value")
		c.Redirect(http.StatusSeeOther, "/app/product/"+code+"/edit")
		return
	}
	stockInt32 := int32(stockInt64)

	priceFloat64, err := strconv.ParseFloat(price, 64)
	if err != nil {
		s.logger.Warn("Invalid price value in edit product handler")
		s.setErrorMessage(c, "Invalid price value")
		c.Redirect(http.StatusSeeOther, "/app/product/"+code+"/edit")
		return
	}

	succ, err := s.orchestrator.UpdateProduct(code, name, sizeModel, color, description, stockInt32, priceFloat64)
	if err != nil {
		s.logger.Error("Error updating the product with code '%s': %v", code, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}
	if !succ {
		s.logger.Warn("Failed update for product with code '%s'", code)
		s.setErrorMessage(c, fmt.Sprintf("Product update failed: product not found"))
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	s.logger.Info("Successful update of product with code '%s'", code)
	c.Redirect(http.StatusSeeOther, "/app/product/")
}

func (s *HTTPWebServer) DeleteProductHandler(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		s.logger.Warn("Missing code parameter in delete product handler")
		s.setErrorMessage(c, "Missing product code")
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	succ, err := s.orchestrator.DeleteProduct(code)
	if err != nil {
		s.logger.Error("Error deleting the product with code '%s': %v", code, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}
	if !succ {
		s.logger.Warn("Failed delete of product with code '%s'", code)
		s.setErrorMessage(c, fmt.Sprintf("Product delete failed: product not found"))
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	s.logger.Info("Successful delete of product with code '%s'", code)
	c.Redirect(http.StatusSeeOther, "/app/product/")
}

func (s *HTTPWebServer) ProductGetHandler(c *gin.Context) {
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	code := c.Param("code")
	if code == "" {
		s.logger.Warn("Missing code parameter in get product handler")
		s.setErrorMessage(c, "Missing product code")
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	succ, prod, err := s.orchestrator.GetProduct(code)
	if err != nil {
		s.logger.Error("Error during retrieval of product with code '%s': %v", code, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}
	if !succ {
		s.logger.Warn("Product with code '%s' not found")
		s.setErrorMessage(c, "Product not found")
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	c.HTML(http.StatusOK, "buyProduct.tmpl", gin.H{
		"Title":       "Buy Product - Ecommerce",
		"Username":    username,
		"Code":        prod.Code,
		"Name":        prod.Name,
		"Size":        prod.Size.String(),
		"Color":       prod.Color,
		"Description": prod.Description,
		"Stock":       prod.Stock,
		"Price":       prod.Price,
		"Flash":       s.getFlashMessage(c),
		"Error":       s.getErrorMessage(c),
		"IsLoggedIn":  isLoggedIn,
	})
}

/*
* UTILITIES
 */

func (s *HTTPWebServer) setFlashMessage(c *gin.Context, message string) {
	session := sessions.Default(c)
	session.Set("flash", message)
	_ = session.Save()
}

func (s *HTTPWebServer) setErrorMessage(c *gin.Context, message string) {
	session := sessions.Default(c)
	session.Set("error", message)
	_ = session.Save()
}

func (s *HTTPWebServer) getFlashMessage(c *gin.Context) string {
	session := sessions.Default(c)
	flash := session.Get("flash")
	if flash != nil {
		session.Delete("flash")
		_ = session.Save()
		return flash.(string)
	}
	return ""
}

func (s *HTTPWebServer) getErrorMessage(c *gin.Context) string {
	session := sessions.Default(c)
	errorMsg := session.Get("error")
	if errorMsg != nil {
		session.Delete("error")
		_ = session.Save()
		return errorMsg.(string)
	}
	return ""
}
