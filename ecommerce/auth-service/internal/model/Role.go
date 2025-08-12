package model

// Role represents a user role within the system (e.g., admin, client).
//
// Fields:
//   - ID: unique identifier for the role (auto-incremented primary key).
//   - Name: unique name of the role (e.g., "admin", "client"); required and must be unique.
type Role struct {
    ID   int    `gorm:"column:id;primaryKey"`
    Name string `gorm:"column:name;type:varchar(20);uniqueIndex;not null"`
}