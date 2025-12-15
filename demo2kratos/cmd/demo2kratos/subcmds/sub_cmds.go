package subcmds

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-xlan/go-migrate/cobramigration"
	"github.com/go-xlan/go-migrate/migrationparam"
	"github.com/go-xlan/go-migrate/migrationstate"
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
// ./bin/demo2kratos migrate new-script create --version-type TIME --description create_table
// ./bin/demo2kratos migrate new-script create --version-type TIME --description alter_schema
// ./bin/demo2kratos migrate new-script create --version-type TIME --description alter_schema --allow-empty-script true
// ./bin/demo2kratos migrate new-script create --version-type TIME --description alter_column
//
// Update migration scripts:
// 更新迁移脚本:
// ./bin/demo2kratos migrate new-script update
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
// Check migration status:
// 检查迁移状态:
// ./bin/demo2kratos migrate status
//
// Note: Use caution with rollback operations to avoid unintended destructive actions
// 注意: 回退操作要谨慎，避免误操作导致问题
// ./bin/demo2kratos migrate migrate dec (use with caution)
func NewMigrateCmd(logger log.Logger) *cobra.Command {
	var debugMode bool

	var rootCmd = &cobra.Command{
		Use:   "migrate",
		Short: "migrate",
		Long:  "migrate",
		Args:  cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			migrationparam.SetDebugMode(debugMode)
		},
	}
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug mode")

	const scriptsInRoot = "./scripts"

	param := migrationparam.NewMigrationParam(
		func() *gorm.DB {
			cfg := appcfg.ParseConfig(cfgpath.ConfigPath)
			dsn := must.Nice(cfg.Data.Database.Source)
			db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{}))
			return db
		},
		func(db *gorm.DB) *migrate.Migrate {
			rawDB := rese.P1(db.DB())
			migrationDriver := rese.V1(sqlite3migrate.WithInstance(rawDB, &sqlite3migrate.Config{}))
			return rese.P1(newmigrate.NewWithScriptsAndDatabase(
				&newmigrate.ScriptsAndDatabaseParam{
					ScriptsInRoot:    scriptsInRoot,
					DatabaseName:     "sqlite3",
					DatabaseInstance: migrationDriver,
				},
			))
		},
	)
	rootCmd.AddCommand(newscripts.NewScriptCmd(&newscripts.Config{
		Param:   param,
		Options: newscripts.NewOptions(scriptsInRoot),
		Objects: models.Objects(),
	}))
	rootCmd.AddCommand(cobramigration.NewMigrateCmd(param))
	rootCmd.AddCommand(previewmigrate.NewPreviewCmd(param, scriptsInRoot))
	rootCmd.AddCommand(migrationstate.NewStatusCmd(&migrationstate.Config{
		Param:       param,
		ScriptsPath: scriptsInRoot,
		Objects:     models.Objects(),
	}))

	return rootCmd
}
