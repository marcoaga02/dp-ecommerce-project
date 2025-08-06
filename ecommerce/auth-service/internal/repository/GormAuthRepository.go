package repository

import (
	"fmt"
	
	"github.com/google/uuid"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal/model"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)


// GormAuthRepository implements the AuthDB interface using GORM as the ORM layer.
//
// It provides methods for user authentication, registration, password management, 
// and role handling by interacting with the underlying database.
type GormAuthRepository struct {
	db *gorm.DB
	logger logger.Logger
}

// Constants representing user roles in the database.
const (
    RoleClient = 1 // Regular client user
    RoleAdmin  = 2 // Administrative user
)

// roleMapDb2Enum maps integer role values from the database to protobuf enum values.
var roleMapDb2Enum = map[int]pb.Role{
    RoleClient: pb.Role_CLIENT,
    RoleAdmin: pb.Role_ADMIN,
}

// roleMapEnum2Db maps protobuf enum role values to integer values used in the database.
var roleMapEnum2Db = map[pb.Role]int{
    pb.Role_CLIENT: RoleClient,
    pb.Role_ADMIN: RoleAdmin,
}

// NewGormAuthRepository initializes and returns a new GormAuthRepository.
//
// Parameters:
//   - db: GORM database connection
//   - logger: logger instance for logging messages
//
// Returns:
//   - *GormAuthRepository: a pointer to the initialized repository
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
//   - pb.Role: the role of the authenticated user (only valid if authentication is successful)
//   - error: non-nil if an unexpected error occurs (e.g., DB failure, unknown user role)
//
// Notes:
//   - If the user is not found or the password is incorrect, the returned error is nil and success is false.
//   - If the user's role is not recognized, an error is returned and login fails.
func (r *GormAuthRepository) Login(username, password string) (bool, pb.Role, error) {
	user, err := r.getUserByUsername(username)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Warn("User %s not found", username)
			return false, pb.Role_UNSPECIFIED, nil
		}
		r.logger.Error("Error retrieving user with username '%s': %v", username, err)
		return false, pb.Role_UNSPECIFIED, err
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		r.logger.Warn("Invalid password for user '%s'", username)
		return false, pb.Role_UNSPECIFIED, nil
	}

	role, ok := roleMapDb2Enum[user.RoleID]
	if !ok {
		r.logger.Error("Unknown role '%d' for user '%s'", user.RoleID, username)
		return false, pb.Role_UNSPECIFIED, fmt.Errorf("Invalid role '%d' for user", user.RoleID)
	}

	return true, role, nil
}


