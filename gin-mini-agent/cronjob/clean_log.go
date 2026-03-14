// Package cronjob 提供定时任务功能
//
// 该包定义应用程序的定时任务。
// 主要功能包括：
//   - 日志文件清理
//
// 任务调度:
//   - 由 initialize/cron.go 中的 Cron 函数初始化
//   - 使用 robfig/cron 库进行调度
package cronjob

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"gin-mini-agent/pkg/global"
)

// CleanLog 清理过期日志文件的定时任务
//
// 该结构体实现了 cron.Job 接口，用于定时清理过期的日志文件。
// 清理规则根据配置文件中的 logs.max_age 参数确定。
//
// 功能特点:
//   - 自动清理超过保留天数的日志文件
//   - 支持清理 .log 和 .log.gz 文件
//   - 记录清理统计信息
//   - 错误恢复机制
//
// 配置示例:
//
//	logs:
//	  path: "./logs"
//	  max_age: 7  # 保留 7 天
//
// 调度配置:
//   - 执行时间: 每天凌晨 1 点
//   - Cron 表达式: "0 0 1 * * *"
type CleanLog struct{}

// Run 执行日志清理任务
//
// 该方法实现了 cron.Job 接口，由定时调度器调用。
//
// 执行流程:
//  1. 记录任务开始时间
//  2. 计算过期时间点
//  3. 遍历日志目录，删除过期文件
//  4. 记录清理统计信息
//
// 错误处理:
//   - 使用 defer + recover 捕获 panic
//   - 单个文件删除失败不影响其他文件
//   - 记录所有错误和警告
//
// 日志输出:
//   - 任务开始: INFO 级别
//   - 任务完成: INFO 级别，包含删除数量和耗时
//   - 任务失败: ERROR 级别
//   - 文件删除: DEBUG 级别
func (u CleanLog) Run() {
	// 记录任务开始时间
	startTime := time.Now()
	global.Log.Info("开始执行日志清理任务", slog.String("startTime", startTime.Format("2006-01-02 15:04:05")))

	// 延迟捕获 panic
	defer func() {
		if panicErr := recover(); panicErr != nil {
			global.Log.Error("cronjob定时任务:CleanLog执行失败",
				slog.Any("error", panicErr),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()

	// 获取配置中的日志路径和保留天数
	logPath := global.Conf.Logs.Path
	maxAge := global.Conf.Logs.MaxAge // 配置中设置的最大保留天数

	// 计算过期时间
	// 早于此时间的文件将被删除
	expirationTime := time.Now().AddDate(0, 0, -maxAge)

	// 清理日志文件
	deletedCount, err := u.deleteExpiredLogFiles(logPath, expirationTime)
	if err != nil {
		global.Log.Error("日志清理任务执行失败", slog.String("error", err.Error()))
		return
	}

	// 记录清理统计信息
	global.Log.Info("日志清理任务完成",
		slog.Int("deletedCount", deletedCount),
		slog.String("logPath", logPath),
		slog.Int("maxAge", maxAge),
		slog.Duration("duration", time.Since(startTime)),
	)
}

// deleteExpiredLogFiles 删除过期的日志文件
//
// 该方法遍历日志目录，删除修改时间早于过期时间的日志文件。
//
// 参数:
//   - logPath: 日志目录路径
//   - expirationTime: 过期时间点
//
// 返回:
//   - int: 删除的文件数量
//   - error: 遍历过程中的错误
//
// 处理的文件类型:
//   - .log: 普通日志文件
//   - .log.gz: 压缩的日志文件
//
// 注意事项:
//   - 跳过目录
//   - 单个文件删除失败不影响其他文件
//   - 记录所有删除操作
func (u CleanLog) deleteExpiredLogFiles(logPath string, expirationTime time.Time) (int, error) {
	deletedCount := 0

	// 遍历日志目录
	err := filepath.Walk(logPath, func(path string, info os.FileInfo, err error) error {
		// 处理访问错误
		if err != nil {
			global.Log.Warn("访问日志文件失败", slog.String("path", path), slog.String("error", err.Error()))
			return nil // 继续遍历其他文件
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件扩展名，只处理日志文件
		ext := filepath.Ext(path)
		if ext != ".log" && ext != ".log.gz" {
			return nil
		}

		// 检查文件是否过期
		// 使用文件修改时间判断
		if info.ModTime().Before(expirationTime) {
			// 删除过期文件
			err := os.Remove(path)
			if err != nil {
				global.Log.Warn("删除日志文件失败", slog.String("path", path), slog.String("error", err.Error()))
				return nil // 继续处理其他文件
			}

			// 统计删除数量
			deletedCount++

			// 记录删除操作
			global.Log.Debug("已删除过期日志文件", slog.String("path", path))
		}

		return nil
	})

	return deletedCount, err
}
