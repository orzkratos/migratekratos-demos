package models

import "gorm.io/gorm"

// Record represents a database record
// 数据库记录模型
type Record struct {
	gorm.Model
	Message string `gorm:"type:varchar(255)"`
}

// TableName returns the table name
// 返回表名
func (*Record) TableName() string {
	return "records"
}
