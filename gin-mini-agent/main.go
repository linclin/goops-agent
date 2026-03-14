// Package main 应用程序入口
//
// 该包是应用程序的主入口点，负责启动 HTTP 服务器。
// 主要功能包括：
//   - 初始化应用程序组件
//   - 启动 HTTP 服务器
//   - 处理优雅关闭
//
// 构建信息:
//   - GitBranch: Git 分支名
//   - GitRevision: Git 提交哈希
//   - GitCommitLog: Git 提交信息
//   - BuildTime: 构建时间
//   - BuildGoVersion: Go 版本
//
// 构建命令:
//
//	go build -ldflags "-X main.GitBranch=$(git branch --show-current) \
//	  -X main.GitRevision=$(git rev-parse HEAD) \
//	  -X main.GitCommitLog=$(git log -1 --pretty=%s) \
//	  -X main.BuildTime=$(date +%Y-%m-%d_%H:%M:%S) \
//	  -X main.BuildGoVersion=$(go version)"
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"gin-mini-agent/initialize"
	"gin-mini-agent/pkg/global"
)

// 构建信息变量
// 初始化为 unknown，如果编译时没有传入这些值，则为 unknown
var (
	// GitBranch Git 分支名
	GitBranch = "unknown"

	// GitRevision Git 提交哈希
	GitRevision = "unknown"

	// GitCommitLog Git 提交信息
	GitCommitLog = "unknown"

	// BuildTime 构建时间
	BuildTime = "unknown"

	// BuildGoVersion Go 版本
	BuildGoVersion = "unknown"
)

// init 初始化函数
//
// 该函数在 main 函数之前执行，负责初始化应用程序组件。
//
// 初始化顺序:
//  1. 打印构建信息
//  2. 加载配置文件
//  3. 初始化日志系统
//  4. 初始化验证器
//  5. 初始化定时任务
//  6. 初始化 AI Agent
func init() {
	// 输出构建信息
	fmt.Fprint(os.Stdout, buildInfo())

	// 初始化配置文件
	initialize.InitConfig()

	// 初始化日志系统
	initialize.Logger()

	// 初始化验证器（中文）
	initialize.Validate("zh")

	// 初始化定时任务
	initialize.Cron()

	// 初始化 AI Agent
	initialize.InitAiAgent()
}

// main 主函数
//
// 该函数是应用程序的入口点，负责启动 HTTP 服务器和处理优雅关闭。
//
// 启动流程:
//  1. 初始化路由
//  2. 创建 HTTP 服务器
//  3. 在 goroutine 中启动服务器
//  4. 等待中断信号
//  5. 优雅关闭服务器
//
// 优雅关闭:
//   - 监听 SIGINT (Ctrl+C) 和 SIGTERM 信号
//   - 收到信号后等待 5 秒让正在处理的请求完成
//   - 强制关闭服务器
//
// 错误处理:
//   - 使用 defer + recover 捕获 panic
//   - 将错误信息写入日志
func main() {
	// 延迟捕获 panic
	defer func() {
		if err := recover(); err != nil {
			// 将异常写入日志
			global.Log.Error(fmt.Sprintf("项目启动失败: %v\n堆栈信息: %v", err, string(debug.Stack())))
		}
	}()

	// 初始化路由
	r := initialize.Routers()

	// 构建服务器地址
	host := "0.0.0.0"
	port := global.Conf.System.Port
	address := fmt.Sprintf("%s:%d", host, port)

	// 创建 HTTP 服务器
	// 参考地址: https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/server.go
	srv := &http.Server{
		Addr:    address,
		Handler: r,
	}

	// 在 goroutine 中启动服务器
	// 这样不会阻塞优雅关闭的处理
	go func() {
		global.Log.Info(fmt.Sprintf("HTTP Server is running at %s/%s", address, global.Conf.System.UrlPathPrefix))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			global.Log.Error(fmt.Sprint("listen error: ", err))
		}
	}()

	// 等待中断信号以优雅关闭服务器
	// 超时时间为 5 秒
	quit := make(chan os.Signal, 1)

	// 监听信号
	// kill (无参数) 默认发送 syscall.SIGTERM
	// kill -2 是 syscall.SIGINT
	// kill -9 是 syscall.SIGKILL 但无法捕获，所以不需要添加
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	global.Log.Info("Shutting down server...")

	// 创建超时上下文
	// 给服务器 5 秒时间完成正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		global.Log.Error(fmt.Sprint("Server forced to shutdown: ", err))
	}

	global.Log.Info("Server exiting")
}

// buildInfo 返回构建信息
//
// 该函数返回多行格式的构建信息字符串。
//
// 返回:
//   - string: 构建信息，包含分支、提交、时间、版本等
//
// 信息格式:
//
//	GitBranch=main
//	GitRevision=abc123...
//	GitCommitLog=feat: add new feature
//	BuildTime=2024-01-01_12:00:00
//	GoVersion=go1.21.0
//	runtime=linux/amd64
func buildInfo() string {
	return fmt.Sprintf("GitBranch=%s\nGitRevision=%s\nGitCommitLog=%s\nBuildTime=%s\nGoVersion=%s\nruntime=%s/%s\n",
		GitBranch, GitRevision, GitCommitLog, BuildTime, BuildGoVersion, runtime.GOOS, runtime.GOARCH)
}
