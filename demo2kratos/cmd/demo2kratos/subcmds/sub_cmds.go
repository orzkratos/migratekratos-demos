package subcmds

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-xlan/go-migrate/cobramigration"
	"github.com/go-xlan/go-migrate/newmigrate"
	"github.com/go-xlan/go-migrate/newscripts"
	"github.com/go-xlan/go-migrate/previewmigrate"
	"github.com/golang-migrate/migrate/v4"
	sqlite3migrate "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/cfgpath"
	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/appcfg"
	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/models"
	"github.com/spf13/cobra"
	"github.com/yyle88/must"
	"github.com/yyle88/rese"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewVersionCmd creates version command
// 创建版本命令
func NewVersionCmd(serviceName, version string, logger log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			slog := log.NewHelper(logger)
			slog.Infof("service-name: %s version: %s", serviceName, version)
		},
	}
}

// NewMigrateCmd creates migrate command with database access
// 创建带数据库访问的 migrate 命令
//
// Example commands:
// 示例命令:
//
// Create migration scripts:
// 创建迁移脚本:
// ./bin/demo2kratos migrate next-script create --version-type TIME --description create_table
// ./bin/demo2kratos migrate next-script create --version-type TIME --description alter_schema
// ./bin/demo2kratos migrate next-script create --version-type TIME --description alter_schema --allow-empty-script true
// ./bin/demo2kratos migrate next-script create --version-type TIME --description alter_column
//
// Update migration scripts:
// 更新迁移脚本:
// ./bin/demo2kratos migrate next-script update
//
// Execute migrations:
// 执行迁移:
// ./bin/demo2kratos migrate migrate all
// ./bin/demo2kratos migrate migrate inc
//
// Preview migrations:
// 预览迁移:
// ./bin/demo2kratos migrate preview inc
//
// Note: Use caution with rollback operations to avoid unintended destructive actions
// 注意: 回退操作要谨慎，避免误操作导致问题
// ./bin/demo2kratos migrate migrate dec (use with caution)
func NewMigrateCmd(logger log.Logger) *cobra.Command {
	const scriptsInRoot = "./scripts"

	// Lazy initialization: database connection created when command runs
	// 延迟初始化：仅在命令运行时才创建数据库连接
	getDB := func() *gorm.DB {
		cfg := appcfg.ParseConfig(cfgpath.ConfigPath)
		dsn := must.Nice(cfg.Data.Database.Source)
		db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{}))
		return db
	}

	// getMigration function accepts database connection to share single connection
	// 迁移工厂接受数据库连接以共享单个连接
	getMigration := func(db *gorm.DB) *migrate.Migrate {
		sqlDB := rese.P1(db.DB())
		migrationDriver := rese.V1(sqlite3migrate.WithInstance(sqlDB, &sqlite3migrate.Config{}))
		return rese.P1(newmigrate.NewWithScriptsAndDatabase(
			&newmigrate.ScriptsAndDatabaseParam{
				ScriptsInRoot:    scriptsInRoot,
				DatabaseName:     "sqlite3",
				DatabaseInstance: migrationDriver,
			},
		))
	}

	var rootCmd = &cobra.Command{
		Use:   "migrate",
		Short: "migrate",
		Long:  "migrate",
	}
	rootCmd.AddCommand(newscripts.NextScriptCmd(&newscripts.Config{
		GetMigration: getMigration,
		GetDB:        getDB,
		Options:      newscripts.NewOptions(scriptsInRoot),
		Objects: []any{
			&models.Record{},
			&models.Article{},
			&models.Product{},
		},
	}))
	rootCmd.AddCommand(cobramigration.NewMigrateCmd(getDB, getMigration))
	rootCmd.AddCommand(previewmigrate.NewPreviewCmd(getDB, getMigration, scriptsInRoot))

	return rootCmd
}
