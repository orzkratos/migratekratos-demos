package appcfg

import (
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
	"github.com/yyle88/rese"
)

// ParseConfig parses config file and returns Bootstrap config
// 解析配置文件并返回 Bootstrap 配置
func ParseConfig(configPath string) *conf.Bootstrap {
	c := config.New(
		config.WithSource(
			file.NewSource(configPath),
		),
	)
	defer rese.F0(c.Close)

	if err := c.Load(); err != nil {
		panic(err)
	}

	var cfg conf.Bootstrap
	if err := c.Scan(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
