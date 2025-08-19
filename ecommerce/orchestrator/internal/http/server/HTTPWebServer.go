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
			s.LogWarningAndRedirect(c,
				"Attempted unauthorised access to a page requiring authentication",
				"Please log in to access this page",
				"/login",
				http.StatusFound,
			)
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
			s.LogErrorAndRedirect(c,
				fmt.Sprintf("User retrieval error in AdminRequired: %v", err),
				interalServerErrorMsg,
				"/app/",
				http.StatusFound,
			)
			c.Abort()
			return
		}
		if !succ {
			s.LogWarningAndRedirect(c,
				"User retrieval failed in AdminRequired: User not found",
				"User not found",
				"/app/",
				http.StatusFound,
			)
			c.Abort()
			return
		}

		if user.Role != model.RoleAdmin {
			s.LogWarningAndRedirect(c,
				fmt.Sprintf("Access forbidden: user '%s' tried to access an admin route", username),
				"Access forbidden: Admins only",
				"/app/",
				http.StatusFound,
			)
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
	warnMsg := s.getWarnMessage(c)
	errMsg := s.getErrorMessage(c)
	pageTitle := "Welcome - Ecommerce"
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
			"Title":      pageTitle,
			"Username":   username,
			"IsLoggedIn": isLoggedIn,
			"Error":      interalServerErrorMsg,
		})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title":      pageTitle,
		"Username":   username,
		"Role":       roleStr,
		"IsLoggedIn": isLoggedIn,
		"Flash":      flashMsg,
		"Warning":    warnMsg,
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
	warnMsg := s.getWarnMessage(c)
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
		"Warning":    warnMsg,
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
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Login error: %v", err),
			interalServerErrorMsg,
			"/login",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			"Login failed: Invalid username or password",
			"Invalid username or password",
			"/login",
			http.StatusSeeOther,
		)
		return
	}

	session.Set("username", username)
	err = session.Save()
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Session save error: %v", err),
			interalServerErrorMsg,
			"/login",
			http.StatusSeeOther,
		)
		return
	}

	var flash_role_msg string
	if role.String() == model.RoleAdmin.String() {
		flash_role_msg = "You are an administrator!"
	} else {
		flash_role_msg = "You are a client!"
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Successful login for user '%s'", username),
		fmt.Sprintf("Successful login! %s", flash_role_msg),
		"/app/",
		http.StatusSeeOther,
	)
}

// RegisterGetHandler handles the GET to the register route.
//
// If the user is already authenticated, it is redirected to the main page of the website
func (s *HTTPWebServer) RegisterGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	errMsg := s.getErrorMessage(c)
	warnMsg := s.getWarnMessage(c)
	isLoggedIn := false
	session := sessions.Default(c)
	if session.Get("username") != nil {
		c.Redirect(http.StatusFound, "/app/")
		return
	}
	c.HTML(http.StatusOK, "register.html", gin.H{
		"Title":      "Register - Ecommerce",
		"Flash":      flashMsg,
		"Warning":    warnMsg,
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
		s.LogWarningAndRedirect(c,
			"Password confirmation doesn’t match the password in register request",
			"Password confirmation doesn’t match the password",
			"/register",
			http.StatusSeeOther,
		)
		return
	}

	succ, err := s.orchestrator.Register(username, password, email, phone)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Registration error: %v", err),
			interalServerErrorMsg,
			"/register",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			"Registration failed: Username or email already used",
			"Username or email already used",
			"/register",
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Successful registration for user '%s'", username),
		"Registration successful. You can now log in!",
		"/login",
		http.StatusSeeOther,
	)
}

// ChangePasswordGetHandler handles the GET to the change password route
func (s *HTTPWebServer) ChangePasswordGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	warnMsg := s.getWarnMessage(c)
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
		"Warning":    warnMsg,
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
		s.LogWarningAndRedirect(c,
			"Password confirmation doesn’t match the new password in change password request",
			"Password confirmation doesn’t match the new password",
			"/app/user/password",
			http.StatusSeeOther,
		)
		return
	}

	succ, err := s.orchestrator.ChangePassword(username, currentPassword, newPassword)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Password change error: %v", err),
			interalServerErrorMsg,
			"/app/user/password",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			"Password change failed: Invalid current password",
			"Invalid current password",
			"/app/user/password",
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Password changed successfully for user '%s'", username),
		"Password changed successfully",
		"/app/",
		http.StatusSeeOther,
	)
}

