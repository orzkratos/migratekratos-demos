package models

// Objects returns all GORM model objects for migration
// 返回所有用于迁移的 GORM 模型对象
func Objects() []any {
	return []any{
		&Record{},
	}
}
