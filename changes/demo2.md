# Changes

Code differences compared to source project demokratos.

## cmd/demo2kratos/cfgpath/cfg_path.go (+5 -0)

```diff
@@ -0,0 +1,5 @@
+package cfgpath
+
+// ConfigPath is the config path.
+// 配置文件路径
+var ConfigPath string
```

## cmd/demo2kratos/main.go (+33 -16)

```diff
@@ -1,19 +1,22 @@
 package main
 
 import (
-	"flag"
+	"fmt"
 	"os"
 
 	"github.com/go-kratos/kratos/v2"
-	"github.com/go-kratos/kratos/v2/config"
-	"github.com/go-kratos/kratos/v2/config/file"
 	"github.com/go-kratos/kratos/v2/log"
 	"github.com/go-kratos/kratos/v2/middleware/tracing"
 	"github.com/go-kratos/kratos/v2/transport/grpc"
 	"github.com/go-kratos/kratos/v2/transport/http"
+	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/cfgpath"
+	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/subcmds"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/appcfg"
+	"github.com/spf13/cobra"
 	"github.com/yyle88/done"
 	"github.com/yyle88/must"
+	"github.com/yyle88/must/mustslice"
 	"github.com/yyle88/rese"
 )
 
@@ -23,12 +26,10 @@
 	Name string
 	// Version is the version of the compiled software.
 	Version string
-	// flagconf is the config flag.
-	flagconf string
 )
 
 func init() {
-	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
+	fmt.Println("service-name:", Name)
 }
 
 func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
@@ -46,7 +47,6 @@
 }
 
 func main() {
-	flag.Parse()
 	logger := log.With(log.NewStdLogger(os.Stdout),
 		"ts", log.DefaultTimestamp,
 		"caller", log.DefaultCaller,
@@ -56,18 +56,35 @@
 		"trace.id", tracing.TraceID(),
 		"span.id", tracing.SpanID(),
 	)
-	c := config.New(
-		config.WithSource(
-			file.NewSource(flagconf),
-		),
-	)
-	defer rese.F0(c.Close)
 
-	must.Done(c.Load())
+	var rootCmd = &cobra.Command{
+		Use:   "demo2kratos",
+		Short: "A Kratos microservice with database migration",
+		Run: func(cmd *cobra.Command, args []string) {
+			mustslice.None(args)
+			if cfg := appcfg.ParseConfig(cfgpath.ConfigPath); cfg.Server.AutoRun {
+				runApp(cfg, logger)
+			}
+		},
+	}
+	rootCmd.PersistentFlags().StringVarP(&cfgpath.ConfigPath, "conf", "c", "./configs", "config path, eg: --conf=config.yaml")
 
-	var cfg conf.Bootstrap
-	must.Done(c.Scan(&cfg))
+	rootCmd.AddCommand(&cobra.Command{
+		Use:   "run",
+		Short: "Start the application",
+		Run: func(cmd *cobra.Command, args []string) {
+			cfg := appcfg.ParseConfig(cfgpath.ConfigPath)
+			runApp(cfg, logger)
+		},
+	})
 
+	rootCmd.AddCommand(subcmds.NewVersionCmd(Name, Version, logger))
+	rootCmd.AddCommand(subcmds.NewMigrateCmd(logger))
+
+	must.Done(rootCmd.Execute())
+}
+
+func runApp(cfg *conf.Bootstrap, logger log.Logger) {
 	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
 	defer cleanup()
 
```

## cmd/demo2kratos/subcmds/sub_cmds.go (+112 -0)