// UserProfileGetHandler handles the GET to the user profile route
func (s *HTTPWebServer) UserProfileGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	warnMsg := s.getWarnMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	pageTitle := "User Profile - Ecommerce"
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}
	succ, user, err := s.orchestrator.GetUser(username)
	if err != nil {
		s.logger.Error("User retrieval error: %v", err)
		c.HTML(http.StatusInternalServerError, "profile.html", gin.H{
			"Title":      pageTitle,
			"Username":   username,
			"Error":      interalServerErrorMsg,
			"IsLoggedIn": isLoggedIn,
		})
		return
	}
	if !succ {
		s.logger.Warn("User retrieval failed: User not found")
		c.HTML(http.StatusNotFound, "profile.html", gin.H{
			"Title":      pageTitle,
			"Username":   username,
			"Warning":    "User not found",
			"IsLoggedIn": isLoggedIn,
		})
		return
	}

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"Title":      pageTitle,
		"Username":   username,
		"Email":      user.Email,
		"Phone":      user.Phone,
		"Role":       user.Role.String(),
		"Flash":      flashMsg,
		"Warning":    warnMsg,
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
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("User update error: %v", err),
			interalServerErrorMsg,
			"/app/user/profile",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			"User update failed",
			"User update failed",
			"/app/user/profile",
			http.StatusSeeOther,
		)
		return
	}
	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Profile successfully updated for user '%s'", username),
		"User updated successfully",
		"/app/user/profile",
		http.StatusSeeOther,
	)
}

// UsersGetHandler handles the GET to the "see list of users" route
func (s *HTTPWebServer) UsersGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	warnMsg := s.getWarnMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	pageTitle := "Manage Users - Ecommerce"
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	succ, users, err := s.orchestrator.GetUsers()
	if err != nil {
		s.logger.Error("Error retrieving users: %v", err)
		c.HTML(http.StatusInternalServerError, "adminUsers.html", gin.H{
			"Title":      pageTitle,
			"Username":   username,
			"Error":      interalServerErrorMsg,
			"IsLoggedIn": isLoggedIn,
		})
		return
	}
	if !succ {
		s.logger.Warn("No users found")
		c.HTML(http.StatusInternalServerError, "adminUsers.html", gin.H{
			"Title":      pageTitle,
			"Username":   username,
			"Warning":    "No users found",
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
		"Title":      pageTitle,
		"Users":      usersForTemplate,
		"Username":   username,
		"Flash":      flashMsg,
		"Warning":    warnMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// SetUserRolePostHandler handle the POST to the set user role route
func (s *HTTPWebServer) SetUserRolePostHandler(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		s.LogWarningAndRedirect(c,
			"Username parameter missing in the set user role handler",
			"Missing username parameter",
			"/app/admin/users",
			http.StatusSeeOther,
		)
		return
	}

	newRole := c.PostForm("role")

	var role model.Role
	role, ok := model.RoleMapStrToRole[newRole]
	if !ok {
		role = model.RoleUnspecified
	}

	if role == model.RoleUnspecified {
		s.LogWarningAndRedirect(c,
			"Invalid role specified in set user role request",
			"Invalid role specified",
			"/app/admin/users",
			http.StatusSeeOther,
		)
		return
	}

	succ, err := s.orchestrator.SetUserRole(username, role)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error during role update for user %s: %v", username, err),
			interalServerErrorMsg,
			"/app/admin/users",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Role update failed for user %s", username),
			"Role setting failed",
			"/app/admin/users",
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Role successfully updated for user '%s'", username),
		fmt.Sprintf("Role successfully updated for user '%s'", username),
		"/app/admin/users",
		http.StatusSeeOther,
	)
}

// LogoutHandler handles the GET to the logout route
func (s *HTTPWebServer) LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		s.logger.Error("Session save error on logout: %v", err)
		s.setWarnMessage(c, "There was an issue logging out, but you have been logged out")
	} else {
		s.setFlashMessage(c, "You have been logged out!")
	}

	c.Redirect(http.StatusFound, "/login")
}

/*
* PRODUCT HANDLERS
 */

