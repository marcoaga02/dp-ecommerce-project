package model

type User struct {
	ID       string `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex"`
	PasswordHash string
	Email    string
	Phone    string
}