```diff
@@ -0,0 +1,112 @@
+package subcmds
+
+import (
+	"github.com/go-kratos/kratos/v2/log"
+	"github.com/go-xlan/go-migrate/cobramigration"
+	"github.com/go-xlan/go-migrate/migrationparam"
+	"github.com/go-xlan/go-migrate/migrationstate"
+	"github.com/go-xlan/go-migrate/newmigrate"
+	"github.com/go-xlan/go-migrate/newscripts"
+	"github.com/go-xlan/go-migrate/previewmigrate"
+	"github.com/golang-migrate/migrate/v4"
+	sqlite3migrate "github.com/golang-migrate/migrate/v4/database/sqlite3"
+	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/cfgpath"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/appcfg"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/models"
+	"github.com/spf13/cobra"
+	"github.com/yyle88/must"
+	"github.com/yyle88/rese"
+	"gorm.io/driver/sqlite"
+	"gorm.io/gorm"
+)
+
+// NewVersionCmd creates version command
+// 创建版本命令
+func NewVersionCmd(serviceName, version string, logger log.Logger) *cobra.Command {
+	return &cobra.Command{
+		Use:   "version",
+		Short: "Print version info",
+		Run: func(cmd *cobra.Command, args []string) {
+			slog := log.NewHelper(logger)
+			slog.Infof("service-name: %s version: %s", serviceName, version)
+		},
+	}
+}
+
+// NewMigrateCmd creates migrate command with database access
+// 创建带数据库访问的 migrate 命令
+//
+// Example commands:
+// 示例命令:
+//
+// Create migration scripts:
+// 创建迁移脚本:
+// ./bin/demo2kratos migrate new-script create --version-type TIME --description create_table
+// ./bin/demo2kratos migrate new-script create --version-type TIME --description alter_schema
+// ./bin/demo2kratos migrate new-script create --version-type TIME --description alter_schema --allow-empty-script true
+// ./bin/demo2kratos migrate new-script create --version-type TIME --description alter_column
+//
+// Update migration scripts:
+// 更新迁移脚本:
+// ./bin/demo2kratos migrate new-script update
+//
+// Execute migrations:
+// 执行迁移:
+// ./bin/demo2kratos migrate migrate all
+// ./bin/demo2kratos migrate migrate inc
+//
+// Preview migrations:
+// 预览迁移:
+// ./bin/demo2kratos migrate preview inc
+//
+// Check migration status:
+// 检查迁移状态:
+// ./bin/demo2kratos migrate status
+//
+// Note: Use caution with rollback operations to avoid unintended destructive actions
+// 注意: 回退操作要谨慎，避免误操作导致问题
+// ./bin/demo2kratos migrate migrate dec (use with caution)
+func NewMigrateCmd(logger log.Logger) *cobra.Command {
+	var rootCmd = &cobra.Command{
+		Use:   "migrate",
+		Short: "migrate",
+		Long:  "migrate",
+	}
+
+	const scriptsInRoot = "./scripts"
+
+	migrationparam.SetDebugMode(true)
+	param := migrationparam.NewMigrationParam(
+		func() *gorm.DB {
+			cfg := appcfg.ParseConfig(cfgpath.ConfigPath)
+			dsn := must.Nice(cfg.Data.Database.Source)
+			db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{}))
+			return db
+		},
+		func(db *gorm.DB) *migrate.Migrate {
+			rawDB := rese.P1(db.DB())
+			migrationDriver := rese.V1(sqlite3migrate.WithInstance(rawDB, &sqlite3migrate.Config{}))
+			return rese.P1(newmigrate.NewWithScriptsAndDatabase(
+				&newmigrate.ScriptsAndDatabaseParam{
+					ScriptsInRoot:    scriptsInRoot,
+					DatabaseName:     "sqlite3",
+					DatabaseInstance: migrationDriver,
+				},
+			))
+		},
+	)
+	rootCmd.AddCommand(newscripts.NewScriptCmd(&newscripts.Config{
+		Param:   param,
+		Options: newscripts.NewOptions(scriptsInRoot),
+		Objects: models.Objects(),
+	}))
+	rootCmd.AddCommand(cobramigration.NewMigrateCmd(param))
+	rootCmd.AddCommand(previewmigrate.NewPreviewCmd(param, scriptsInRoot))
+	rootCmd.AddCommand(migrationstate.NewStatusCmd(&migrationstate.Config{
+		Param:       param,
+		ScriptsPath: scriptsInRoot,
+		Objects:     models.Objects(),
+	}))
+
+	return rootCmd
+}
```

## configs/config.yaml (+3 -2)

```diff
@@ -5,10 +5,11 @@
   grpc:
     addr: 0.0.0.0:29000
     timeout: 1s
+  auto_run: true
 data:
   database:
-    driver: mysql
-    source: root:root@tcp(127.0.0.1:3306)/test?parseTime=True&loc=Local
+    driver: sqlite
+    source: ./bin/demo2kratos.db
   redis:
     addr: 127.0.0.1:6379
     read_timeout: 0.2s
```