// ListProductsGetHandler handles the GET to the "list all products" route
func (s *HTTPWebServer) ListProductsGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	warnMsg := s.getWarnMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	pageTitle := "Product Catalogue - Ecommerce"
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	role, err := s.getUserRole(username)
	roleStr, ok := model.RoleMapRoleToStr[role]
	if err != nil || !ok {
		s.logger.Error("Error fetching user role or unknown role")
		c.HTML(http.StatusInternalServerError, "productsList.html", gin.H{
			"Title":      pageTitle,
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
			"Title":      pageTitle,
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
			"Title":      pageTitle,
			"Username":   username,
			"Role":       roleStr,
			"Warning":    "No products found",
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
		"Title":      pageTitle,
		"Products":   prodsForTemplate,
		"Username":   username,
		"Role":       roleStr,
		"Flash":      flashMsg,
		"Warning":    warnMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// NewProductGetHandler handles the GET to the "create a new product" route
func (s *HTTPWebServer) NewProductGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	warnMsg := s.getWarnMessage(c)
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
		"Warning":    warnMsg,
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
		s.LogWarningAndRedirect(c,
			"Invalid size specified in new product handler",
			"Invalid size specified",
			"/app/product/new",
			http.StatusSeeOther,
		)
		return
	}

	stockInt64, err := strconv.ParseInt(stock, 10, 32)
	if err != nil {
		s.LogWarningAndRedirect(c,
			"Invalid stock value in new product handler",
			"Invalid stock value",
			"/app/product/new",
			http.StatusSeeOther,
		)
		return
	}
	stockInt32 := uint32(stockInt64)

	priceFloat64, err := strconv.ParseFloat(price, 64)
	if err != nil {
		s.LogWarningAndRedirect(c,
			"Invalid price value in new product handler",
			"Invalid price value",
			"/app/product/new",
			http.StatusSeeOther,
		)
		return
	}

	succ, err := s.orchestrator.CreateProduct(code, name, sizeModel, color, description, stockInt32, priceFloat64)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error creating the new product with code '%s': %v", code, err),
			interalServerErrorMsg,
			"/app/product/new",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed to create product with code '%s'", code),
			fmt.Sprintf("Product creation failed: already exists a product with the same code '%s'", code),
			"/app/product/new",
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Product with code '%s' successfully created", code),
		fmt.Sprintf("Product with code '%s' successfully created", code),
		"/app/product/",
		http.StatusSeeOther,
	)
}

// EditProductGetHandler handles the GET to the "edit an existent product" route
func (s *HTTPWebServer) EditProductGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	warnMsg := s.getWarnMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	code := c.Param("code")
	if code == "" {
		s.LogWarningAndRedirect(c,
			"Missing code parameter in edit product handler",
			"Product code missing",
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}

	succ, prod, err := s.orchestrator.GetProduct(code)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error retrieving product with code '%s': %v", code, err),
			interalServerErrorMsg,
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Product with code '%s' not found", code),
			"Product not found",
			"/app/product/",
			http.StatusSeeOther,
		)
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
		"Warning":    warnMsg,
		"Error":      errMsg,
		"IsLoggedIn": isLoggedIn,
	})
}

// EditProductPostHandler handles the POST to the "edit an existent product" route
func (s *HTTPWebServer) EditProductPostHandler(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		s.LogWarningAndRedirect(c,
			"Missing code parameter in edit product handler",
			"Product code missing",
			"/app/product/",
			http.StatusSeeOther,
		)
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
		s.LogWarningAndRedirect(c,
			"Invalid size specified in edit product handler",
			"Invalid size specified",
			fmt.Sprintf("/app/product/%s/edit", code),
			http.StatusSeeOther,
		)
		return
	}

	stockInt64, err := strconv.ParseInt(stock, 10, 32)
	if err != nil {
		s.LogWarningAndRedirect(c,
			"Invalid stock value in edit product handler",
			"Invalid stock value",
			fmt.Sprintf("/app/product/%s/edit", code),
			http.StatusSeeOther,
		)
		return
	}
	stockInt32 := uint32(stockInt64)

	priceFloat64, err := strconv.ParseFloat(price, 64)
	if err != nil {
		s.LogWarningAndRedirect(c,
			"Invalid price value in edit product handler",
			"Invalid price value",
			fmt.Sprintf("/app/product/%s/edit", code),
			http.StatusSeeOther,
		)
		return
	}

	succ, err := s.orchestrator.UpdateProduct(code, name, sizeModel, color, description, stockInt32, priceFloat64)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error while updating product with code '%s': %v", code, err),
			interalServerErrorMsg,
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Update unsuccessful for product with code '%s'", code),
			"Product update failed: product not found",
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Successful update of product with code '%s'", code),
		fmt.Sprintf("Successful update of product with code '%s'", code),
		"/app/product/",
		http.StatusSeeOther,
	)
}

