package initialize

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"gin-mini-agent/cronjob"
	"gin-mini-agent/pkg/global"
)

// Cron 初始化定时任务
//
// 该函数初始化应用程序的定时任务调度器。
// 使用 robfig/cron 库实现定时任务管理。
//
// 时区配置:
//   - 使用 Asia/Shanghai 时区（北京时间）
//
// 调度器配置:
//   - WithLocation: 设置时区
//   - WithSeconds: 支持秒级调度（6 位 cron 表达式）
//   - WithLogger: 配置日志输出
//
// 定时任务:
//   - CleanLog: 每天凌晨 1 点清理超过一周的日志文件
//
// Cron 表达式格式（6 位）:
//
//	秒 分 时 日 月 周
//	0  0  1  *  *  *  // 每天凌晨 1 点执行
//
// 任务链配置:
//   - Recover: 任务 panic 时自动恢复
//   - SkipIfStillRunning: 如果上一次任务还在运行，跳过本次执行
//
// 注意事项:
//   - 如果时区加载失败，会触发 panic
func Cron() {
	// 加载时区
	// 使用北京时间（Asia/Shanghai）
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(fmt.Sprintf("初始化Cron定时任务失败: %v", err))
	}

	// 创建日志记录器
	logger := log.Default()

	// 创建 Cron 调度器
	// 配置:
	// - 时区: Asia/Shanghai
	// - 秒级调度: 支持 6 位 cron 表达式
	// - 日志: 使用默认日志记录器
	c := cron.New(
		cron.WithLocation(loc),
		cron.WithSeconds(),
		cron.WithLogger(cron.PrintfLogger(logger)),
	)

	// 添加定时任务
	// 表达式: "0 0 1 * * *" 表示每天凌晨 1 点执行
	// 任务链:
	// - Recover: 捕获 panic，防止任务崩溃影响调度器
	// - SkipIfStillRunning: 如果上次任务未完成，跳过本次执行
	c.AddJob(
		"0 0 1 * * *",
		cron.NewChain(
			cron.Recover(cron.PrintfLogger(logger)),
			cron.SkipIfStillRunning(cron.PrintfLogger(logger)),
		).Then(&cronjob.CleanLog{}),
	)

	// 启动调度器
	c.Start()

	global.Log.Info("初始化Cron定时任务完成")
}
