package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal/model"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/auth-service/internal/repository"
	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	pb "github.com/marcoaga02/dp-ecommerce-project/ecommerce/proto/auth"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupDockerDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf(
		"root:%s@tcp(%s:%s)/%s",
		getEnvOrFail(t, "DB_PASSWORD"),
		getEnvOrFail(t, "DB_HOST"),
		getEnvOrFail(t, "DB_PORT"),
		getEnvOrFail(t, "DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to Docker DB: %v", err)
	}

	return db
}

func getEnvOrFail(t *testing.T, key string) string {
	val := os.Getenv(key)
	if val == "" {
		t.Fatalf("Environment variable %s not set", key)
	}
	return val
}

func cleanDB(t *testing.T, db *gorm.DB) {
	if err := db.Exec("TRUNCATE TABLE users").Error; err != nil {
		t.Fatalf("Failed to truncate users table: %v", err)
	}
}

func createDefaultUsers(t *testing.T, db *gorm.DB, repo *repository.GormAuthRepository) {
	hashPass1, err := repo.HashPassword("pass1")
	if err != nil {
		t.Fatalf("Failed to hash password 'pass1': %v", err)
	}
	hashPass2, err := repo.HashPassword("pass2")
	if err != nil {
		t.Fatalf("Failed to hash password 'pass2': %v", err)
	}
	users := []model.User{
		{Username: "user1", PasswordHash: hashPass1, Email: "user1@example.com", Phone: "1111111111"},
		{Username: "user2", PasswordHash: hashPass2, Email: "user2@example.com", Phone: "2222222222"},
	}

	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("Failed to create default users with GORM: %v", err)
	}
}

func setupTestAuthRepo(t *testing.T) (*repository.GormAuthRepository, *gorm.DB) {
	db := setupDockerDB(t)
	cleanDB(t, db)
	authRepo := repository.NewGormAuthRepository(db, logger.NewStdLogger(logger.Info, "gorm-auth-repo-test"))
	createDefaultUsers(t, db, authRepo)
	return authRepo, db
}

func setupTestEmptyAuthRepo(t *testing.T) (*repository.GormAuthRepository, *gorm.DB) {
	db := setupDockerDB(t)
	cleanDB(t, db)
	authRepo := repository.NewGormAuthRepository(db, logger.NewStdLogger(logger.Info, "gorm-auth-repo-test"))
	return authRepo, db
}

func TestLoginWithCorrectCredentials(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)

	succ, user, err := authRepo.Login("user2", "pass2")
	if err != nil {
		t.Fatalf("Expected login to succeed, got error: %v", err)
	}
	if !succ {
		t.Fatalf("Expected login to succeed, got not succ")
	}
	if user.Username != "user2" {
		t.Fatalf("Expected username 'user2', got '%s'", user.Username)
	}
	if user.Email != "user2@example.com" {
		t.Fatalf("Expected email 'user2@example.com', got '%s'", user.Email)
	}
	if user.Phone != "2222222222" {
		t.Fatalf("Expected phone '2222222222', got '%s'", user.Phone)
	}
	if user.Role != pb.Role_CLIENT {
		t.Fatalf("Expected role 'CLIENT', got '%s'", user.Role.String())
	}
}

func TestLoginWithWrongUsername(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)
	succ, user, err := authRepo.Login("userUnk", "pass2")
	if err != nil {
		t.Fatalf("Expected no error for non-existent user, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected login to fail for non-existent user")
	}
	if user != nil {
		t.Fatalf("Expected no user returned for non-existent username")
	}
}

func TestLoginWithWrongPassword(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)
	succ, user, err := authRepo.Login("user1", "passWrong")
	if err != nil {
		t.Fatalf("Expected no error for wrong password, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected login to fail for wrong password")
	}
	if user != nil {
		t.Fatalf("Expected no user returned for wrong password")
	}
}

func TestLoginWithWrongUsernameAndPassword(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)
	succ, user, err := authRepo.Login("userUnk", "passWrong")
	if err != nil {
		t.Fatalf("Expected no error for wrong username and password, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected login to fail for wrong username and password")
	}
	if user != nil {
		t.Fatalf("Expected no user returned for wrong username and password")
	}
}

