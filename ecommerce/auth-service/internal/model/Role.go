package model

type Role struct {
    ID   int    `gorm:"primaryKey"`
    Name string `gorm:"uniqueIndex;not null"`
}