package model

type Product struct {
	ID          string `gorm:"column:id;type:char(36);primaryKey"`
    Code        string `gorm:"column:code;type:varchar(32);uniqueIndex;not null"`
    Name        string `gorm:"column:name;type:varchar(255);not null"`
    SizeID      int    `gorm:"column:size_id;not null"` // foreign key
    Size        Size   `gorm:"foreignKey:SizeID"`      // GORM relationship
    Color       string `gorm:"column:color;type:varchar(30);not null"`
    Description string `gorm:"column:description;type:varchar(255);not null"`
    Stock       int32    `gorm:"column:stock;type:int;not null"`
}
