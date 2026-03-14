// Package api 提供 HTTP API 处理函数
//
// 该包定义应用程序的 HTTP API 处理函数。
// 主要功能包括：
//   - 健康检查接口
//
// API 结构:
//   - /heatch_check: 健康检查接口
//   - /api/v1/*: 业务 API 接口
package api

import (
	"github.com/gin-gonic/gin"

	"gin-mini-agent/models"
)

// HeathCheck 健康检查接口
//
// 该函数处理健康检查请求，用于负载均衡和监控。
// 返回成功状态表示服务正常运行。
//
// 请求方法: GET
// 请求路径: /heatch_check
//
// 响应示例:
//
//	{
//	    "request_id": "abc123",
//	    "success": true,
//	    "data": "操作成功",
//	    "msg": "健康检查完成",
//	    "total": 0
//	}
//
// 使用场景:
//   - 负载均衡器健康检查
//   - Kubernetes 存活探针
//   - 监控系统状态检测
func HeathCheck(c *gin.Context) {
	models.OkWithDetailed("健康检查完成", models.CustomError[models.Ok], c)
}
