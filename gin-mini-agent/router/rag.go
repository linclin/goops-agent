package router

import (
	"github.com/gin-gonic/gin"

	"gin-mini-agent/api/v1/rag"
	"gin-mini-agent/pkg/global"
)

// InitRagRouter 初始化 RAG 路由
//
// 该函数注册 RAG（检索增强生成）相关的 HTTP 路由。
// 所有路由都需要 Basic Auth 认证。
//
// 参数:
//   - r: Gin 路由组（通常是 /api/v1）
//
// 返回:
//   - gin.IRoutes: 注册的路由
//
// 路由列表:
//   - POST /rag/index: 知识库索引接口
//
// 认证配置:
//   - 使用 Basic Auth
//   - 用户名和密码从配置文件读取
//
// 使用示例:
//
//	v1Group := r.Group("v1")
//	router.InitRagRouter(v1Group)
func InitRagRouter(r *gin.RouterGroup) (R gin.IRoutes) {
	// 创建 RAG 路由组
	// 添加 Basic Auth 认证中间件
	router := r.Group("rag", gin.BasicAuth(gin.Accounts{
		global.Conf.Auth.User: global.Conf.Auth.Password,
	}))
	{
		// 知识库索引接口
		// 用于将文档添加到向量数据库
		router.POST("/index", rag.RagIndex)
	}
	return router
}
