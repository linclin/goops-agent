package initialize

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/natefinch/lumberjack"

	"gin-mini-agent/pkg/global"
)

// DailyLogWriter 按天分割的日志写入器
//
// 该结构体实现了 io.Writer 接口，支持按天自动分割日志文件。
// 每天会创建一个新的日志文件，便于日志管理和归档。
//
// 功能特点:
//   - 按日期自动分割日志文件
//   - 支持日志文件大小限制
//   - 支持日志文件数量限制
//   - 支持日志文件压缩
//
// 使用示例:
//
//	writer := &DailyLogWriter{
//	    filename:   "/var/log/app-%s.log",
//	    maxSize:    100,  // MB
//	    maxBackups: 10,
//	    maxAge:     30,   // days
//	    compress:   true,
//	}
type DailyLogWriter struct {
	// filename 日志文件名格式
	// 使用 %s 作为日期占位符
	// 例如: /var/log/app-%s.log -> /var/log/app-2024-01-01.log
	filename string

	// maxSize 单个日志文件最大大小（MB）
	// 超过此大小会触发日志轮转
	maxSize int

	// maxBackups 保留的旧日志文件最大数量
	// 超过此数量会删除最旧的日志文件
	maxBackups int

	// maxAge 保留旧日志文件的最大天数
	// 超过此天数的日志文件会被删除
	maxAge int

	// compress 是否压缩旧日志文件
	// 使用 gzip 压缩，节省磁盘空间
	compress bool

	// currentLog 当前日志写入器
	currentLog *lumberjack.Logger

	// lastDate 上次写入的日期
	// 用于判断是否需要切换日志文件
	lastDate string
}

// Write 实现 io.Writer 接口
//
// 该方法写入日志数据，并自动检查是否需要切换到新的日志文件。
//
// 参数:
//   - p: 要写入的字节数据
//
// 返回:
//   - n: 写入的字节数
//   - err: 写入错误
func (w *DailyLogWriter) Write(p []byte) (n int, err error) {
	// 检查是否需要切换到新的日志文件
	w.rotateIfNeeded()
	return w.currentLog.Write(p)
}

// rotateIfNeeded 检查是否需要切换到新的日志文件
//
// 该方法检查当前日期是否与上次写入日期不同，
// 如果不同则关闭当前日志文件并创建新的日志文件。
//
// 切换条件:
//   - 日期发生变化（跨天）
//
// 切换流程:
//  1. 关闭当前日志文件
//  2. 使用新日期创建日志文件名
//  3. 创建新的 lumberjack.Logger 实例
func (w *DailyLogWriter) rotateIfNeeded() {
	// 获取当前日期
	currentDate := time.Now().Format("2006-01-02")

	// 检查日期是否变化
	if currentDate != w.lastDate {
		// 关闭当前日志文件
		if w.currentLog != nil {
			w.currentLog.Close()
		}

		// 创建新的日志文件名
		// 使用日期格式化文件名
		newFilename := fmt.Sprintf(w.filename, currentDate)

		// 创建新的日志写入器
		w.currentLog = &lumberjack.Logger{
			Filename:   newFilename,
			MaxSize:    w.maxSize,
			MaxBackups: w.maxBackups,
			MaxAge:     w.maxAge,
			Compress:   w.compress,
		}

		// 更新最后写入日期
		w.lastDate = currentDate
	}
}

// Logger 初始化日志系统
//
// 该函数初始化应用程序的日志系统，包括：
//   - 创建日志目录
//   - 配置日志格式（JSON）
//   - 配置日志输出（控制台 + 文件）
//   - 配置 panic 日志
//
// 日志配置:
//   - 格式: JSON
//   - 输出: 控制台 + 文件
//   - 分割: 按天分割
//   - 级别: 从配置文件读取
//
// Panic 日志:
//   - 单独存储到 panic.log 文件
//   - 用于捕获程序崩溃信息
//
// 注意事项:
//   - 如果创建日志目录失败，会触发 panic
//   - 如果创建 panic 日志失败，会触发 panic
func Logger() {
	// 创建日志目录
	if err := os.MkdirAll(global.Conf.Logs.Path, 0755); err != nil {
		panic(fmt.Sprintf("创建日志目录失败: %v", err))
	}

	// 配置按天分割的日志文件名格式
	// %s 会被替换为日期，如: app-2024-01-01.log
	fileNamePattern := filepath.Join(global.Conf.Logs.Path, global.Conf.System.AppName+"-%s.log")

	// 创建按天分割的日志写入器
	logWriter := &DailyLogWriter{
		filename:   fileNamePattern,
		maxSize:    global.Conf.Logs.MaxSize,
		maxBackups: global.Conf.Logs.MaxBackups,
		maxAge:     global.Conf.Logs.MaxAge,
		compress:   global.Conf.Logs.Compress,
	}

	// 配置日志处理器选项
	logOpts := slog.HandlerOptions{
		// AddSource 是否添加源码位置信息
		AddSource: true,
		// Level 日志级别
		Level: global.Conf.Logs.Level,
	}

	// 创建日志记录器
	// 输出到控制台和文件
	logger := slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stdout, logWriter), &logOpts))

	// 设置为默认日志记录器
	slog.SetDefault(logger)

	// 设置全局日志变量
	global.Log = logger
	global.Log.Info("初始化日志完成")

	// 创建 panic 日志文件
	// 用于捕获程序崩溃信息
	panicFile, err := os.Create(fmt.Sprintf("%s/panic.log", global.Conf.Logs.Path))
	if err != nil {
		global.Log.Info(fmt.Sprint("初始化panic日志完成错误", err.Error()))
		panic(err)
	}

	// 设置 panic 输出到文件
	debug.SetCrashOutput(panicFile, debug.CrashOptions{})
	global.Log.Info("初始化panic日志完成")
}
