package middleware

import (
	"gin-mini-agent/models"
	"gin-mini-agent/pkg/global"
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// 全局异常处理中间件
func Exception(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			// 将异常写入日志
			global.Log.Error("未知panic异常",
				slog.Any("error", err),
				slog.String("stack", string(debug.Stack())),
			)
			models.FailWithDetailed("未知panic异常", models.CustomError[models.InternalServerError], c)
			c.Abort()
			return
		}
	}()
	c.Next()
}
