package middleware

import (
	"bytes"
	"gin-mini-agent/pkg/global"
	"io"
	"log/slog"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type AccessLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w AccessLogWriter) Write(p []byte) (int, error) {
	if n, err := w.body.Write(p); err != nil {
		global.Log.Info("AccessLogWriter Write", slog.String("error", err.Error()))
		return n, err
	}
	return w.ResponseWriter.Write(p)
}

func (w AccessLogWriter) WriteString(p string) (int, error) {
	if n, err := w.body.WriteString(p); err != nil {
		global.Log.Info("AccessLogWriter WriteString", slog.String("error", err.Error()))
		return n, err
	}
	return w.ResponseWriter.WriteString(p)
}

// 访问日志
func AccessLog(c *gin.Context) {
	bodyWriter := &AccessLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = bodyWriter
	var bodyBytes []byte // 我们需要的body内容
	// 从原有Request.Body读取
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		global.Log.Info("请求体获取错误", slog.String("error", err.Error()))
	}
	// 新建缓冲区并替换原有Request.body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	requestId := requestid.Get(c)
	c.Set("RequestId", requestId)
	// 开始时间
	startTime := time.Now()
	// 处理请求
	c.Next()
	// 结束时间
	endTime := time.Now()
	// 执行时间
	execTime := endTime.Sub(startTime)
	// 请求方式
	reqMethod := c.Request.Method
	// 请求路由
	reqUri := c.Request.RequestURI
	// 请求体
	reqBody := string(bodyBytes)
	// 状态码
	statusCode := c.Writer.Status()
	// 返回体
	respBody := bodyWriter.body.String()
	// 请求IP
	clientIP := c.ClientIP()
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