func TestLoginWithEmptyUsername(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)
	succ, user, err := authRepo.Login("", "pass1")
	if err != nil {
		t.Fatalf("Expected no error for empty username, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected login to fail for empty username")
	}
	if user != nil {
		t.Fatalf("Expected no user returned for empty username")
	}
}

func TestLoginWithEmptyPasswordAndCorrectUsername(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)
	succ, user, err := authRepo.Login("user1", "")
	if err != nil {
		t.Fatalf("Expected no error for empty password, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected login to fail for empty password")
	}
	if user != nil {
		t.Fatalf("Expected no user returned for empty password")
	}
}

func TestLoginWithEmptyPasswordAndWrongUsername(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)
	succ, user, err := authRepo.Login("userUnk", "")
	if err != nil {
		t.Fatalf("Expected no error for empty password, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected login to fail for empty password")
	}
	if user != nil {
		t.Fatalf("Expected no user returned for empty password")
	}
}

func TestLoginWithEmptyUsernameAndPassword(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)
	succ, user, err := authRepo.Login("", "")
	if err != nil {
		t.Fatalf("Expected no error for empty username and password, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected login to fail for empty username and password")
	}
	if user != nil {
		t.Fatalf("Expected no user returned for empty username and password")
	}
}

func TestRegisterNewUser(t *testing.T) {
	authRepo, db := setupTestAuthRepo(t)

	success, err := authRepo.Register("newuser", "newpass", "new@example.com", "1234567890")
	if err != nil {
		t.Fatalf("Expected registration to succeed, got error: %v", err)
	}
	if !success {
		t.Fatalf("Expected registration to succeed")
	}

	var user model.User
	err = db.Where("username = ?", "newuser").First(&user).Error
	if err != nil {
		t.Fatalf("User not found in database after registration: %v", err)
	}
	if user.Email != "new@example.com" {
		t.Fatalf("Expected email 'new@example.com', got '%s'", user.Email)
	}
	if user.Phone != "1234567890" {
		t.Fatalf("Expected phone '1234567890', got '%s'", user.Phone)
	}
	if user.RoleID != model.RoleClient {
		t.Fatalf("Expected RoleClient, got %d", user.RoleID)
	}
	if user.PasswordHash == "newpass" {
		t.Fatalf("Password should be hashed, found plaintext")
	}
	if !authRepo.CheckPasswordHash("newpass", user.PasswordHash) {
		t.Fatalf("Password hash verification failed - hash doesn't match original password")
	}
}

func TestRegisterExistingUsername(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)

	success, err := authRepo.Register("user1", "newpass", "new@example.com", "1234567890")
	if err != nil {
		t.Fatalf("Expected no error for existing user registration, got: %v", err)
	}
	if success {
		t.Fatalf("Expected registration to fail for existing username")
	}
}

func TestRegisterExistingEmail(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)

	success, err := authRepo.Register("usernew", "newpass", "user2@example.com", "1234567890")
	if err != nil {
		t.Fatalf("Expected no error for existing email registration, got: %v", err)
	}
	if success {
		t.Fatalf("Expected registration to fail for existing email")
	}
}

func TestRegisterExistingUsernameAndEmail(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)

	success, err := authRepo.Register("user2", "newpass", "user2@example.com", "1234567890")
	if err != nil {
		t.Fatalf("Expected no error for registration of existing username and email, got: %v", err)
	}
	if success {
		t.Fatalf("Expected registration to fail for registration of existing username and email")
	}
}

