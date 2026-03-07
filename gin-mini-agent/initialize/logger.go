package initialize

import (
	"fmt"
	"gin-mini-agent/pkg/global"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/natefinch/lumberjack"
)

// 按天分割的日志写入器
type DailyLogWriter struct {
	filename   string
	maxSize    int
	maxBackups int
	maxAge     int
	compress   bool
	currentLog *lumberjack.Logger
	lastDate   string
}

// Write 实现 io.Writer 接口
func (w *DailyLogWriter) Write(p []byte) (n int, err error) {
	// 检查是否需要切换到新的日志文件
	w.rotateIfNeeded()
	return w.currentLog.Write(p)
}

// rotateIfNeeded 检查是否需要切换到新的日志文件
func (w *DailyLogWriter) rotateIfNeeded() {
	currentDate := time.Now().Format("2006-01-02")
	if currentDate != w.lastDate {
		// 关闭当前日志文件
		if w.currentLog != nil {
			w.currentLog.Close()
		}
		// 创建新的日志文件
		newFilename := fmt.Sprintf(w.filename, currentDate)
		w.currentLog = &lumberjack.Logger{
			Filename:   newFilename,
			MaxSize:    w.maxSize,
			MaxBackups: w.maxBackups,
			MaxAge:     w.maxAge,
			Compress:   w.compress,
		}
		w.lastDate = currentDate
	}
}

// 初始化日志
func Logger() {
	// 创建日志目录
	if err := os.MkdirAll(global.Conf.Logs.Path, 0755); err != nil {
		panic(fmt.Sprintf("创建日志目录失败: %v", err))
	}

	// 按天分割的日志文件名格式
	fileNamePattern := filepath.Join(global.Conf.Logs.Path, global.Conf.System.AppName+"-%s.log")
	logWriter := &DailyLogWriter{
		filename:   fileNamePattern,
		maxSize:    global.Conf.Logs.MaxSize,
		maxBackups: global.Conf.Logs.MaxBackups,
		maxAge:     global.Conf.Logs.MaxAge,
		compress:   global.Conf.Logs.Compress,
	}

	logOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     global.Conf.Logs.Level,
	}
	logger := slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stdout, logWriter), &logOpts))
	slog.SetDefault(logger)
	global.Log = logger
	global.Log.Info("初始化日志完成")
	panicFile, err := os.Create(fmt.Sprintf("%s/panic.log", global.Conf.Logs.Path))
	if err != nil {
		global.Log.Info(fmt.Sprint("初始化panic日志完成错误", err.Error()))
		panic(err)
	}
	debug.SetCrashOutput(panicFile, debug.CrashOptions{})
	global.Log.Info("初始化panic日志完成")
}
