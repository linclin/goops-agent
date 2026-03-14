// Package middleware 提供 HTTP 中间件功能
//
// 该包提供应用程序的 HTTP 中间件，包括：
//   - 访问日志记录
//   - 全局异常处理
//
// 中间件执行顺序:
//  1. CORS: 跨域处理
//  2. RequestID: 请求 ID 生成
//  3. Slog: 结构化日志
//  4. AccessLog: 访问日志记录
//  5. Exception: 异常捕获
package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	"gin-mini-agent/pkg/global"
)

// AccessLogWriter 访问日志写入器
//
// 该结构体包装了 gin.ResponseWriter，用于捕获响应内容。
// 实现了 Write 和 WriteString 方法，同时写入缓冲区和原始响应写入器。
//
// 功能特点:
//   - 捕获响应内容到缓冲区
//   - 不影响原始响应输出
//   - 支持日志记录响应内容
type AccessLogWriter struct {
	// ResponseWriter 原始响应写入器
	gin.ResponseWriter

	// body 响应内容缓冲区
	body *bytes.Buffer
}

// Write 实现 io.Writer 接口
//
// 该方法将数据同时写入缓冲区和原始响应写入器。
//
// 参数:
//   - p: 要写入的字节数据
//
// 返回:
//   - int: 写入的字节数
//   - error: 写入错误
func (w AccessLogWriter) Write(p []byte) (int, error) {
	// 先写入缓冲区
	if n, err := w.body.Write(p); err != nil {
		global.Log.Info("AccessLogWriter Write", slog.String("error", err.Error()))
		return n, err
	}
	// 再写入原始响应写入器
	return w.ResponseWriter.Write(p)
}

// WriteString 写入字符串
//
// 该方法将字符串同时写入缓冲区和原始响应写入器。
//
// 参数:
//   - p: 要写入的字符串
//
// 返回:
//   - int: 写入的字节数
//   - error: 写入错误
func (w AccessLogWriter) WriteString(p string) (int, error) {
	// 先写入缓冲区
	if n, err := w.body.WriteString(p); err != nil {
		global.Log.Info("AccessLogWriter WriteString", slog.String("error", err.Error()))
		return n, err
	}
	// 再写入原始响应写入器
	return w.ResponseWriter.WriteString(p)
}

// AccessLog 访问日志中间件
//
// 该中间件记录每个请求的详细信息，包括：
//   - 请求 ID
//   - 请求方法
//   - 请求 URI
//   - 请求体
//   - 响应体
//   - 状态码
//   - 执行时间
//   - 客户端 IP
//
// 使用示例:
//
//	r.Use(middleware.AccessLog)
//
// 日志格式:
//
//	{
//	    "level": "INFO",
//	    "msg": "接口访问日志",
//	    "requestId": "abc123",
//	    "method": "POST",
//	    "uri": "/api/v1/agent/chat",
//	    "reqBody": "{\"query\":\"hello\"}",
//	    "respBody": "...",
//	    "statusCode": 200,
//	    "execTime": "1.234s",
//	    "clientIP": "127.0.0.1"
//	}
func AccessLog(c *gin.Context) {
	// 创建响应写入器包装
	bodyWriter := &AccessLogWriter{
		body:           bytes.NewBufferString(""),
		ResponseWriter: c.Writer,
	}
	c.Writer = bodyWriter

	// 读取请求体
	// 注意：需要读取后重新设置，否则后续处理无法获取
	var bodyBytes []byte
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		global.Log.Info("请求体获取错误", slog.String("error", err.Error()))
	}

	// 重新设置请求体
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 获取请求 ID
	requestId := requestid.Get(c)
	c.Set("RequestId", requestId)

	// 记录开始时间
	startTime := time.Now()

	// 处理请求
	c.Next()

	// 记录结束时间
	endTime := time.Now()

	// 计算执行时间
	execTime := endTime.Sub(startTime)

	// 获取请求信息
	reqMethod := c.Request.Method      // 请求方法
	reqUri := c.Request.RequestURI     // 请求 URI
	reqBody := string(bodyBytes)       // 请求体
	statusCode := c.Writer.Status()    // 状态码
	respBody := bodyWriter.body.String() // 响应体
	clientIP := c.ClientIP()           // 客户端 IP

	// 记录访问日志
	global.Log.Info("接口访问日志",
		slog.String("requestId", requestId),
		slog.String("method", reqMethod),
		slog.String("uri", reqUri),
		slog.String("reqBody", reqBody),
		slog.String("respBody", respBody),
		slog.Int("statusCode", statusCode),
		slog.Duration("execTime", execTime),
		slog.String("clientIP", clientIP),
	)
}