func TestChangePasswordSuccessfully(t *testing.T) {
	authRepo, db := setupTestAuthRepo(t)

	succ, err := authRepo.ChangePassword("user1", "pass1", "newpass")
	if err != nil {
		t.Fatalf("Expected no error for change password with correct data, got: %v", err)
	}
	if !succ {
		t.Fatalf("No expected change password to fail with correct data")
	}

	var user model.User
	err = db.Where("username = ?", "user1").First(&user).Error
	if err != nil {
		t.Fatalf("User not found in database after password change: %v", err)
	}
	if user.Email != "user1@example.com" {
		t.Fatalf("Expected same email 'user1@example.com' after password change, got '%s'", user.Email)
	}
	if user.Phone != "1111111111" {
		t.Fatalf("Expected same phone '1111111111' after password change, got '%s'", user.Phone)
	}
	if user.RoleID != model.RoleClient {
		t.Fatalf("Expected same role RoleClient after password change, got %d", user.RoleID)
	}
	if user.PasswordHash == "newpass" {
		t.Fatalf("Password should be hashed, found plaintext")
	}
	if !authRepo.CheckPasswordHash("newpass", user.PasswordHash) {
		t.Fatalf("Password hash verification failed - hash doesn't match the new password")
	}
}

func TestChangePasswordUsingCurrentPasswordAsTheNewOne(t *testing.T) {
	authRepo, db := setupTestAuthRepo(t)

	succ, err := authRepo.ChangePassword("user1", "pass1", "pass1")
	if err != nil {
		t.Fatalf("Expected no error for change password using the current password as the new one, got: %v", err)
	}
	if !succ {
		t.Fatalf("No expected change password to fail using the current password as the new one")
	}

	var user model.User
	err = db.Where("username = ?", "user1").First(&user).Error
	if err != nil {
		t.Fatalf("User not found in database after password change: %v", err)
	}
	if user.PasswordHash == "pass1" {
		t.Fatalf("Password should be hashed, found plaintext")
	}
	if !authRepo.CheckPasswordHash("pass1", user.PasswordHash) {
		t.Fatalf("Password hash verification failed - hash doesn't match the new password")
	}
}

func TestChangePasswordWithWrongUsername(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)

	succ, err := authRepo.ChangePassword("userUnk", "pass1", "passnew")
	if err != nil {
		t.Fatalf("Expected no error for change password using non-existent username, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected change password to fail using non-existent username")
	}
}

func TestChangePasswordWithWrongCurrentPassword(t *testing.T) {
	authRepo, db := setupTestAuthRepo(t)

	succ, err := authRepo.ChangePassword("user1", "passWrong", "passnew")
	if err != nil {
		t.Fatalf("Expected no error for change password using wrong current passwornd, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected change password to fail using wrong current password")
	}

	var user model.User
	err = db.Where("username = ?", "user1").First(&user).Error
	if err != nil {
		t.Fatalf("User not found in database after unsuccessful password change: %v", err)
	}
	if !authRepo.CheckPasswordHash("pass1", user.PasswordHash) {
		t.Fatalf("Password hash verification failed - hash doesn't match the current password (change password failed but the password has been changes)")
	}
}

func TestUpdateUserEmailSuccessful(t *testing.T) {
	authRepo, db := setupTestAuthRepo(t)
	succ, err := authRepo.UpdateUser("user1", "new@example.com", "", pb.Role_UNSPECIFIED)
	if err != nil {
		t.Fatalf("Expected no error for user update with correct email, got: %v", err)
	}
	if !succ {
		t.Fatalf("Expected no failure for user update with correct email")
	}

	var user model.User
	err = db.Where("username = ?", "user1").First(&user).Error
	if err != nil {
		t.Fatalf("User not found in database after user update: %v", err)
	}
	if user.Email != "new@example.com" {
		t.Fatalf("Expected email 'new@example.com' after user email update, got '%s'", user.Email)
	}
	if user.Phone != "1111111111" {
		t.Fatalf("Expected same phone '1111111111' after user email update, got '%s'", user.Phone)
	}
	if user.RoleID != model.RoleClient {
		t.Fatalf("Expected same role RoleClient after user email update, got %d", user.RoleID)
	}
	if !authRepo.CheckPasswordHash("pass1", user.PasswordHash) {
		t.Fatalf("Password hash verification failed - expected same password after email update")
	}
}

