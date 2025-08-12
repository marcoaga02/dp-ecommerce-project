package model

// User represents a user account in the system.
//
// Fields:
//   - ID: unique identifier (UUID) for the user.
//   - Username: unique username used for login.
//   - PasswordHash: hashed password for secure authentication.
//   - Email: unique email address associated with the user.
//   - Phone: phone number.
//   - RoleID: foreign key referencing the user's role.
//   - Role: GORM relation to the Role entity.
type User struct {
	ID           string `gorm:"column:id;type:char(36);primaryKey"`
	Username     string `gorm:"column:username;type:varchar(32);uniqueIndex;not null"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null"`
	Email        string `gorm:"column:email;type:varchar(64);uniqueIndex;not null"`
	Phone        string `gorm:"column:phone;type:varchar(20);not null"`
	RoleID       int    `gorm:"column:role_id;not null;default:1"` // foreign key
	Role         Role   `gorm:"foreignKey:RoleID"`                 // GORM relationship
}
