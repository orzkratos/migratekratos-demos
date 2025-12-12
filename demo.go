// Package demokratos provides demo projects based on Kratos microservice framework
// Contains demo1kratos and demo2kratos as complete Kratos project examples
// Serves as template and reference when starting new Kratos projects
//
// demokratos 提供基于 Kratos 微服务框架的演示项目
// 包含 demo1kratos 和 demo2kratos 作为完整的 Kratos 项目示例
// 在启动新 Kratos 项目时可作为模板和参考
package demokratos

import (
	"github.com/orzkratos/demokratos/demo1kratos"
	"github.com/orzkratos/demokratos/demo2kratos"
)

// GetDemo1Path returns source root path of demo1kratos project
//
// GetDemo1Path 返回 demo1kratos 项目的源代码根目录路径
func GetDemo1Path() string {
	return demo1kratos.SourceRoot()
}

// GetDemo2Path returns source root path of demo2kratos project
//
// GetDemo2Path 返回 demo2kratos 项目的源代码根目录路径
func GetDemo2Path() string {
	return demo2kratos.SourceRoot()
}