func TestUpdateUserPhoneSuccessful(t *testing.T) {
	authRepo, db := setupTestAuthRepo(t)
	succ, err := authRepo.UpdateUser("user1", "", "333000", pb.Role_UNSPECIFIED)
	if err != nil {
		t.Fatalf("Expected no error for user update with correct phone, got: %v", err)
	}
	if !succ {
		t.Fatalf("Expected no failure for user update with correct phone")
	}

	var user model.User
	err = db.Where("username = ?", "user1").First(&user).Error
	if err != nil {
		t.Fatalf("User not found in database after user update: %v", err)
	}
	if user.Email != "user1@example.com" {
		t.Fatalf("Expected same email 'user1@example.com' after user phone update, got '%s'", user.Email)
	}
	if user.Phone != "333000" {
		t.Fatalf("Expected new phone '333000' after user phone update, got '%s'", user.Phone)
	}
	if user.RoleID != model.RoleClient {
		t.Fatalf("Expected same role RoleClient after user phone update, got %d", user.RoleID)
	}
	if !authRepo.CheckPasswordHash("pass1", user.PasswordHash) {
		t.Fatalf("Password hash verification failed - expected same password after phone update")
	}
}

func TestUpdateUserRoleSuccessful(t *testing.T) {
	authRepo, db := setupTestAuthRepo(t)
	succ, err := authRepo.UpdateUser("user1", "", "", pb.Role_ADMIN)
	if err != nil {
		t.Fatalf("Expected no error for user update with correct role, got: %v", err)
	}
	if !succ {
		t.Fatalf("Expected no failure for user update with correct role")
	}

	var user model.User
	err = db.Where("username = ?", "user1").First(&user).Error
	if err != nil {
		t.Fatalf("User not found in database after user update: %v", err)
	}
	if user.Email != "user1@example.com" {
		t.Fatalf("Expected same email 'user1@example.com' after user role update, got '%s'", user.Email)
	}
	if user.Phone != "1111111111" {
		t.Fatalf("Expected same phone '1111111111' after user role update, got '%s'", user.Phone)
	}
	if user.RoleID != model.RoleAdmin {
		t.Fatalf("Expected new role RoleAdmin after user role update, got %d", user.RoleID)
	}
	if !authRepo.CheckPasswordHash("pass1", user.PasswordHash) {
		t.Fatalf("Password hash verification failed - expected same password after role update")
	}
}

func TestUpdateUserPhoneEmailRoleSuccessful(t *testing.T) {
	authRepo, db := setupTestAuthRepo(t)
	succ, err := authRepo.UpdateUser("user1", "new@example.com", "333000", pb.Role_ADMIN)
	if err != nil {
		t.Fatalf("Expected no error for user update with correct role, got: %v", err)
	}
	if !succ {
		t.Fatalf("Expected no failure for user update with correct role")
	}

	var user model.User
	err = db.Where("username = ?", "user1").First(&user).Error
	if err != nil {
		t.Fatalf("User not found in database after user update: %v", err)
	}
	if user.Email != "new@example.com" {
		t.Fatalf("Expected new email 'new@example.com' after full user update, got '%s'", user.Email)
	}
	if user.Phone != "333000" {
		t.Fatalf("Expected new phone '333000' after full user update, got '%s'", user.Phone)
	}
	if user.RoleID != model.RoleAdmin {
		t.Fatalf("Expected new role RoleAdmin after full user update, got %d", user.RoleID)
	}
	if !authRepo.CheckPasswordHash("pass1", user.PasswordHash) {
		t.Fatalf("Password hash verification failed - expected same password after full user update")
	}
}

func TestUpdateUserWithNoRequiredChanges(t *testing.T) {
	authRepo, db := setupTestAuthRepo(t)
	succ, err := authRepo.UpdateUser("user1", "", "", pb.Role_UNSPECIFIED)
	if err != nil {
		t.Fatalf("Expected no error for user update with no required updates, got: %v", err)
	}
	if !succ {
		t.Fatalf("Expected no failure for user update with no required updates")
	}

	var user model.User
	err = db.Where("username = ?", "user1").First(&user).Error
	if err != nil {
		t.Fatalf("User not found in database after user update with no required updates: %v", err)
	}
	if user.Email != "user1@example.com" {
		t.Fatalf("Expected same email 'user1@example.com' after user update with no required changes, got '%s'", user.Email)
	}
	if user.Phone != "1111111111" {
		t.Fatalf("Expected same phone '1111111111' after user update with no required changes, got '%s'", user.Phone)
	}
	if user.RoleID != model.RoleClient {
		t.Fatalf("Expected same role RoleClient after user update with no required changes, got %d", user.RoleID)
	}
	if !authRepo.CheckPasswordHash("pass1", user.PasswordHash) {
		t.Fatalf("Password hash verification failed - expected same password after user update with no required changes")
	}
}