// DeleteProductHandler handles the POST to the "delete an existent product" route
func (s *HTTPWebServer) DeleteProductHandler(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		s.LogWarningAndRedirect(c,
			"Missing code parameter in delete product handler",
			"Missing product code",
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}

	succ, err := s.orchestrator.RemoveProductFromAllCarts(code) // before removing the product, we try to remove it from all carts
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error removing the product with code '%s' from all carts: %v", code, err),
			interalServerErrorMsg,
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.logger.Warn("Failed removal of product with code '%s' from all carts: no cart containing the product was found.", code)
		// it is not a problem, let's continue with the product deletion
	}

	succ, err = s.orchestrator.DeleteProduct(code)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error deleting the product with code '%s': %v", code, err),
			interalServerErrorMsg,
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed deletion of product with code '%s'", code),
			"Product deletion failed: product not found",
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Successful deletion of product with code '%s'", code),
		"Successful deletion of the product",
		"/app/product/",
		http.StatusSeeOther,
	)
}

// ProductGetHandler handles the GET to the "see the product" route
func (s *HTTPWebServer) ProductGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	warnMsg := s.getWarnMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	code := c.Param("code")
	if code == "" {
		s.LogWarningAndRedirect(c,
			"Missing code parameter in get product handler",
			"Missing product code",
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}

	var available uint32 = 0

	succ, prod, err := s.orchestrator.GetProduct(code)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error retrieving product with code '%s': %v", code, err),
			interalServerErrorMsg,
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Product with code '%s' not found", code),
			"Product not found",
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}

	available = prod.Stock
	var numberInCart uint32 = 0

	succ, item, err := s.orchestrator.GetItemFromCart(username, code)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error retrieving information of product with code '%s' into the cart of user '%s'", code, username),
			interalServerErrorMsg,
			"/app/product/",
			http.StatusSeeOther,
		)
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
		"Warning":      warnMsg,
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
	warnMsg := s.getWarnMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	pageTitle := "User Cart - Ecommerce"
	itemsNumber := 0
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}
	var totalCartValue float64
	totalCartValue = 0

	succ, items, err := s.orchestrator.GetListOfProductsIntoCart(username)
	if err != nil {
		s.logger.Error("Error retrieving cart products for user '%s': %v", username, err)
		c.HTML(http.StatusInternalServerError, "cart.html", gin.H{
			"Title":       pageTitle,
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
			"Title":       pageTitle,
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
		if err != nil {
			s.LogErrorAndRedirect(c,
				fmt.Sprintf("Error retrieving information of product with code '%s'", item.ProdCode),
				interalServerErrorMsg,
				"/app/",
				http.StatusSeeOther,
			)
			return
		}
		if !succ {
			s.LogWarningAndRedirect(c,
				fmt.Sprintf("Failed information retrieval for product with code '%s'", item.ProdCode),
				"Failed product retrieval: product not found",
				"/app/",
				http.StatusSeeOther,
			)
			return
		}
		if item.Quantity > prod.Stock {
			s.logger.Warn("Quantity of items for product with code '%s' is greather than the stock in the cart of user '%s': "+
				"reducing quantity to '%d'", item.ProdCode, username, prod.Stock)
			succ, err := s.orchestrator.UpdateQuantityOfProductIntoCart(username, item.ProdCode, prod.Stock)
			if err != nil {
				s.LogErrorAndRedirect(c,
					fmt.Sprintf("Error updating quantity for product with code '%s' in the cart of user '%s': %v", item.ProdCode, username, err),
					interalServerErrorMsg,
					"/app/",
					http.StatusSeeOther,
				)
				return
			}
			if !succ {
				s.LogWarningAndRedirect(c,
					fmt.Sprintf("Failed quantity update for product with code '%s' in the cart of user '%s'", item.ProdCode, username),
					"Quantity update failed",
					"/app/",
					http.StatusSeeOther,
				)
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
		"Title":       pageTitle,
		"Username":    username,
		"CartItems":   cartViewItems,
		"Total":       totalCartValue,
		"IsLoggedIn":  isLoggedIn,
		"ItemsNumber": itemsNumber,
		"Flash":       flashMsg,
		"Warning":     warnMsg,
		"Error":       errMsg,
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
		s.LogWarningAndRedirect(c,
			"Invalid quantity value in add item to cart handler",
			"Invalid quantity value",
			fmt.Sprintf("/app/product/%s", code),
			http.StatusSeeOther,
		)
		return
	}
	quantityUint32 := uint32(quantityInt64)

	succ, prod, err := s.orchestrator.GetProduct(code) // retrieving the stock
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error retrieving information of product with code '%s'", code),
			interalServerErrorMsg,
			fmt.Sprintf("/app/product/%s", code),
			http.StatusSeeOther,
		)
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed information retrieval for product with code '%s'", code),
			"Failed product retrieval: product not found",
			fmt.Sprintf("/app/product/%s", code),
			http.StatusSeeOther,
		)
		return
	}

	totalQuantity := quantityUint32
	succ, item, err := s.orchestrator.GetItemFromCart(username, code) // retrieving the quantity of product already in the cart
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error retrieving information for product with code '%s'", code),
			interalServerErrorMsg,
			fmt.Sprintf("/app/product/%s", code),
			http.StatusSeeOther,
		)
		return
	}
	if succ {
		totalQuantity = totalQuantity + item.Quantity
	}

	if totalQuantity > prod.Stock {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed to add product with code '%s' to cart of user '%s': the total quantity is greather than the stock", code, username),
			"Too many items to add: the quantity is greather than the stock",
			fmt.Sprintf("/app/product/%s", code),
			http.StatusSeeOther,
		)
		return
	}

	succ, err = s.orchestrator.AddProductToCart(username, code, quantityUint32)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error adding product '%s' to the cart of user '%s': %v", code, username, err),
			interalServerErrorMsg,
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed to add product with code '%s' to the cart of user '%s'", code, username),
			"Addition to cart failed",
			"/app/product/",
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Successful addition of %d products '%s' to the user '%s' cart", quantityUint32, code, username),
		fmt.Sprintf("Successful addition of %d products to the cart", quantityUint32),
		"/app/cart/",
		http.StatusSeeOther,
	)
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
		s.LogWarningAndRedirect(c,
			"Invalid quantity value in update item quantity into cart handler",
			"Invalid quantity value",
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}
	quantityUint32 := uint32(quantityInt64)

	succ, prod, err := s.orchestrator.GetProduct(code) // retrieving the stock
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error retrieving information of product with code '%s'", code),
			interalServerErrorMsg,
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed information retrieval for product with code '%s'", code),
			"Failed product retrieval: product not found",
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}

	if quantityUint32 > prod.Stock {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed addition of product with code '%s' to the cart of user '%s': the total quantity is greather than the stock", code, username),
			"Too many items to add: the quantity is greather than the stock",
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}

	succ, err = s.orchestrator.UpdateQuantityOfProductIntoCart(username, code, quantityUint32)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error updating quantity for product '%s' into the cart of user '%s': %v", code, username, err),
			interalServerErrorMsg,
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed quantity update for product '%s' into the cart of user '%s'", code, username),
			"Quantity update failed",
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Successful quantity update to %d, for product with code '%s' into the cart of user '%s'", quantityUint32, code, username),
		fmt.Sprintf("Successful quantity update to %d products into the cart", quantityUint32),
		"/app/cart/",
		http.StatusSeeOther,
	)
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
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error removing product with code '%s' from the cart of user '%s': %v", code, username, err),
			interalServerErrorMsg,
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed removal of product with code '%s' from the cart of user '%s'", code, username),
			"Product removal failed",
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Successful removal of product with code '%s' from the user '%s' cart", code, username),
		"Successful removal of the product from the cart",
		"/app/cart/",
		http.StatusSeeOther,
	)
}

