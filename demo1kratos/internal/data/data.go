package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-xlan/go-migrate/checkmigration"
	"github.com/google/wire"
	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
	"github.com/orzkratos/demokratos/demo1kratos/internal/pkg/models"
	"github.com/yyle88/must"
	"github.com/yyle88/rese"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	loggergorm "gorm.io/gorm/logger"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	db *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	dsn := must.Nice(c.Database.Source)
	db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: loggergorm.Default.LogMode(loggergorm.Info),
	}))

	// Check if migration scripts are missing
	// 检查是否缺少迁移脚本
	checkmigration.CheckMigrate(db, models.Objects())

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		must.Done(rese.P1(db.DB()).Close())
	}
	return &Data{
		db: db,
	}, cleanup, nil
}

// DB returns the gorm database instance
// 返回 gorm 数据库实例
func (d *Data) DB() *gorm.DB {
	return d.db
}