func TestUpdateUserWithWrongUsername(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)
	succ, err := authRepo.UpdateUser("userWrong", "a@a.com", "333", pb.Role_ADMIN)
	if err != nil {
		t.Fatalf("Expected no error for user update with wrong username, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected failure for user update with wrong username")
	}
}

func TestGetUserWithCorrectUsername(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)
	succ, user, err := authRepo.GetUser("user2")
	if err != nil {
		t.Fatalf("Expected no error for user retrieval with correct username, got: %v", err)
	}
	if !succ {
		t.Fatalf("Expected no failure for user retrieval with correct username")
	}
	if user == nil {
		t.Fatalf("Unexpected nil user during retrieval with correct username")
	}

	if user.Username != "user2" {
		t.Fatalf("Expected username 'user2', got '%s'", user.Username)
	}
	if user.Email != "user2@example.com" {
		t.Fatalf("Expected email 'user2@example.com', got '%s'", user.Email)
	}
	if user.Phone != "2222222222" {
		t.Fatalf("Expected phone '2222222222', got '%s'", user.Phone)
	}
	if user.Role != pb.Role_CLIENT {
		t.Fatalf("Expected role 'CLIENT', got '%s'", user.Role.String())
	}
}

func TestGetUserWithWrongUsername(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)

	succ, user, err := authRepo.GetUser("userUnk")
	if err != nil {
		t.Fatalf("Expected no error for user retrieval with wrong username, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected no success for user retrieval with wrong username")
	}
	if user != nil {
		t.Fatalf("Wxpected nil user during retrieval with wrong username")
	}
}

func TestGetUsersWithTwoUsersInDb(t *testing.T) {
	authRepo, _ := setupTestAuthRepo(t)

	succ, users, err := authRepo.GetUsers()
	if err != nil {
		t.Fatalf("Expected no error for users retrieval, got: %v", err)
	}
	if !succ {
		t.Fatalf("Expected no failure for users retrieval")
	}
	if users == nil {
		t.Fatalf("Unexpected nil users during retrieval")
	}
	if len(users) != 2 {
		t.Fatalf("Expected list of two users")
	}

	foundUser1 := false
	foundUser2 := false

	for _, u := range users {
		if u.Username == "user1" {
			foundUser1 = true
			if u.Email != "user1@example.com" {
				t.Fatalf("Expected email 'user1@example.com', got '%s'", u.Email)
			}
			if u.Phone != "1111111111" {
				t.Fatalf("Expected phone '1111111111', got '%s'", u.Phone)
			}
			if u.Role != pb.Role_CLIENT {
				t.Fatalf("Expected role 'CLIENT', got '%s'", u.Role.String())
			}
		}
		if u.Username == "user2" {
			foundUser2 = true
			if u.Email != "user2@example.com" {
				t.Fatalf("Expected email 'user2@example.com', got '%s'", u.Email)
			}
			if u.Phone != "2222222222" {
				t.Fatalf("Expected phone '2222222222', got '%s'", u.Phone)
			}
			if u.Role != pb.Role_CLIENT {
				t.Fatalf("Expected role 'CLIENT', got '%s'", u.Role.String())
			}
		}
	}
	if !foundUser1 || !foundUser2 {
		t.Fatalf("Expected to find both 'user1' and 'user2', but one or both has not been found")
	}
}

func TestGetUsersWithEmptyUserList(t *testing.T) {
	authRepo, _ := setupTestEmptyAuthRepo(t)
	succ, users, err := authRepo.GetUsers()
	if err != nil {
		t.Fatalf("Expected no error for users retrieval with empty user list, got: %v", err)
	}
	if succ {
		t.Fatalf("Expected failure for users retrieval with empty user list")
	}
	if users != nil {
		t.Fatalf("Unexpected non-nil user list during retrieval")
	}
}
