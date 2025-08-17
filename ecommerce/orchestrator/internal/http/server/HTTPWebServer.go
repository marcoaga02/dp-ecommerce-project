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

// NewHTTPWebServer returns a new instance of HTTPWebServer
func NewHTTPWebServer(router *gin.Engine, orchestrator *orchestrator.ServiceOrchestrator, logger logger.Logger) *HTTPWebServer {
	return &HTTPWebServer{
		router:       router,
		orchestrator: orchestrator,
		logger:       logger,
	}
}

// Run starts the execution of the webserver
func (s *HTTPWebServer) Run(addr string) error {
	s.logger.Info("Starting HTTP server on '%s'", addr)
	return s.router.Run(addr)
}

// AuthRequired acts as a middleware to ensure that a user is authenticated before accessing some reserved routes
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

// AdminRequired acts as a middleware to ensure that a user is an administrator before accessing some reserved routes
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

// RootHandler handles the GET to the root of the website.
//
// If the user is authenticated, it is redirected to the main page of the website; otherwise, it is redirected to the login page
func (s *HTTPWebServer) RootHandler(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("username") != nil {
		c.Redirect(http.StatusFound, "/app/")
		return
	}
	c.Redirect(http.StatusFound, "/login")
}

// IndexHandler handles the GET to the main page of the website
func (s *HTTPWebServer) IndexHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	role, err := s.getUserRole(username)
	roleStr, ok := model.RoleMapRoleToStr[role]
	if err != nil || !ok {
		s.logger.Error("Error fetching user role or unknown role")
		c.HTML(http.StatusInternalServerError, "index.html", gin.H{
			"Title":      "Welcome - Ecommerce",
			"Username":   username,
			"IsLoggedIn": isLoggedIn,
			"Error":      interalServerErrorMsg,
		})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title":      "Welcome - Ecommerce",
		"Username":   username,
		"Role":       roleStr,
		"IsLoggedIn": isLoggedIn,
		"Flash":      flashMsg,
		"Error":      errMsg,
	})
}

/*
* USER HANDLERS
 */

// LoginGetHandler handles the GET to the login route.
//
// If the user is already authenticated, it is redirected to the main page of the website
func (s *HTTPWebServer) LoginGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := false
	session := sessions.Default(c)
	if session.Get("username") != nil {
		c.Redirect(http.StatusFound, "/app/")
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title":      "Login - Ecommerce",
		"Flash":      flashMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// LoginPostHandler handles the POST of the login route
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

	s.logger.Info("Successful login for user '%s'", username)
	s.setFlashMessage(c, fmt.Sprintf("Login successfull! %s", flash_role_msg))
	c.Redirect(http.StatusSeeOther, "/app/")
}