// ClearCartPostHandler handles a POST to the "remove all products from cart" route
func (s *HTTPWebServer) ClearCartPostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	succ, err := s.orchestrator.RemoveAllProductsFromCart(username)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error removing all products from the cart of user '%s': %v", username, err),
			interalServerErrorMsg,
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed removal of all products from the cart of user '%s'", username),
			"Cart clearing failed",
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Successful cart clear for user '%s'", username),
		"Successful removal of all products from the cart",
		"/app/product/",
		http.StatusSeeOther,
	)
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
	warnMsg := s.getWarnMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	pageTitle := "User Orders - Ecommerce"
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	succ, orders, err := s.orchestrator.GetOrdersListByUsername(username)
	if err != nil {
		s.logger.Error("Orders retrieval error for user '%s': %v", username, err)
		c.HTML(http.StatusInternalServerError, "orders.html", gin.H{
			"Title":      pageTitle,
			"Username":   username,
			"IsLoggedIn": isLoggedIn,
			"Error":      interalServerErrorMsg,
		})
		return
	}
	if !succ {
		s.logger.Warn("Orders retrieval failed for user '%s'", username)
		c.HTML(http.StatusInternalServerError, "orders.html", gin.H{
			"Title":      pageTitle,
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
		"Title":      pageTitle,
		"Username":   username,
		"IsLoggedIn": isLoggedIn,
		"Orders":     ordersView,
		"Flash":      flashMsg,
		"Warning":    warnMsg,
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
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error retrieving products from the cart of user '%s': %v", username, err),
			interalServerErrorMsg,
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed retrieval of products from the cart of user '%s'", username),
			"The cart is empty",
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}

	var orderItems []*model.OrderItem
	var isOrderFeasible bool = true // set to false in the for cycle if an item has a quantity greather than the product stock
	var itemsReducedQuantity []string
	stockProd := make(map[string]uint32)

	for _, item := range itemsInCart {
		succ, prod, err := s.orchestrator.GetProduct(item.ProdCode)
		if err != nil {
			s.LogErrorAndRedirect(c,
				fmt.Sprintf("Error retrieving product with code '%s': %v", item.ProdCode, err),
				interalServerErrorMsg,
				"/app/cart/",
				http.StatusSeeOther,
			)
			return
		}
		if !succ {
			s.LogWarningAndRedirect(c,
				fmt.Sprintf("Failed information retrieval for product with code '%s'", item.ProdCode),
				"Product not found",
				"/app/cart/",
				http.StatusSeeOther,
			)
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
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Unable to proceed with the order of user '%s' due to lack of products: %v", username, itemsReducedQuantity),
			"Unable to place the order: product shortage",
			"/app/cart/", // this route will update the quantity of products into the cart to make the order feasible
			http.StatusSeeOther,
		)
		return
	}

	for _, item := range itemsInCart {
		newStock := stockProd[item.ProdCode] - item.Quantity
		succ, err := s.orchestrator.UpdateProduct(item.ProdCode, "", model.SizeUnspecified, "", "", newStock, -1) // updates only the stock
		if err != nil {
			s.LogErrorAndRedirect(c,
				fmt.Sprintf("Error updating stock for product with code '%s': %v", item.ProdCode, err),
				interalServerErrorMsg,
				"/app/cart/",
				http.StatusSeeOther,
			)
			return
		}
		if !succ {
			s.LogWarningAndRedirect(c,
				fmt.Sprintf("Failed to update stock for product with code '%s'", item.ProdCode),
				"Failed product update: Product not found",
				"/app/cart/",
				http.StatusSeeOther,
			)
			return
		}
	}

	succ, orderId, err := s.orchestrator.CreateOrder(username, orderItems)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error creating order for user '%s': %v", username, err),
			interalServerErrorMsg,
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed to create order for user '%s'", username),
			"Order creation failed",
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}
	s.logger.Info("Successful creation of the order with ID '%d' for user '%s'", orderId, username)
	s.setFlashMessage(c, fmt.Sprintf("Successful creation of the order with ID '%d'", orderId))

	succ, err = s.orchestrator.RemoveAllProductsFromCart(username)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error clearing the cart of user '%s', although the successful order: %v", username, err),
			interalServerErrorMsg,
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed to clear the cart of user '%s', although the successful order: %v", username),
			"Unable to clear the cart although the successful order",
			"/app/cart/",
			http.StatusSeeOther,
		)
		return
	}

	s.logger.Info("Successful removal of all products from the cart of user '%s'", username)
	c.Redirect(http.StatusSeeOther, "/app/order/")
}