// Register creates a new user with the provided username, password, email, and phone.
//
// Parameters:
//   - username: desired unique username
//   - password: plaintext password (will be hashed before storing)
//   - email: unique email address
//   - phone: phone number
//
// Returns:
//   - bool: true if registration is successful, false if username or email already exists
//   - error: non-nil if a system or database error occurs
func (r *GormAuthRepository) Register(username, password, email, phone string) (bool, error) {
	var existingUser model.User
	err := r.db.
		Where("username = ? OR email = ?", username, email).
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
		RoleID:		  RoleClient, // default for every new user
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
//   - bool: true if password updated successfully, false if old password incorrect or an error occurs
//   - error: non-nil if unexpected failure occurs
func (r *GormAuthRepository) ChangePassword(username, old_password, new_password string) (bool, error) {
	success, _, err := r.Login(username, old_password)

	if err != nil {
		r.logger.Error("Error during the login of the user '%s': %v", username, err)
		return false, err
	}

	if !success {
		r.logger.Warn("Incorrect old password for user '%s'", username)
		return false, nil
	}

	newHashedPassword, err := HashPassword(new_password)
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

	return true, nil
}


// SetUserRole updates the role of the user identified by the given username.
//
// It first checks if the user exists; if not, returns an error.
//
// Then, it verifies the current role of the user:
//   - If the current role is invalid, returns an error.
//   - If the current role is the same as the requested role, returns false without error.
//
// Next, it validates the requested role and if valid, updates the user's role in the database.
//
// Parameters:
//   - username: the unique username of the user whose role will be updated
//   - role: the new role to assign to the user (pb.Role enum)
//
// Returns:
//   - bool: true if the role was updated, false if no update was needed (e.g., role was already set) or an error occurs
//   - error: non-nil if any error occurs
func (r *GormAuthRepository) SetUserRole(username string, role pb.Role) (bool, error) {
	existingUser, currentRole, err := r.getUserAndRoleByUsername(username)
	if err != nil {
		return false, err
	}

	if currentRole == role {
		r.logger.Warn("User '%s' already has role %v", username, role)
		return false, nil
	}

	newRole, ok := roleMapEnum2Db[role]
	if !ok {
		r.logger.Error("Attempted to set unknown role '%v' for user '%s'", role, username)
		return false, fmt.Errorf("Invalid role value '%v' provided", role)
	}

	err = r.db.
		Model(existingUser).
		Update("role_id", newRole).Error
	if err != nil {
		r.logger.Error("Error updating role for user '%s': %v", username, err)
		return false, err
	}

	return true, nil
}


// GetUserRole retrieves the role of a user given its username.
//
// It returns a boolean indicating whether the role was found, the role itself (as a protobuf enum),
// and an error if something went wrong (e.g., user not found or role invalid).
//
// Parameters:
//   - username: the username of the user whose role is being retrieved.
//
// Returns:
//   - bool: true if the role was successfully retrieved; false otherwise.
//   - pb.Role: the user's role as defined in the protobuf enum (pb.Role_UNSPECIFIED if an error occurs).
//   - error: non-nil if an error occurred during the process (e.g., DB error, unknown role).
func (r *GormAuthRepository) GetUserRole(username string) (bool, pb.Role, error) {
	_, role, err := r.getUserAndRoleByUsername(username)
	if err != nil {
		return false, pb.Role_UNSPECIFIED, err
	}
	return true, role, nil
}


// getUserByUsername fetches a user from the database by its username.
//
// It performs a search in the users table and returns the user model
// if found, or an error if not found or if a DB issue occurred.
//
// Parameters:
//   - username: the username of the user to retrieve.
//
// Returns:
//   - *model.User: pointer to the user model if found.
//   - error: non-nil if user not found or a database error occurs.
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


// getUserAndRoleByUsername retrieves the user by username and maps their DB role to pb.Role.
//
// Parameters:
//   - username: the username of the user to retrieve.
//
// Returns:
//   - *model.User: the user model if found.
//   - pb.Role: the user's role as protobuf enum (pb.Role).
//   - error: non-nil if the user is not found or the role is invalid.
func (r *GormAuthRepository) getUserAndRoleByUsername(username string) (*model.User, pb.Role, error) {
	user, err := r.getUserByUsername(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Error("User with username '%s' does not exist", username)
		} else {
			r.logger.Error("Error retrieving user '%s': %v", username, err)
		}
		return nil, pb.Role_UNSPECIFIED, err
	}

	role, ok := roleMapDb2Enum[user.RoleID]
	if !ok {
		r.logger.Error("Unknown role '%d' found in DB for user '%s'", user.RoleID, username)
		return nil, pb.Role_UNSPECIFIED, fmt.Errorf("Invalid role '%d' found in DB for user", user.RoleID)
	}

	return user, role, nil
}


// CreateDefaultUsers ensures that the default users — one admin and one client — exist in the database.
//
// It checks for the existence of these users by username or email, and skips creation if they are already present.
// If any database error occurs during the checking or creation process, it returns the error.
//
// Returns:
//   - error: non-nil if any database operation fails
func (r *GormAuthRepository) CreateDefaultUsers() error {
	defaultUsers := []model.User{
		{
			Username:     "admin",
			Email:        "admin@example.com",
			Phone:        "0000000000",
			PasswordHash: "admin123",
			RoleID: 	  RoleAdmin,
		},
		{
			Username:     "demo-client",
			Email:        "demo-client@example.com",
			Phone:        "1111111111",
			PasswordHash: "demo123",
			RoleID: 	  RoleClient,
		},
	}

	for _, user := range defaultUsers {
        _, err := r.getUserByUsername(user.Username)
        if err == nil {
            r.logger.Info("Default user '%s' already exists", user.Username)
            continue
        }
        if err != nil && err != gorm.ErrRecordNotFound {
            r.logger.Error("Error checking existence of user '%s': %v", user.Username, err)
            return err
        }

        success, err := r.Register(user.Username, user.PasswordHash, user.Email, user.Phone)
        if err != nil {
            r.logger.Error("Error registering default user '%s': %v", user.Username, err)
            return err
        }
        if success {
            r.logger.Info("Default user '%s' created successfully", user.Username)
			if user.RoleID == RoleAdmin {
				changed, err := r.SetUserRole(user.Username, pb.Role_ADMIN)
				if err != nil {
					r.logger.Error("Failed to set role ADMIN for user '%s': %v", user.Username, err)
					return err
				}
				if changed {
					r.logger.Info("Role ADMIN set successfully for user '%s'", user.Username)
				} else {
					r.logger.Warn("Role for user '%s' was already ADMIN", user.Username)
				}
			}
        } else {
            r.logger.Warn("Default user '%s' not created: another user has the same username or email", user.Username)
        }
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