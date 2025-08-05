package model

type User struct {
    ID           string `gorm:"type:char(36);primaryKey"`
    Username     string `gorm:"uniqueIndex;not null"`
    PasswordHash string `gorm:"not null"`
    Email        string `gorm:"uniqueIndex;not null"`
    Phone        string
    RoleID       int    `gorm:"not null;default:1"` // foreign key
    Role         Role   `gorm:"foreignKey:RoleID"`  // GORM relation
}