// ViewOrderDetailsGetHandler handles a GET on the "view order details" route
func (s *HTTPWebServer) ViewOrderDetailsGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	warnMsg := s.getWarnMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	orderId := c.Param("id")
	orderIdInt, err := strconv.Atoi(orderId)
	if err != nil {
		s.LogWarningAndRedirect(c,
			"Invalid order ID in view order details handler",
			"Invalid order ID",
			"/app/order/",
			http.StatusSeeOther,
		)
		return
	}

	role, err := s.getUserRole(username)
	roleStr, ok := model.RoleMapRoleToStr[role]
	if err != nil || !ok {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error fetching user role or unknown role"),
			interalServerErrorMsg,
			"/app/order/",
			http.StatusSeeOther,
		)
		return
	}

	succ, order, err := s.orchestrator.GetOrder(int32(orderIdInt))
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error retrieving order with ID '%s': %v", orderIdInt, err),
			interalServerErrorMsg,
			"/app/order/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed to retrieve order with ID '%d': order not found", orderIdInt),
			"Order not found",
			"/app/order/",
			http.StatusSeeOther,
		)
		return
	}

	if order.Username != username && role != model.RoleAdmin {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("User '%s' attempted to see details about order with ID '%d'", username, orderIdInt),
			"Authorisation denied: you are not authorised to access the order details",
			"/app/",
			http.StatusSeeOther,
		)
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
		"Role":       roleStr,
		"IsLoggedIn": isLoggedIn,
		"Order": gin.H{
			"OrderId":     order.ID,
			"Username":    order.Username,
			"Status":      order.Status.String(),
			"TotalAmount": order.TotalAmount,
			"Items":       orderItemsView,
		},
		"Flash":   flashMsg,
		"Warning": warnMsg,
		"Error":   errMsg,
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
		s.LogWarningAndRedirect(c,
			"Invalid order ID in cancel order handler",
			"Invalid order ID",
			"/app/order/",
			http.StatusSeeOther,
		)
		return
	}
	orderIdInt32 := int32(orderIdInt)

	succ, order, err := s.orchestrator.GetOrder(orderIdInt32)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error retrieving order with ID '%s': %v", orderIdInt32, err),
			interalServerErrorMsg,
			"/app/order/",
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed to retrieve order with ID '%d': order not found", orderIdInt32),
			"Order not found",
			"/app/order/",
			http.StatusSeeOther,
		)
		return
	}

	if order.Username != username {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("User '%s' attempted to see details about order with ID '%d'", username, orderIdInt32),
			"Authorisation denied: you are not authorised to access the order details",
			"/app/",
			http.StatusSeeOther,
		)
		return
	}

	succ, err = s.orchestrator.UpdateOrderStatus(orderIdInt32, model.StatusCanceled)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error updating the status of order with ID '%d': %v", orderIdInt32, err),
			interalServerErrorMsg,
			fmt.Sprintf("/app/order/%s", orderId),
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed status update for order with ID '%d'", orderIdInt32),
			"Order cancellation failed",
			fmt.Sprintf("/app/order/%s", orderId),
			http.StatusSeeOther,
		)
		return
	}

	for _, item := range order.Items {
		found, prod, err := s.orchestrator.GetProduct(item.ProductCode)
		if err != nil {
			s.LogErrorAndRedirect(c,
				fmt.Sprintf("Error retrieving product with code '%s': %v", item.ProductCode, err),
				interalServerErrorMsg,
				fmt.Sprintf("/app/order/%s", orderId),
				http.StatusSeeOther,
			)
			return
		}
		if !found || prod == nil { // item removed from the catalogue: the stock update will not be performed
			s.logger.Warn("Product '%s' not found, skipping stock update", item.ProductCode)
			continue
		}

		newStock := prod.Stock + item.Quantity
		succ, err := s.orchestrator.UpdateProduct(item.ProductCode, "", model.SizeUnspecified, "", "", newStock, -1) // updates only the stock
		if err != nil {
			s.LogErrorAndRedirect(c,
				fmt.Sprintf("Error updating stock of product '%s'", item.ProductCode),
				interalServerErrorMsg,
				fmt.Sprintf("/app/order/%s", orderId),
				http.StatusSeeOther,
			)
			return
		}
		if !succ {
			s.LogWarningAndRedirect(c,
				fmt.Sprintf("Failed to update stock for product '%s'", item.ProductCode),
				"Product update failed: product not found",
				fmt.Sprintf("/app/order/%s", orderId),
				http.StatusSeeOther,
			)
			return
		}
		s.logger.Info("Stock updated successfully for product with code '%s'", item.ProductCode)
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Order with ID '%d' successfully canceled", orderIdInt32),
		"Order successfuly canceled",
		fmt.Sprintf("/app/order/%s", orderId),
		http.StatusSeeOther,
	)
}