## internal/conf/conf.pb.go (+11 -2)

```diff
@@ -78,6 +78,7 @@
 	state         protoimpl.MessageState `protogen:"open.v1"`
 	Http          *Server_HTTP           `protobuf:"bytes,1,opt,name=http,proto3" json:"http,omitempty"`
 	Grpc          *Server_GRPC           `protobuf:"bytes,2,opt,name=grpc,proto3" json:"grpc,omitempty"`
+	AutoRun       bool                   `protobuf:"varint,3,opt,name=auto_run,json=autoRun,proto3" json:"auto_run,omitempty"`
 	unknownFields protoimpl.UnknownFields
 	sizeCache     protoimpl.SizeCache
 }
@@ -126,6 +127,13 @@
 	return nil
 }
 
+func (x *Server) GetAutoRun() bool {
+	if x != nil {
+		return x.AutoRun
+	}
+	return false
+}
+
 type Data struct {
 	state         protoimpl.MessageState `protogen:"open.v1"`
 	Database      *Data_Database         `protobuf:"bytes,1,opt,name=database,proto3" json:"database,omitempty"`
@@ -426,10 +434,11 @@
 	"kratos.api\x1a\x1egoogle/protobuf/duration.proto\"]\n" +
 	"\tBootstrap\x12*\n" +
 	"\x06server\x18\x01 \x01(\v2\x12.kratos.api.ServerR\x06server\x12$\n" +
-	"\x04data\x18\x02 \x01(\v2\x10.kratos.api.DataR\x04data\"\xb8\x02\n" +
+	"\x04data\x18\x02 \x01(\v2\x10.kratos.api.DataR\x04data\"\xd3\x02\n" +
 	"\x06Server\x12+\n" +
 	"\x04http\x18\x01 \x01(\v2\x17.kratos.api.Server.HTTPR\x04http\x12+\n" +
-	"\x04grpc\x18\x02 \x01(\v2\x17.kratos.api.Server.GRPCR\x04grpc\x1ai\n" +
+	"\x04grpc\x18\x02 \x01(\v2\x17.kratos.api.Server.GRPCR\x04grpc\x12\x19\n" +
+	"\bauto_run\x18\x03 \x01(\bR\aautoRun\x1ai\n" +
 	"\x04HTTP\x12\x18\n" +
 	"\anetwork\x18\x01 \x01(\tR\anetwork\x12\x12\n" +
 	"\x04addr\x18\x02 \x01(\tR\x04addr\x123\n" +
```

## internal/conf/conf.proto (+1 -0)

```diff
@@ -23,6 +23,7 @@
   }
   HTTP http = 1;
   GRPC grpc = 2;
+  bool auto_run = 3;
 }
 
 message Data {
```

## internal/data/data.go (+27 -2)

```diff
@@ -2,8 +2,15 @@
 
 import (
 	"github.com/go-kratos/kratos/v2/log"
+	"github.com/go-xlan/go-migrate/checkmigration"
 	"github.com/google/wire"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/models"
+	"github.com/yyle88/must"
+	"github.com/yyle88/rese"
+	"gorm.io/driver/sqlite"
+	"gorm.io/gorm"
+	loggergorm "gorm.io/gorm/logger"
 )
 
 // ProviderSet is data providers.
@@ -11,13 +18,31 @@
 
 // Data .
 type Data struct {
-	// TODO wrapped database client
+	db *gorm.DB
 }
 
 // NewData .
 func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
+	dsn := must.Nice(c.Database.Source)
+	db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
+		Logger: loggergorm.Default.LogMode(loggergorm.Info),
+	}))
+
+	// Check if migration scripts are missing
+	// 检查是否缺少迁移脚本
+	checkmigration.CheckMigrate(db, models.Objects())
+
 	cleanup := func() {
 		log.NewHelper(logger).Info("closing the data resources")
+		must.Done(rese.P1(db.DB()).Close())
 	}
-	return &Data{}, cleanup, nil
+	return &Data{
+		db: db,
+	}, cleanup, nil
+}
+
+// DB returns the gorm database instance
+// 返回 gorm 数据库实例
+func (d *Data) DB() *gorm.DB {
+	return d.db
 }
```