// RegisterGetHandler handles the GET to the register route.
//
// If the user is already authenticated, it is redirected to the main page of the website
func (s *HTTPWebServer) RegisterGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := false
	session := sessions.Default(c)
	if session.Get("username") != nil {
		c.Redirect(http.StatusFound, "/app/")
		return
	}
	c.HTML(http.StatusOK, "register.html", gin.H{
		"Title":      "Register - Ecommerce",
		"Flash":      flashMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// RegisterPostHandler handles the POST to the register route
func (s *HTTPWebServer) RegisterPostHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	confirmPass := c.PostForm("confirm-password")
	email := c.PostForm("email")
	phone := c.PostForm("phone")

	if password != confirmPass {
		s.logger.Warn("Password and confirmation do not match in register request")
		s.setErrorMessage(c, "Your password confirmation doesn’t match the password.")
		c.Redirect(http.StatusSeeOther, "/register")
		return
	}

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

	s.logger.Info("Successful registration for user '%s'", username)
	s.setFlashMessage(c, "Registration successful. You can now log in!")
	c.Redirect(http.StatusSeeOther, "/login")
}

// ChangePasswordGetHandler handles the GET to the change password route
func (s *HTTPWebServer) ChangePasswordGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	c.HTML(http.StatusOK, "changePassword.html", gin.H{
		"Title":      "Change Password - Ecommerce",
		"Username":   username,
		"Flash":      flashMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// ChangePasswordPostHandler handles the POST to the change password route
func (s *HTTPWebServer) ChangePasswordPostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}
	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPass := c.PostForm("confirm-password")

	if newPassword != confirmPass {
		s.logger.Warn("Password and confirmation do not match in change password request")
		s.setErrorMessage(c, "Your password confirmation doesn’t match the new password.")
		c.Redirect(http.StatusSeeOther, "/app/user/password")
		return
	}

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

	s.logger.Info("Successful password change for user '%s'", username)
	s.setFlashMessage(c, "Password changed successfully!")
	c.Redirect(http.StatusSeeOther, "/app/")
}

// UserProfileGetHandler handles the GET to the user profile route
func (s *HTTPWebServer) UserProfileGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}
	succ, user, err := s.orchestrator.GetUser(username)
	if err != nil {
		s.logger.Error("User retrieval error: %v", err)
		c.HTML(http.StatusInternalServerError, "profile.html", gin.H{
			"Title":      "User Profile - Ecommerce",
			"Username":   username,
			"Error":      interalServerErrorMsg,
			"IsLoggedIn": isLoggedIn,
		})
		return
	}
	if !succ {
		s.logger.Warn("User retrieval failed: User not found")
		c.HTML(http.StatusNotFound, "profile.html", gin.H{
			"Title":      "User Profile - Ecommerce",
			"Username":   username,
			"Error":      "User not found",
			"IsLoggedIn": isLoggedIn,
		})
		return
	}

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"Title":      "User Profile - Ecommerce",
		"Username":   username,
		"Email":      user.Email,
		"Phone":      user.Phone,
		"Role":       user.Role.String(),
		"Flash":      flashMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// UserProfilePostHandler handles the POST to the user profile route
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

	s.logger.Info("Successful profile update for user '%s'", username)
	s.setFlashMessage(c, "User updated successfully!")
	c.Redirect(http.StatusSeeOther, "/app/user/profile")
}