// OrdersListAllGetHandler handles a GET ont the "view all orders" route (admins only)
func (s *HTTPWebServer) OrdersListAllGetHandler(c *gin.Context) {
	flashMsg := s.getFlashMessage(c)
	warnMsg := s.getWarnMessage(c)
	errMsg := s.getErrorMessage(c)
	isLoggedIn := true
	pageTitle := "Order List Admin - Ecommerce"
	userRole := model.RoleAdmin.String() // to enter this route, the middleware AdminRequired() is called; thus, the user who arrives here is an administrator
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	succ, orders, err := s.orchestrator.GetAllOrdersList()
	if err != nil {
		s.logger.Error("Orders retrieval error: %v", username, err)
		c.HTML(http.StatusInternalServerError, "orders.html", gin.H{
			"Title":      pageTitle,
			"Username":   username,
			"Role":       userRole,
			"IsLoggedIn": isLoggedIn,
			"Error":      interalServerErrorMsg,
		})
		return
	}
	if !succ {
		s.logger.Warn("Orders retrieval failed for user '%s'", username)
		c.HTML(http.StatusInternalServerError, "orders.html", gin.H{
			"Title":      pageTitle,
			"Username":   username,
			"Role":       userRole,
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
		"Title":      pageTitle,
		"Username":   username,
		"Role":       userRole,
		"IsLoggedIn": isLoggedIn,
		"Orders":     ordersView,
		"Flash":      flashMsg,
		"Warning":    warnMsg,
		"Error":      errMsg,
	})
}

// UpdateOrderStatusPostHandler handles a POST on the "update order status" route
func (s *HTTPWebServer) UpdateOrderStatusPostHandler(c *gin.Context) {
	username, ok := s.getUsernameFromSessionOrRedirect(c)
	if !ok {
		return // method getUsernameFromSessionOrRedirect already did the redirect
	}

	orderId := c.Param("id")
	orderIdInt, err := strconv.Atoi(orderId)
	if err != nil {
		s.LogWarningAndRedirect(c,
			"Invalid order ID in cancel order handler",
			"Invalid order ID",
			"/app/order/all",
			http.StatusSeeOther,
		)
		return
	}
	orderIdInt32 := int32(orderIdInt)

	strStatus := c.PostForm("status")

	newStatus, ok := model.StatusMapStrToStatus[strStatus]
	if !ok {
		s.LogWarningAndRedirect(c,
			"Unknown order status in update order status handler",
			"Unknown order status",
			fmt.Sprintf("/app/order/%d", orderIdInt32),
			http.StatusSeeOther,
		)
		return
	}

	succ, err := s.orchestrator.UpdateOrderStatus(orderIdInt32, newStatus)
	if err != nil {
		s.LogErrorAndRedirect(c,
			fmt.Sprintf("Error updating the status of order with ID '%d': %v", orderIdInt32, err),
			interalServerErrorMsg,
			fmt.Sprintf("/app/order/%d", orderIdInt32),
			http.StatusSeeOther,
		)
		return
	}
	if !succ {
		s.LogWarningAndRedirect(c,
			fmt.Sprintf("Failed status update for order with ID '%d'", orderIdInt32),
			"Order status update failed",
			fmt.Sprintf("/app/order/%d", orderIdInt32),
			http.StatusSeeOther,
		)
		return
	}

	s.LogInfoAndRedirect(c,
		fmt.Sprintf("Status of order with ID '%d' successfully updated by user '%s'", orderIdInt32, username),
		"Order status successfuly updated",
		fmt.Sprintf("/app/order/%d", orderIdInt32),
		http.StatusSeeOther,
	)
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

// setWarnMessage saves in the session a warning message to be displayed into the HTML
func (s *HTTPWebServer) setWarnMessage(c *gin.Context, message string) {
	session := sessions.Default(c)
	session.Set("warning", message)
	_ = session.Save()
}

// setErrorMessage saves in the session an error message to be displayed into the HTML
func (s *HTTPWebServer) setErrorMessage(c *gin.Context, message string) {
	session := sessions.Default(c)
	session.Set("error", message)
	_ = session.Save()
}

// getFlashMessage retrieves from the session a flash message to be displayed into the HTML
func (s *HTTPWebServer) getFlashMessage(c *gin.Context) string {
	session := sessions.Default(c)
	flashMsg := session.Get("flash")
	if flashMsg != nil {
		session.Delete("flash")
		_ = session.Save()
		return flashMsg.(string)
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

// LogInfoAndRedirect logs the message, with logger.Level [INFO], sets the flash message and performs the redirect
func (s *HTTPWebServer) LogInfoAndRedirect(c *gin.Context, logMsg string, userMsg string, redirectPath string, statusCode int) {
	s.logger.Info(logMsg)
	s.setFlashMessage(c, userMsg)
	c.Redirect(statusCode, redirectPath)
}

// LogWarningAndRedirect logs the message, with logger.Level [WARN], sets the warning message and performs the redirect
func (s *HTTPWebServer) LogWarningAndRedirect(c *gin.Context, logMsg string, userMsg string, redirectPath string, statusCode int) {
	s.logger.Warn(logMsg)
	s.setWarnMessage(c, userMsg)
	c.Redirect(statusCode, redirectPath)
}

// LogErrorAndRedirect logs the message, with logger.Level [ERROR], sets the error message and performs the redirect
func (s *HTTPWebServer) LogErrorAndRedirect(c *gin.Context, logMsg string, userMsg string, redirectPath string, statusCode int) {
	s.logger.Error(logMsg)
	s.setErrorMessage(c, userMsg)
	c.Redirect(statusCode, redirectPath)
}
