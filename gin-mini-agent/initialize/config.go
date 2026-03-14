// Package initialize 提供应用程序初始化功能
//
// 该包负责应用程序启动时的初始化工作。
// 主要功能包括：
//   - 配置文件加载和热更新
//   - 日志系统初始化
//   - 验证器初始化
//   - AI Agent 初始化
//   - 路由初始化
//   - 定时任务初始化
//
// 初始化顺序:
//  1. InitConfig: 加载配置文件
//  2. Logger: 初始化日志系统
//  3. Validate: 初始化验证器
//  4. Cron: 初始化定时任务
//  5. InitAiAgent: 初始化 AI Agent
//  6. Routers: 初始化路由
package initialize

import (
	"fmt"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"gin-mini-agent/pkg/global"
)

// 配置文件常量
const (
	// configType 配置文件类型
	configType = "yml"

	// configPath 配置文件路径
	configPath = "./conf"

	// devConfig 开发环境配置文件
	devConfig = "config.se.yml"

	// stageConfig 测试环境配置文件
	stageConfig = "config.st.yml"

	// prodConfig 生产环境配置文件
	prodConfig = "config.prd.yml"
)

// InitConfig 初始化配置文件
//
// 该函数负责加载应用程序配置文件，并支持热更新。
// 根据环境变量 RunMode 选择不同的配置文件。
//
// 环境变量:
//   - RunMode=dev: 使用 config.se.yml（开发环境）
//   - RunMode=st: 使用 config.st.yml（测试环境）
//   - RunMode=prd: 使用 config.prd.yml（生产环境）
//   - 默认: 使用 config.se.yml
//
// 配置文件格式: YAML
// 配置文件路径: ./conf/
//
// 热更新:
//   - 监听配置文件变更
//   - 文件修改后自动重新加载
//   - 无需重启服务即可生效
//
// 环境变量绑定:
//   - 支持通过环境变量覆盖配置
//   - 环境变量名使用下划线分隔，如: SYSTEM_PORT
//
// 注意事项:
//   - 如果配置文件加载失败，会触发 panic
//   - 配置变更时会打印日志
func InitConfig() {
	// 创建 Viper 实例
	v := viper.New()

	// 根据环境变量选择配置文件
	env := os.Getenv("RunMode")
	configName := devConfig
	switch env {
	case "st":
		configName = stageConfig
	case "prd":
		configName = prodConfig
	}

	// 设置配置文件信息
	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath(configPath)

	// 设置环境变量替换规则
	// 配置键中的 "." 替换为 "_"
	// 例如: system.port -> SYSTEM_PORT
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定环境变量
	// 自动读取与环境变量同名的配置
	v.AutomaticEnv()

	// 读取配置文件
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("初始化配置文件失败: %v", err))
	}

	// 将配置解析到结构体
	if err := v.Unmarshal(&global.Conf); err != nil {
		panic(fmt.Sprintf("初始化配置文件失败: %v", err))
	}

	// 监听配置文件变更，实现热更新
	v.WatchConfig()

	// 配置变更回调函数
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("配置文件:%s 发生变更:%s\n", e.Name, e.Op)

		// 重新解析配置到结构体
		if err := v.Unmarshal(&global.Conf); err != nil {
			panic(fmt.Sprintf("初始化配置文件失败: %v", err))
		}
	})
}
