package initialize

import (
	"fmt"
	"gin-mini-agent/cronjob"
	"gin-mini-agent/pkg/global"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// 初始化定时任务
func Cron() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(fmt.Sprintf("初始化Cron定时任务失败: %v", err))
	}
	logger := log.Default()
	c := cron.New(cron.WithLocation(loc), cron.WithSeconds(), cron.WithLogger(cron.PrintfLogger(logger)))
	//清理超过一周的日志文件
	c.AddJob("0 0 1 * * *", cron.NewChain(cron.Recover(cron.PrintfLogger(logger)), cron.SkipIfStillRunning(cron.PrintfLogger(logger))).Then(&cronjob.CleanLog{}))
	c.Start()
	global.Log.Info("初始化Cron定时任务完成")
}
