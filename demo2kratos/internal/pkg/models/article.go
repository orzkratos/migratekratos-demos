package models

import "gorm.io/gorm"

// Article represents a blog article
// 文章模型
type Article struct {
	gorm.Model
	Title   string `gorm:"type:varchar(200);not null"`
	Content string `gorm:"type:text"`
	Author  string `gorm:"type:varchar(100)"`
	Status  string `gorm:"type:varchar(20);default:'draft'"` // draft, published, archived // 草稿、已发布、已归档
}

// TableName returns the table name
// 返回表名
func (*Article) TableName() string {
	return "articles"
}
