package model

type Product struct {
	Code        string `gorm:"column:code;type:varchar(32);primaryKey"`
	Name        string `gorm:"column:name;type:varchar(255);not null"`
	SizeID      int    `gorm:"column:size_id;not null"` // foreign key
	Size        Size   `gorm:"foreignKey:SizeID"`       // GORM relationship
	Color       string `gorm:"column:color;type:varchar(30);not null"`
	Description string `gorm:"column:description;type:varchar(255);not null"`
	Stock       int32  `gorm:"column:stock;type:int;not null"`
	Price       float64 `gorm:"column:price;type:double(10,2);not null"`
}