// UsersGetHandler handles the GET to the "see list of users" route
func (s *HTTPWebServer) UsersGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	succ, users, err := s.orchestrator.GetUsers()
	if err != nil {
		s.logger.Error("Error retrieving users: %v", err)
		c.HTML(http.StatusInternalServerError, "adminUsers.html", gin.H{
			"Title":      "Manage Users - Ecommerce",
			"Username":   username,
			"Error":      interalServerErrorMsg,
			"IsLoggedIn": isLoggedIn,
		})
		return
	}
	if !succ {
		s.logger.Warn("No users found")
		c.HTML(http.StatusInternalServerError, "adminUsers.html", gin.H{
			"Title":      "Manage Users - Ecommerce",
			"Username":   username,
			"Error":      "No users found",
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
	c.HTML(http.StatusOK, "adminUsers.html", gin.H{
		"Title":      "Manage Users - Ecommerce",
		"Users":      usersForTemplate,
		"Username":   username,
		"Flash":      flashMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// SetUserRolePostHandler handle the POST to the set user role route
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

	s.logger.Info("Role updated successfully for user '%s'", username)
	s.setFlashMessage(c, fmt.Sprintf("Role updated successfully for user '%s'", username))
	c.Redirect(http.StatusSeeOther, "/app/admin/users")
}

// LogoutHandler handles the GET to the logout route
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

/*
* PRODUCT HANDLERS
 */

// ListProductsGetHandler handles the GET to the "list all products" route
func (s *HTTPWebServer) ListProductsGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	role, err := s.getUserRole(username)
	roleStr, ok := model.RoleMapRoleToStr[role]
	if err != nil || !ok {
		s.logger.Error("Error fetching user role or unknown role")
		c.HTML(http.StatusInternalServerError, "productsList.html", gin.H{
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
		c.HTML(http.StatusInternalServerError, "productsList.html", gin.H{
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
		c.HTML(http.StatusInternalServerError, "productsList.html", gin.H{
			"Title":      "Products List - Ecommerce",
			"Username":   username,
			"Role":       roleStr,
			"Error":      "No products found",
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
	c.HTML(http.StatusOK, "productsList.html", gin.H{
		"Title":      "Products List - Ecommerce",
		"Products":   prodsForTemplate,
		"Username":   username,
		"Role":       roleStr,
		"Flash":      flashMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// NewProductGetHandler handles the GET to the "create a new product" route
func (s *HTTPWebServer) NewProductGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	c.HTML(http.StatusOK, "createProduct.html", gin.H{
		"Title":      "Products List - Ecommerce",
		"Sizes":      model.AllSizes,
		"Username":   username,
		"Flash":      flashMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// NewProductPostHandler handles the POST to the "create a new product" route
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
	stockInt32 := uint32(stockInt64)

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

	s.logger.Info("Product with code '%s' successfully created", code)
	s.setFlashMessage(c, fmt.Sprintf("Product with code '%s' successfully created", code))
	c.Redirect(http.StatusSeeOther, "/app/product/")
}

// EditProductGetHandler handles the GET to the "edit an existent product" route
func (s *HTTPWebServer) EditProductGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
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
		s.logger.Warn("Product with code '%s' not found", code)
		s.setErrorMessage(c, "Product not found")
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	c.HTML(http.StatusOK, "editProduct.html", gin.H{
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
		"Flash":      flashMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// EditProductPostHandler handles the POST to the "edit an existent product" route
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
	stockInt32 := uint32(stockInt64)

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
	s.setFlashMessage(c, fmt.Sprintf("Successful update of product with code '%s'", code))
	c.Redirect(http.StatusSeeOther, "/app/product/")
}

// DeleteProductHandler handles the POST to the "delete an existent product" route
func (s *HTTPWebServer) DeleteProductHandler(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		s.logger.Warn("Missing code parameter in delete product handler")
		s.setErrorMessage(c, "Missing product code")
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	succ, err := s.orchestrator.RemoveProductFromAllCarts(code) // before removing the product, i try to remove it from all carts
	if err != nil {
		s.logger.Error("Error removing the product with code '%s' from all carts: %v", code, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}
	if !succ {
		s.logger.Warn("Failed removal of product with code '%s' from all carts: no products found in carts", code)
		// it is not a problem, let's continue with the product deletion
	}

	succ, err = s.orchestrator.DeleteProduct(code)
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
	s.setFlashMessage(c, fmt.Sprintf("Successful delete of the product"))
	c.Redirect(http.StatusSeeOther, "/app/product/")
}

// ProductGetHandler handles the GET to the "see the product" route
func (s *HTTPWebServer) ProductGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
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

	var available uint32 = 0

	succ, prod, err := s.orchestrator.GetProduct(code)
	if err != nil {
		s.logger.Error("Error during retrieval of product with code '%s': %v", code, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}
	if !succ {
		s.logger.Warn("Product with code '%s' not found", code)
		s.setErrorMessage(c, "Product not found")
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	available = prod.Stock
	var numberInCart uint32 = 0

	succ, item, err := s.orchestrator.GetItemFromCart(username, code)
	if err != nil {
		s.logger.Warn("Erro while retrieving info about product with code '%s' into the cart of user '%s'", code, username)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}
	if succ {
		numberInCart = item.Quantity
		available = prod.Stock - numberInCart // number of product that can be added to the cart (to be under the stock quantity)
	}

	c.HTML(http.StatusOK, "buyProduct.html", gin.H{
		"Title":        "Buy Product - Ecommerce",
		"Username":     username,
		"Code":         prod.Code,
		"Name":         prod.Name,
		"Size":         prod.Size.String(),
		"Color":        prod.Color,
		"Description":  prod.Description,
		"Stock":        prod.Stock,
		"NumberInCart": numberInCart,
		"Price":        prod.Price,
		"Available":    available, // number of product that can be added to the cart (to be under the stock quantity)
		"Flash":        flashMsg,
		"Error":        errMsg,
		"IsLoggedIn":   isLoggedIn,
	})
}

/*
* CART HANDLERS
 */

// CartItemView utility struct for item fields to be displayed in the cart
type CartItemView struct {
	Code       string
	Name       string
	Price      float64
	Stock      uint32
	Quantity   uint32
	TotalPrice float64
}

// ListCartItemsGetHandler handles a GET to the "show product into cart" route
func (s *HTTPWebServer) ListCartItemsGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	warnMsg := s.getWarnMessage(c)
	itemsNumber := 0
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}
	var totalCartValue float64
	totalCartValue = 0

	succ, items, err := s.orchestrator.GetListOfProductsIntoCart(username)
	if err != nil {
		s.logger.Error("Cart products retrieval error for user '%s': %v", username, err)
		c.HTML(http.StatusInternalServerError, "cart.html", gin.H{
			"Title":       "User Cart - Ecommerce",
			"Total":       totalCartValue,
			"Username":    username,
			"Error":       interalServerErrorMsg,
			"ItemsNumber": itemsNumber,
			"IsLoggedIn":  isLoggedIn,
		})
		return
	}
	if !succ {
		s.logger.Warn("Cart products retrieval failed")
		c.HTML(http.StatusInternalServerError, "cart.html", gin.H{
			"Title":       "User Cart - Ecommerce",
			"Total":       totalCartValue,
			"Username":    username,
			"Warning":     "The cart is empty",
			"ItemsNumber": itemsNumber,
			"IsLoggedIn":  isLoggedIn,
		})
		return
	}

	var cartViewItems []CartItemView
	var itemsReducedQuantity []string

	for _, item := range items {
		succ, prod, err := s.orchestrator.GetProduct(item.ProdCode)
		if err != nil || !succ {
			s.logger.Warn("Failed to get details for product with code '%s'", item.ProdCode)
			s.setErrorMessage(c, interalServerErrorMsg)
			c.Redirect(http.StatusSeeOther, "/app/")
			return
		}
		if item.Quantity > prod.Stock {
			s.logger.Warn("Quantity of items for product with code '%s' is greather than the stock in the cart of user '%s': "+
				"reducing quantity to '%d'", item.ProdCode, username, prod.Stock)
			succ, err := s.orchestrator.UpdateQuantityOfProductIntoCart(username, item.ProdCode, prod.Stock)
			if err != nil {
				s.logger.Error("Error while updating quantity for product '%s' into the cart of user '%s': %v", item.ProdCode, username, err)
				s.setErrorMessage(c, interalServerErrorMsg)
				c.Redirect(http.StatusSeeOther, "/app/")
				return
			}
			if !succ {
				s.logger.Warn("Failed quantity update for product '%s' into the cart of user '%s'", item.ProdCode, username)
				s.setErrorMessage(c, "Quantity update failed")
				c.Redirect(http.StatusSeeOther, "/app/")
				return
			}

			itemsReducedQuantity = append(itemsReducedQuantity, prod.Name)
			s.logger.Info("Successfully reduced quantity of product with code '%s' in the cart of user '%s'", item.ProdCode, username)
			item.Quantity = prod.Stock
		}

		itemsNumber = itemsNumber + 1

		totalItemPrice := (prod.Price * float64(item.Quantity))
		totalCartValue += totalItemPrice

		cartViewItems = append(cartViewItems, CartItemView{
			Code:       item.ProdCode,
			Name:       prod.Name,
			Price:      prod.Price,
			Stock:      prod.Stock,
			Quantity:   item.Quantity,
			TotalPrice: totalItemPrice,
		})
	}

	if len(itemsReducedQuantity) > 0 {
		msg := "Due to stock shortage, the quantity of the following products in your cart has been reduced:\n"
		for _, item := range itemsReducedQuantity {
			msg += "- " + item + "\n"
		}

		warnMsg = msg
	}

	c.HTML(http.StatusOK, "cart.html", gin.H{
		"Title":       "User Cart - Ecommerce",
		"Username":    username,
		"CartItems":   cartViewItems,
		"Total":       totalCartValue,
		"IsLoggedIn":  isLoggedIn,
		"ItemsNumber": itemsNumber,
		"Flash":       flashMsg,
		"Error":       errMsg,
		"Warning":     warnMsg,
	})
}

// AddItemToCartPostHandler handles a POST to the "add product into cart" route
func (s *HTTPWebServer) AddItemToCartPostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}
	code := c.PostForm("code")
	quantity := c.PostForm("quantity")

	quantityInt64, err := strconv.ParseInt(quantity, 10, 32)
	if err != nil {
		s.logger.Warn("Invalid quantity value in add item to cart handler")
		s.setErrorMessage(c, "Invalid quantity value")
		c.Redirect(http.StatusSeeOther, "/app/product/"+code)
		return
	}
	quantityUint32 := uint32(quantityInt64)

	succ, prod, err := s.orchestrator.GetProduct(code) // retrieving the stock
	if err != nil || !succ {
		s.logger.Warn("Failed to get details for product with code '%s'", code)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/"+code)
		return
	}

	totalQuantity := quantityUint32
	succ, item, err := s.orchestrator.GetItemFromCart(username, code) // retrieving the quantity of product already in the cart
	if err != nil {
		s.logger.Error("Error whiile retrieving details for product with code '%s'", code)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/"+code)
		return
	}
	if succ {
		totalQuantity = totalQuantity + item.Quantity
	}

	if totalQuantity > prod.Stock {
		s.logger.Warn("Failed addition of product '%s' to the cart of user '%s': the total quantity is greather than the stock", code, username)
		s.setErrorMessage(c, "Too many items to add: the quantity is greather than the stock")
		c.Redirect(http.StatusSeeOther, "/app/product/"+code)
		return
	}

	succ, err = s.orchestrator.AddProductToCart(username, code, quantityUint32)
	if err != nil {
		s.logger.Error("Error while adding product '%s' to the cart of user '%s': %v", code, username, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}
	if !succ {
		s.logger.Warn("Failed addition of product '%s' to the cart of user '%s'", code, username)
		s.setErrorMessage(c, "Addition to cart failed")
		c.Redirect(http.StatusSeeOther, "/app/product/")
		return
	}

	s.logger.Info("Successful addition of %d products '%s' to the user '%s' cart", quantityUint32, code, username)
	s.setFlashMessage(c, fmt.Sprintf("Successful addition of %d products to the cart", quantityUint32))
	c.Redirect(http.StatusSeeOther, "/app/cart/")
}

// UpdateQuantityItemIntoCartPostHandler handles a POST to the "update number of products into cart" route
func (s *HTTPWebServer) UpdateQuantityItemIntoCartPostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	code := c.Param("code")
	quantity := c.PostForm("quantity")

	quantityInt64, err := strconv.ParseInt(quantity, 10, 32)
	if err != nil {
		s.logger.Warn("Invalid quantity value in update item quantity into cart handler")
		s.setErrorMessage(c, "Invalid quantity value")
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}
	quantityUint32 := uint32(quantityInt64)

	succ, prod, err := s.orchestrator.GetProduct(code) // retrieving the stock
	if err != nil || !succ {
		s.logger.Warn("Failed to get details for product with code '%s'", code)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}

	if quantityUint32 > prod.Stock {
		s.logger.Warn("Failed addition of product '%s' to the cart of user '%s': the total quantity is greather than the stock", code, username)
		s.setErrorMessage(c, "Too many items to add: the quantity is greather than the stock")
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}

	succ, err = s.orchestrator.UpdateQuantityOfProductIntoCart(username, code, quantityUint32)
	if err != nil {
		s.logger.Error("Error while updating quantity for product '%s' into the cart of user '%s': %v", code, username, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}
	if !succ {
		s.logger.Warn("Failed quantity update for product '%s' into the cart of user '%s'", code, username)
		s.setErrorMessage(c, "Quantity update failed")
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}

	s.logger.Info("Successful quantity update to %d products '%s' into the user '%s' cart", quantityUint32, code, username)
	s.setFlashMessage(c, fmt.Sprintf("Successful quantity update to %d products into the cart", quantityUint32))
	c.Redirect(http.StatusSeeOther, "/app/cart/")
}

// RemoveItemFromCartPostHandler handles a POST to the "remove product from cart" route
func (s *HTTPWebServer) RemoveItemFromCartPostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	code := c.Param("code")

	succ, err := s.orchestrator.RemoveProductFromCart(username, code)
	if err != nil {
		s.logger.Error("Error while removing product '%s' from the cart of user '%s': %v", code, username, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}
	if !succ {
		s.logger.Warn("Failed removal of product '%s' from the cart of user '%s'", code, username)
		s.setErrorMessage(c, "Product removal failed")
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}

	s.logger.Info("Successful removal of product '%s' from the user '%s' cart", code, username)
	s.setFlashMessage(c, "Successful removal of all products from the cart")
	c.Redirect(http.StatusSeeOther, "/app/cart/")
}

// ClearCartPostHandler handles a POST to the "remove all products from cart" route
func (s *HTTPWebServer) ClearCartPostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	succ, err := s.orchestrator.RemoveAllProductsFromCart(username)
	if err != nil {
		s.logger.Error("Error while removing all productss from the cart of user '%s': %v", username, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}
	if !succ {
		s.logger.Warn("Failed removal of all products from the cart of user '%s'", username)
		s.setErrorMessage(c, "Cart clear failed")
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}

	s.logger.Info("Successful cart clear for user '%s'", username)
	s.setFlashMessage(c, "Successful removal of all products from the cart")
	c.Redirect(http.StatusSeeOther, "/app/product/")
}

/*
* ORDER HANDLERS
 */

type OrdersListView struct {
	OrderId     int32
	Status      string
	TotalAmount float64
	Items       []OrderItemsView
}

type OrderItemsView struct {
	ProductCode   string
	Name          string
	Price         float64
	Quantity      uint32
	PartialAmount float64
}

// OrdersListByUsernameGetHandler handles a GET to the "view user order list" route
func (s *HTTPWebServer) OrdersListByUsernameGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	succ, orders, err := s.orchestrator.GetOrdersListByUsername(username)
	if err != nil {
		s.logger.Error("Orders retrieval error for user '%s': %v", username, err)
		c.HTML(http.StatusInternalServerError, "orders.html", gin.H{
			"Title":      "User Orders - Ecommerce",
			"Username":   username,
			"IsLoggedIn": isLoggedIn,
			"Error":      interalServerErrorMsg,
		})
		return
	}
	if !succ {
		s.logger.Warn("Orders retrieval failed for user '%s'", username)
		c.HTML(http.StatusInternalServerError, "orders.html", gin.H{
			"Title":      "User Orders - Ecommerce",
			"Username":   username,
			"IsLoggedIn": isLoggedIn,
			"Warning":    "Order list is empty",
		})
		return
	}

	var ordersView []OrdersListView

	for _, order := range orders {
		ordersView = append(ordersView, OrdersListView{
			OrderId:     order.ID,
			Status:      order.Status.String(),
			TotalAmount: order.TotalAmount,
		})
	}

	c.HTML(http.StatusOK, "orders.html", gin.H{
		"Title":      "User Orders - Ecommerce",
		"Username":   username,
		"IsLoggedIn": isLoggedIn,
		"Orders":     ordersView,
		"Flash":      flashMsg,
		"Error":      errMsg,
	})
}

// CreateOrderPostHandler handles a POST on the "create order" route
func (s *HTTPWebServer) CreateOrderPostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	succ, itemsInCart, err := s.orchestrator.GetListOfProductsIntoCart(username)
	if err != nil {
		s.logger.Error("Cart products retrieval error for user '%s': %v", username, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}
	if !succ {
		s.logger.Warn("Cart products retrieval failed")
		s.setErrorMessage(c, "The cart is empty")
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}

	var orderItems []*model.OrderItem
	var isOrderFeasible bool = true // set to false in the for cycle if an item has a quantity greather than the product stock
	var itemsReducedQuantity []string
	stockProd := make(map[string]uint32)

	for _, item := range itemsInCart {
		succ, prod, err := s.orchestrator.GetProduct(item.ProdCode)
		if err != nil {
			s.logger.Error("Error during retrieval of product with code '%s': %v", item.ProdCode, err)
			s.setErrorMessage(c, interalServerErrorMsg)
			c.Redirect(http.StatusSeeOther, "/app/cart/")
			return
		}
		if !succ {
			s.logger.Warn("Product with code '%s' not found", item.ProdCode)
			s.setErrorMessage(c, "Product not found")
			c.Redirect(http.StatusSeeOther, "/app/cart/")
			return
		}
		if item.Quantity > prod.Stock {
			isOrderFeasible = false
			itemsReducedQuantity = append(itemsReducedQuantity, prod.Code)
		}

		stockProd[item.ProdCode] = prod.Stock

		orderItems = append(orderItems, &model.OrderItem{
			ProductCode: item.ProdCode,
			Name:        prod.Name,
			Price:       prod.Price,
			Quantity:    item.Quantity,
		})
	}
	if !isOrderFeasible {
		s.logger.Warn("Impossibility to continue with the order of user '%s' due to lack of products: %v", username, itemsReducedQuantity)
		s.setErrorMessage(c, "Impossible to place the order: product shortage")
		c.Redirect(http.StatusSeeOther, "/app/cart/") // this route will update the quantity of products into the cart to make the order feasible
		return
	}

	for _, item := range itemsInCart {
		newStock := stockProd[item.ProdCode] - item.Quantity
		succ, err := s.orchestrator.UpdateProduct(item.ProdCode, "", model.SizeUnspecified, "", "", newStock, -1) // updates only the stock
		if err != nil || !succ {
			s.logger.Error("Error or failure while updating stock for product '%s'", item.ProdCode)
			s.setErrorMessage(c, interalServerErrorMsg)
			c.Redirect(http.StatusSeeOther, "/app/cart/")
		}
	}

	succ, orderId, err := s.orchestrator.CreateOrder(username, orderItems)
	if err != nil || !succ {
		s.logger.Error("Error or failure while creating the order for user '%s'", username)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}
	s.logger.Info("Successful creation of the order with ID '%d' for user '%s'", orderId, username)
	s.setFlashMessage(c, fmt.Sprintf("Successful creation of the order with ID '%d'", orderId))

	succ, err = s.orchestrator.RemoveAllProductsFromCart(username)
	if err != nil || !succ {
		s.logger.Warn("Impossibility to remove all products form the cart of user '%s' although the order confirmation", username)
		s.setErrorMessage(c, "Impossibility to clear the cart although the successful order")
		c.Redirect(http.StatusSeeOther, "/app/cart/")
		return
	}

	s.logger.Info("Successful removal of all products from the cart of user '%s'", username)
	c.Redirect(http.StatusSeeOther, "/app/order/")
}

// ViewOrderDetailsGetHandler handles a GET on the "view order details" route
func (s *HTTPWebServer) ViewOrderDetailsGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	orderId := c.Param("id")
	orderIdInt, err := strconv.Atoi(orderId)
	if err != nil {
		s.logger.Warn("Invalid order ID in view order details handler")
		s.setErrorMessage(c, "Invalid order ID")
		c.Redirect(http.StatusSeeOther, "/app/order/")
		return
	}

	succ, order, err := s.orchestrator.GetOrder(int32(orderIdInt))
	if err != nil {
		s.logger.Error("Error while retrieving order with ID '%s': %v", orderIdInt, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/order/")
		return
	}
	if !succ {
		s.logger.Warn("Order with ID '%d' not found", orderIdInt)
		s.setErrorMessage(c, "Order not found")
		c.Redirect(http.StatusSeeOther, "/app/order/")
		return
	}

	if order.Username != username {
		s.logger.Info("User '%s' attempted to see details about order with ID '%d'", orderIdInt)
		s.setErrorMessage(c, "You are not authorized to view details about this order")
		c.Redirect(http.StatusSeeOther, "/app/order/")
		return
	}

	var orderItemsView []OrderItemsView
	for _, item := range order.Items {
		var partialAmount float64 = item.Price * float64(item.Quantity)

		orderItemsView = append(orderItemsView, OrderItemsView{
			ProductCode:   item.ProductCode,
			Name:          item.Name,
			Price:         item.Price,
			Quantity:      item.Quantity,
			PartialAmount: partialAmount,
		})
	}

	c.HTML(http.StatusOK, "orderDetails.html", gin.H{
		"Title":      "Order Details - Ecommerce",
		"Username":   username,
		"IsLoggedIn": isLoggedIn,
		"Order": gin.H{
			"OrderId":     order.ID,
			"Status":      order.Status.String(),
			"TotalAmount": order.TotalAmount,
			"Items":       orderItemsView,
		},
		"Flash": flashMsg,
		"Error": errMsg,
	})
}

// CancelOrderPostHandler handles a POST on the "cancel order" route
func (s *HTTPWebServer) CancelOrderPostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	orderId := c.Param("id")
	orderIdInt, err := strconv.Atoi(orderId)
	if err != nil {
		s.logger.Warn("Invalid order ID in cancel order handler")
		s.setErrorMessage(c, "Invalid order ID")
		c.Redirect(http.StatusSeeOther, "/app/order/")
		return
	}
	orderIdInt32 := int32(orderIdInt)

	succ, order, err := s.orchestrator.GetOrder(orderIdInt32)
	if err != nil {
		s.logger.Error("Error while retrieving order with ID '%s': %v", orderIdInt32, err)
		s.setErrorMessage(c, interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/order/")
		return
	}
	if !succ {
		s.logger.Warn("Order with ID '%d' not found", orderIdInt32)
		s.setErrorMessage(c, "Order not found")
		c.Redirect(http.StatusSeeOther, "/app/order/")
		return
	}

	if order.Username != username {
		s.logger.Info("User '%s' attempted to see details about order with ID '%d'", username, orderIdInt32)
		s.setErrorMessage(c, "You are not authorized to view details about this order")
		c.Redirect(http.StatusSeeOther, "/app/order/")
		return
	}

	for _, item := range order.Items {
		found, prod, err := s.orchestrator.GetProduct(item.ProductCode)
		if err != nil {
			s.logger.Error("Error retrieving product with code '%s': %v", item.ProductCode, err)
			s.setErrorMessage(c, interalServerErrorMsg)
			c.Redirect(http.StatusSeeOther, "/app/order/")
			return
		}
		if !found || prod == nil { // item removed from the catalogue
			s.logger.Warn("Product '%s' not found, skipping stock update", item.ProductCode)
			continue
		}

		newStock := prod.Stock + item.Quantity
		succ, err := s.orchestrator.UpdateProduct(item.ProductCode, "", model.SizeUnspecified, "", "", newStock, -1) // updates only the stock
		if err != nil || !succ {
			s.logger.Error("Failed to update stock for product '%s'", item.ProductCode)
			s.setErrorMessage(c, interalServerErrorMsg)
			c.Redirect(http.StatusSeeOther, "/app/order/")
			return
		}
		s.logger.Info("Stock updated successfully for product with code '%s'", item.ProductCode)
	}

	succ, err = s.orchestrator.UpdateOrderStatus(orderIdInt32, model.StatusCanceled)
	if err != nil {
		s.logger.Error("Error while updating the status of order with ID '%d': %v", orderIdInt32, err)
		s.setErrorMessage(c, "Order cancellation failed."+interalServerErrorMsg)
		c.Redirect(http.StatusSeeOther, "/app/order/")
		return
	}
	if !succ {
		s.logger.Warn("Status update failed for order with ID '%d'", orderIdInt32)
		s.setErrorMessage(c, "Order cancellation failed")
		c.Redirect(http.StatusSeeOther, "/app/order/")
		return
	}

	s.logger.Info("Order with ID '%d' successfully canceled", orderIdInt32)
	s.setFlashMessage(c, "Order successfuly canceled")
	c.Redirect(http.StatusSeeOther, "/app/order/")
}

/*
* UTILITIES
 */

// getUsernameFromSessionOrRedirect retrieves the user from the current session.
//
// If the username has not been found, a redirect to the login page is performed
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

// getUserRole retrieves the role of a user
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

// setFlashMessage saves in the session a flash message to be displayed into the HTML
func (s *HTTPWebServer) setFlashMessage(c *gin.Context, message string) {
	session := sessions.Default(c)
	session.Set("flash", message)
	_ = session.Save()
}

// setErrorMessage saves in the session an error message to be displayed into the HTML
func (s *HTTPWebServer) setErrorMessage(c *gin.Context, message string) {
	session := sessions.Default(c)
	session.Set("error", message)
	_ = session.Save()
}

// setWarnMessage saves in the session a warning message to be displayed into the HTML
func (s *HTTPWebServer) setWarnMessage(c *gin.Context, message string) {
	session := sessions.Default(c)
	session.Set("warning", message)
	_ = session.Save()
}

// getFlashMessage retrieves from the session a flash message to be displayed into the HTML
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

// getErrorMessage retrieves from the session an error message to be displayed into the HTML
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

// getWarnMessage retrieves from the session a warning message to be displayed into the HTML
func (s *HTTPWebServer) getWarnMessage(c *gin.Context) string {
	session := sessions.Default(c)
	warnMsg := session.Get("warning")
	if warnMsg != nil {
		session.Delete("warning")
		_ = session.Save()
		return warnMsg.(string)
	}
	return ""
}

func (s *HTTPWebServer) LogErrorAndRedirect(c *gin.Context, logMsg string, userMsg string, redirectPath string, statusCode ...int) {
	s.logger.Error(logMsg)
	s.setErrorMessage(c, userMsg)
	
	code := http.StatusSeeOther
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	
	c.Redirect(code, redirectPath)
}

func (s *HTTPWebServer) LogWarningAndRedirect(c *gin.Context, logMsg string, userMsg string, redirectPath string, statusCode ...int) {
	s.logger.Warn(logMsg)
	s.setWarnMessage(c, userMsg)
	
	code := http.StatusSeeOther
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	
	c.Redirect(code, redirectPath)
}