## internal/data/greeter.go (+33 -2)

```diff
@@ -3,8 +3,11 @@
 import (
 	"context"
 
+	"github.com/go-kratos/kratos/v2/errors"
 	"github.com/go-kratos/kratos/v2/log"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/models"
+	"gorm.io/gorm"
 )
 
 type greeterRepo struct {
@@ -21,6 +24,22 @@
 }
 
 func (r *greeterRepo) Save(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
+	db := r.data.DB()
+
+	// Use GORM transaction to save record
+	// 使用 GORM 事务保存记录
+	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
+		record := &models.Record{
+			Message: g.Hello,
+		}
+		if err := tx.Create(record).Error; err != nil {
+			return err
+		}
+		return nil
+	})
+	if err != nil {
+		return nil, errors.New(500, "DB_ERROR", err.Error())
+	}
 	return g, nil
 }
 
@@ -28,8 +47,20 @@
 	return g, nil
 }
 
-func (r *greeterRepo) FindByID(context.Context, int64) (*biz.Greeter, error) {
-	return nil, nil
+func (r *greeterRepo) FindByID(ctx context.Context, id int64) (*biz.Greeter, error) {
+	db := r.data.DB()
+
+	var record models.Record
+	if err := db.WithContext(ctx).First(&record, id).Error; err != nil {
+		if errors.Is(err, gorm.ErrRecordNotFound) {
+			return nil, errors.New(404, "RECORD_NOT_FOUND", err.Error())
+		}
+		return nil, errors.New(500, "DB_ERROR", err.Error())
+	}
+
+	return &biz.Greeter{
+		Hello: record.Message,
+	}, nil
 }
 
 func (r *greeterRepo) ListByHello(context.Context, string) ([]*biz.Greeter, error) {
```

## internal/pkg/appcfg/app_cfg.go (+29 -0)

```diff
@@ -0,0 +1,29 @@
+package appcfg
+
+import (
+	"github.com/go-kratos/kratos/v2/config"
+	"github.com/go-kratos/kratos/v2/config/file"
+	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
+	"github.com/yyle88/rese"
+)
+
+// ParseConfig parses config file and returns Bootstrap config
+// 解析配置文件并返回 Bootstrap 配置
+func ParseConfig(configPath string) *conf.Bootstrap {
+	c := config.New(
+		config.WithSource(
+			file.NewSource(configPath),
+		),
+	)
+	defer rese.F0(c.Close)
+
+	if err := c.Load(); err != nil {
+		panic(err)
+	}
+
+	var cfg conf.Bootstrap
+	if err := c.Scan(&cfg); err != nil {
+		panic(err)
+	}
+	return &cfg
+}
```

## internal/pkg/models/article.go (+19 -0)

```diff
@@ -0,0 +1,19 @@
+package models
+
+import "gorm.io/gorm"
+
+// Article represents a blog article
+// 文章模型
+type Article struct {
+	gorm.Model
+	Title   string `gorm:"type:varchar(200);not null"`
+	Content string `gorm:"type:text"`
+	Author  string `gorm:"type:varchar(100)"`
+	Status  string `gorm:"type:varchar(20);default:'draft'"` // draft, published, archived // 草稿、已发布、已归档
+}
+
+// TableName returns the table name
+// 返回表名
+func (*Article) TableName() string {
+	return "articles"
+}
```

## internal/pkg/models/objects.go (+11 -0)

```diff
@@ -0,0 +1,11 @@
+package models
+
+// Objects returns all GORM model objects for migration
+// 返回所有用于迁移的 GORM 模型对象
+func Objects() []any {
+	return []any{
+		&Record{},
+		&Article{},
+		&Product{},
+	}
+}
```

## internal/pkg/models/product.go (+19 -0)

