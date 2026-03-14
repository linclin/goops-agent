package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"gin-mini-agent/models"
	"gin-mini-agent/pkg/global"
)

// Exception 全局异常处理中间件
//
// 该中间件捕获请求处理过程中的 panic 异常，
// 返回友好的错误信息，避免向用户暴露敏感信息。
//
// 功能特点:
//   - 捕获所有 panic 异常
//   - 记录详细的错误日志和堆栈信息
//   - 返回统一的错误响应格式
//   - 阻止异常继续传播
//
// 使用示例:
//
//	r.Use(middleware.Exception)
//
// 错误响应格式:
//
//	{
//	    "request_id": "abc123",
//	    "success": false,
//	    "data": null,
//	    "msg": "未知panic异常",
//	    "total": 0
//	}
//
// 日志格式:
//
//	{
//	    "level": "ERROR",
//	    "msg": "未知panic异常",
//	    "error": "runtime error: invalid memory address",
//	    "stack": "goroutine 1 [running]:\n..."
//	}
func Exception(c *gin.Context) {
	// 延迟执行，捕获 panic
	defer func() {
		// 检查是否发生 panic
		if err := recover(); err != nil {
			// 将异常写入日志
			// 包含错误信息和堆栈跟踪
			global.Log.Error("未知panic异常",
				slog.Any("error", err),
				slog.String("stack", string(debug.Stack())),
			)

			// 返回友好的错误响应
			models.FailWithDetailed("未知panic异常", models.CustomError[models.InternalServerError], c)

			// 终止请求处理
			c.Abort()
			return
		}
	}()

	// 继续处理请求
	c.Next()
}
