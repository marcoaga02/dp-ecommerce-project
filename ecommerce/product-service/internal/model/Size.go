package model

type Size struct {
	ID   int    `gorm:"column:id;primaryKey"`
    Name string `gorm:"column:name;type:varchar(20);uniqueIndex;not null"`
}