```diff
@@ -0,0 +1,19 @@
+package models
+
+import "gorm.io/gorm"
+
+// Product represents a product item
+// 产品模型
+type Product struct {
+	gorm.Model
+	Name        string  `gorm:"type:varchar(150);not null"`
+	Price       float64 `gorm:"type:decimal(10,2);not null"`
+	Stock       int     `gorm:"type:int;default:0"`
+	Description string  `gorm:"type:text"`
+}
+
+// TableName returns the table name
+// 返回表名
+func (*Product) TableName() string {
+	return "products"
+}
```

## internal/pkg/models/record.go (+16 -0)

```diff
@@ -0,0 +1,16 @@
+package models
+
+import "gorm.io/gorm"
+
+// Record represents a database record
+// 数据库记录模型
+type Record struct {
+	gorm.Model
+	Message string `gorm:"type:varchar(255)"`
+}
+
+// TableName returns the table name
+// 返回表名
+func (*Record) TableName() string {
+	return "records"
+}
```

## scripts/20251110105615_create_table.down.sql (+5 -0)

```diff
@@ -0,0 +1,5 @@
+-- reverse -- CREATE INDEX `idx_records_deleted_at` ON `records`(`deleted_at`);
+DROP INDEX IF EXISTS `idx_records_deleted_at`;
+
+-- reverse -- CREATE TABLE `records` (`id` integer PRIMARY KEY AUTOINCREMENT,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`message` varchar(255));
+DROP TABLE IF EXISTS `records`;
```

## scripts/20251110105615_create_table.up.sql (+10 -0)

```diff
@@ -0,0 +1,10 @@
+CREATE TABLE `records`
+(
+    `id`         integer PRIMARY KEY AUTOINCREMENT,
+    `created_at` datetime,
+    `updated_at` datetime,
+    `deleted_at` datetime,
+    `message`    varchar(255)
+);
+
+CREATE INDEX `idx_records_deleted_at` ON `records` (`deleted_at`);
```

## scripts/20251110110357_create_table.down.sql (+5 -0)

```diff
@@ -0,0 +1,5 @@
+-- reverse -- CREATE INDEX `idx_articles_deleted_at` ON `articles`(`deleted_at`);
+DROP INDEX IF EXISTS `idx_articles_deleted_at`;
+
+-- reverse -- CREATE TABLE `articles` (`id` integer PRIMARY KEY AUTOINCREMENT,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`title` varchar(200) NOT NULL,`content` text,`author` varchar(100),`status` varchar(20) DEFAULT "draft");
+DROP TABLE IF EXISTS `articles`;
```

## scripts/20251110110357_create_table.up.sql (+13 -0)

```diff
@@ -0,0 +1,13 @@
+CREATE TABLE `articles`
+(
+    `id`         integer PRIMARY KEY AUTOINCREMENT,
+    `created_at` datetime,
+    `updated_at` datetime,
+    `deleted_at` datetime,
+    `title`      varchar(200) NOT NULL,
+    `content`    text,
+    `author`     varchar(100),
+    `status`     varchar(20) DEFAULT "draft"
+);
+
+CREATE INDEX `idx_articles_deleted_at` ON `articles` (`deleted_at`);
```

## scripts/20251110110536_create_table.down.sql (+5 -0)

```diff
@@ -0,0 +1,5 @@
+-- reverse -- CREATE INDEX `idx_products_deleted_at` ON `products`(`deleted_at`);
+DROP INDEX IF EXISTS `idx_products_deleted_at`;
+
+-- reverse -- CREATE TABLE `products` (`id` integer PRIMARY KEY AUTOINCREMENT,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`name` varchar(150) NOT NULL,`price` decimal(10,2) NOT NULL,`stock` integer DEFAULT 0,`description` text);
+DROP TABLE IF EXISTS `products`;
```

## scripts/20251110110536_create_table.up.sql (+13 -0)

```diff
@@ -0,0 +1,13 @@
+CREATE TABLE `products`
+(
+    `id`          integer PRIMARY KEY AUTOINCREMENT,
+    `created_at`  datetime,
+    `updated_at`  datetime,
+    `deleted_at`  datetime,
+    `name`        varchar(150)   NOT NULL,
+    `price`       decimal(10, 2) NOT NULL,
+    `stock`       integer DEFAULT 0,
+    `description` text
+);
+
+CREATE INDEX `idx_products_deleted_at` ON `products` (`deleted_at`);
```

