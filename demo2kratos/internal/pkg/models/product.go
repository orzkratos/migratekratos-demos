package models

import "gorm.io/gorm"

// Product represents a product item
// 产品模型
type Product struct {
	gorm.Model
	Name        string  `gorm:"type:varchar(150);not null"`
	Price       float64 `gorm:"type:decimal(10,2);not null"`
	Stock       int     `gorm:"type:int;default:0"`
	Description string  `gorm:"type:text"`
}

// TableName returns the table name
// 返回表名
func (*Product) TableName() string {
	return "products"
}
