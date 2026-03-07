package cronjob

import (
	"gin-mini-agent/pkg/global"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

// 清理超过一周的日志文件
type CleanLog struct {
}

func (u CleanLog) Run() {
	startTime := time.Now()
	global.Log.Info("开始执行日志清理任务", slog.String("startTime", startTime.Format("2006-01-02 15:04:05")))
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
	expirationTime := time.Now().AddDate(0, 0, -maxAge)

	// 清理日志文件
	deletedCount, err := u.deleteExpiredLogFiles(logPath, expirationTime)
	if err != nil {
		global.Log.Error("日志清理任务执行失败", slog.String("error", err.Error()))
		return
	}

	global.Log.Info("日志清理任务完成",
		slog.Int("deletedCount", deletedCount),
		slog.String("logPath", logPath),
		slog.Int("maxAge", maxAge),
		slog.Duration("duration", time.Since(startTime)),
	)
}

// deleteExpiredLogFiles 删除过期的日志文件
func (u CleanLog) deleteExpiredLogFiles(logPath string, expirationTime time.Time) (int, error) {
	deletedCount := 0

	// 遍历日志目录
	err := filepath.Walk(logPath, func(path string, info os.FileInfo, err error) error {
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
		if info.ModTime().Before(expirationTime) {
			// 删除过期文件
			err := os.Remove(path)
			if err != nil {
				global.Log.Warn("删除日志文件失败", slog.String("path", path), slog.String("error", err.Error()))
				return nil // 继续处理其他文件
			}

			deletedCount++
			global.Log.Debug("已删除过期日志文件", slog.String("path", path))
		}

		return nil
	})

	return deletedCount, err
}
