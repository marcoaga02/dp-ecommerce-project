package repository

import (
	"errors"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal/model"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GormAuthRepository implements the AuthServiceInterface interface using GORM as the ORM layer.
type GormAuthRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewGormAuthRepository creates a new instance of GormAuthRepository.
func NewGormAuthRepository(db *gorm.DB, logger logger.Logger) *GormAuthRepository {
	return &GormAuthRepository{db: db, logger: logger}
}

// Login verifies username and password, returns user info if successful.
func (r *GormAuthRepository) Login(username, password string) (bool, *pb.User, error) {
	userModel, err := r.getUserByUsername(username)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("User %s not found", username)
			return false, nil, nil
		}
		r.logger.Error("Error retrieving user with username '%s': %v", username, err)
		return false, nil, err
	}

	if !r.CheckPasswordHash(password, userModel.PasswordHash) {
		r.logger.Warn("Invalid password for user '%s'", username)
		return false, nil, nil
	}

	user, err := model.ModelUserToProtoUser(userModel)
	if err != nil {
		r.logger.Error("Failed to convert user model into protobuf user: %v", err)
		return false, nil, err
	}

	r.logger.Info("User '%s' logged in successfully", username)
	return true, user, nil
}

// Register creates a new user in the database.
func (r *GormAuthRepository) Register(username, password, email, phone string) (bool, error) {
	var existingUser model.User
	err := r.db.
		Where("username = ? OR email = ?", username, email).
		First(&existingUser).Error

	if err == nil {
		r.logger.Warn("User with username '%s' or email '%s' already exists", username, email)
		return false, nil
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		r.logger.Error("Error while checking existing users: %v", err)
		return false, err
	}

	hashedPassword, err := r.HashPassword(password)
	if err != nil {
		r.logger.Error("Error hashing password: %v", err)
		return false, err
	}

	newUser := model.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Email:        email,
		Phone:        phone,
		RoleID:       model.RoleClient, // default for every new user
	}

	if err := r.db.Create(&newUser).Error; err != nil {
		r.logger.Error("Error creating user '%s': %v", username, err)
		return false, err
	}

	r.logger.Info("User '%s' registered successfully", username)
	return true, nil
}

// ChangePassword updates a user's password after verifying old password.
func (r *GormAuthRepository) ChangePassword(username, oldPassword, newPassword string) (bool, error) {
	success, _, err := r.Login(username, oldPassword)
	if err != nil {
		r.logger.Error("Error during the login of the user '%s': %v", username, err)
		return false, err
	}

	if !success {
		r.logger.Warn("Incorrect username or password for user '%s'", username)
		return false, nil
	}

	newHashedPassword, err := r.HashPassword(newPassword)
	if err != nil {
		r.logger.Error("Error hashing the new password for user '%s': %v", username, err)
		return false, err
	}

	err = r.db.
		Model(&model.User{}).
		Where("username = ?", username).
		Update("password_hash", newHashedPassword).Error

	if err != nil {
		r.logger.Error("Error updating password for user '%s': %v", username, err)
		return false, err
	}

	r.logger.Info("Password changed successfully for user '%s'", username)
	return true, nil
}

// UpdateUser updates user email, phone or role.
func (r *GormAuthRepository) UpdateUser(username, email, phone string, role pb.Role) (bool, error) {
	_, err := r.getUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("User '%s' not found", username)
			return false, nil
		}
		r.logger.Error("Error retrieving user with username '%s': %v", username, err)
		return false, err
	}

	updates := map[string]interface{}{}

	if email != "" {
		updates["email"] = email
	}
	if phone != "" {
		updates["phone"] = phone
	}

	roleID := model.ProtoRoleToModelRole(role)
	if roleID != model.RoleUnspecified {
		updates["role_id"] = roleID
	}

	if len(updates) == 0 {
		r.logger.Info("No updates required for user '%s'", username)
		return true, nil
	}

	err = r.db.Model(&model.User{}).
		Where("username = ?", username).
		Updates(updates).Error

	if err != nil {
		r.logger.Error("Failed to update user %s: %v", username, err)
		return false, err
	}

	r.logger.Info("User '%s' updated successfully", username)
	return true, nil
}

// GetUser retrieves a user by username.
func (r *GormAuthRepository) GetUser(username string) (bool, *pb.User, error) {
	userModel, err := r.getUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("User '%s' not found", username)
			return false, nil, nil
		}
		r.logger.Error("Error retrieving user '%s': %v", username, err)
		return false, nil, err
	}

	user, err := model.ModelUserToProtoUser(userModel)
	if err != nil {
		r.logger.Error("Failed to convert user model into protobuf user: %v", err)
		return false, nil, err
	}

	r.logger.Info("User '%s' retrieved successfully", username)
	return true, user, nil
}

// GetUsers retrieves all users.
func (r *GormAuthRepository) GetUsers() (bool, []*pb.User, error) {
	var userModels []model.User
	err := r.db.Preload("Role").Find(&userModels).Error
	if err != nil {
		r.logger.Error("Failed to retrieve users: %v", err)
		return false, nil, err
	}
	if len(userModels) == 0 {
		r.logger.Warn("No users found")
		return false, nil, nil
	}

	var users []*pb.User
	for _, userModel := range userModels {
		userPb, err := model.ModelUserToProtoUser(&userModel)
		if err != nil {
			r.logger.Warn("Skipping user '%s' due to invalid role mapping: %v", userModel.Username, err)
			continue
		}
		users = append(users, userPb)
	}

	r.logger.Info("Retrieved all users successfully")
	return true, users, nil
}

// getUserByUsername fetches user model from DB by username.
func (r *GormAuthRepository) getUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.
		Preload("Role").
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// HashPassword hashes a plaintext password using bcrypt.
func (r *GormAuthRepository) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares plaintext password with hashed password.
func (r *GormAuthRepository) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateDefaultUsers adds default admin and client users if not present.
func (r *GormAuthRepository) CreateDefaultUsers() error {
	type DefaultUser struct {
		Username string
		Email    string
		Phone    string
		Password string
		Role     pb.Role
	}

	defaultUsers := []DefaultUser{
		{Username: "admin", Email: "admin@example.com", Phone: "0000000000", Password: "admin123", Role: pb.Role_ADMIN},
		{Username: "demo-client", Email: "demo-client@example.com", Phone: "1111111111", Password: "demo123", Role: pb.Role_CLIENT},
	}

	for _, user := range defaultUsers {
		success, err := r.Register(user.Username, user.Password, user.Email, user.Phone)
		if err != nil {
			r.logger.Error("Error registering default user '%s': %v", user.Username, err)
			return err
		}
		if success {
			r.logger.Info("Default user '%s' created successfully", user.Username)
			if user.Role == pb.Role_ADMIN {
				updated, err := r.UpdateUser(user.Username, "", "", user.Role)
				if err != nil {
					r.logger.Error("Error updating role for user '%s': %v", user.Username, err)
					return err
				}
				if updated {
					r.logger.Info("Role '%s' set for user '%s'", user.Role.String(), user.Username)
				}
			}
		} else {
			r.logger.Warn("Default user '%s' not created: another user has the same username or email", user.Username)
		}
	}
	return nil
}
