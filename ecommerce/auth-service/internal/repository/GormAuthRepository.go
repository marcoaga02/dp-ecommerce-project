package repository

import (
	"fmt"
	"errors"
	
	"github.com/google/uuid"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal/model"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GormAuthRepository is an implementation of the AuthDB interface
// that uses GORM as the underlying ORM to interact with the database.
type GormAuthRepository struct {
	db *gorm.DB
	logger logger.Logger
}

// NewGormAuthRepository creates a new instance of GormAuthRepository
// using the provided GORM database connection.
func NewGormAuthRepository(db *gorm.DB, logger logger.Logger) *GormAuthRepository {
	return &GormAuthRepository{db: db, logger: logger}
}

// Login verifies the user's credentials by username and password.
//
// Parameters:
//   - username: the username of the user attempting to log in
//   - password: the plaintext password provided by the user
//
// Returns:
//   - bool: true if authentication is successful, false if credentials are invalid
//   - error: non-nil if an unexpected error occurs during database operations
func (r *GormAuthRepository) Login(username, password string) (bool, error) {
	var user model.User
	err := r.db.Where("Username = ?", username).First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Warn("User %s not found", username)
			return false, nil
		}
		r.logger.Error("Error retrieving user with username '%s': %v", username, err)
		return false, err
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		r.logger.Warn("Invalid password for user '%s'", username)
		return false, nil
	}

	return true, nil
}

// Register creates a new user with the provided username, password, email, and phone.
//
// Parameters:
//   - username: desired unique username
//   - password: plaintext password, will be hashed before storing
//   - email: unique email address
//   - phone: optional phone number
//
// Returns:
//   - bool: false if username or email already exists, true on successful registration
//   - error: non-nil if system or database error occurs
func (r *GormAuthRepository) Register(username, password, email, phone string) (bool, error) {
	var existingUser model.User
	err := r.db.
		Where("Username = ? OR Email = ?", username, email).
		First(&existingUser).Error

	if err == nil {
		r.logger.Warn("User with username '%s' or email '%s' already exists", username, email)
		return false, nil
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		r.logger.Error("Error while checking existing users: %v", err)
		return false, err
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		r.logger.Error("Error hashing password: %v", err)
		return false, err
	}

	newUser := model.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: hashedPassword,
		Email:        email,
		Phone:        phone,
	}

	if err := r.db.Create(&newUser).Error; err != nil {
		r.logger.Error("Error creating user '%s': %v", username, err)
		return false, err
	}

	return true, nil
}

// ChangePassword updates the password for the given username if the old password matches.
//
// Parameters:
//   - username: the username whose password will be changed
//   - old_password: current plaintext password, used to verify identity
//   - new_password: new plaintext password to replace the old one
//
// Returns:
//   - bool: true if password updated successfully, false if old password incorrect
//   - error: non-nil if unexpected failure occurs
func (r *GormAuthRepository) ChangePassword(username, old_password, new_password string) (bool, error) {
	correctData, err := r.Login(username, old_password)

	if err != nil {
		r.logger.Error("Error during the login of the user '%s': %v", username, err)
		return false, err
	}

	if !correctData {
		r.logger.Warn("Incorrect old password for user '%s'", username)
		return false, nil
	}

	newHashedPassword, err := HashPassword(new_password)
	if err != nil {
		r.logger.Error("Error hashing the new password for user '%s': %v", username, err)
		return false, err
	}

	err = r.db.Model(&model.User{}).
		Where("Username = ?", username).
		Update("PasswordHash", newHashedPassword).Error
	
	if err != nil {
		r.logger.Error("Error updating password for user '%s': %v", username, err)
		return false, err
	}

	return true, nil
}

// CreateDefaultUsers checks for the existence of default users (admin and demo).
//
// Parameters:
//   - none
//
// Returns:
//   - error: non-nil if any database operation fails
func (r *GormAuthRepository) CreateDefaultUsers() error {
	defaultUsers := []model.User{
		{
			Username:     "admin",
			Email:        "admin@example.com",
			Phone:        "0000000000",
			PasswordHash: HashPasswordOrPanic("admin123"),
		},
		{
			Username:     "demo",
			Email:        "demo@example.com",
			Phone:        "1111111111",
			PasswordHash: HashPasswordOrPanic("demo123"),
		},
	}

	for _, user := range defaultUsers {
		var existing model.User
		err := r.db.
			Where("username = ? OR email = ?", user.Username, user.Email).
			First(&existing).Error

		if err == nil {
			r.logger.Info("Default user '%s' already exists", user.Username)
			continue
		}

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Error("Error checking existence of user '%s': %v", user.Username, err)
			return err
		}

		if err := r.db.Create(&user).Error; err != nil {
			r.logger.Error("Error creating default user '%s': %v", user.Username, err)
			return err
		}

		r.logger.Info("Default user '%s' created successfully", user.Username)
	}

	return nil
}

// HashPassword hashes the given plaintext password using bcrypt.
//
// Parameters:
//   - password: the plaintext password to hash
//
// Returns:
//   - string: the bcrypt hashed password
//   - error: non-nil if hashing fails
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a plaintext password with a hashed password.
//
// Parameters:
//   - password: the plaintext password to verify
//   - hash: the bcrypt hashed password stored in database
//
// Returns:
//   - bool: true if the password matches the hash, false otherwise
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashPasswordOrPanic hashes the given plaintext password using bcrypt.
//
// It panics if hashing fails, intended only for default user setup.
//
// Parameters:
//   - password: the plaintext password to hash
//
// Returns:
//   - string: the bcrypt hashed password
func HashPasswordOrPanic(password string) string {
	hash, err := HashPassword(password)
	if err != nil {
		panic(fmt.Sprintf("Failed to hash default password: %v", err))
	}
	return hash
}