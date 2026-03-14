package initialize

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"

	"gin-mini-agent/api"
	"gin-mini-agent/middleware"
	"gin-mini-agent/pkg/global"
	"gin-mini-agent/router"
)

// Routers 初始化总路由
//
// 该函数创建并配置 Gin 路由引擎，注册所有中间件和路由。
//
// 返回:
//   - *gin.Engine: 配置完成的 Gin 引擎实例
//
// 中间件顺序（按执行顺序）:
//  1. CORS: 跨域资源共享中间件
//  2. RequestID: 请求 ID 中间件，为每个请求生成唯一 ID
//  3. Slog: 结构化日志中间件
//  4. AccessLog: 访问日志中间件
//  5. Exception: 全局异常处理中间件
//
// 路由结构:
//   - /heatch_check: 健康检查接口
//   - /api/v1/*: API 路由组
//     - /rag/*: RAG 相关路由
//     - /agent/*: Agent 相关路由
//
// CORS 配置:
//   - AllowOrigins: 允许所有来源（*）
//   - AllowMethods: GET, POST, PUT, DELETE, OPTIONS
//   - AllowHeaders: Origin, Content-Type, Authorization
//   - AllowCredentials: true
//
// 运行模式:
//   - prd: 生产模式，禁用调试输出
//   - 其他: 开发模式，启用调试输出
func Routers() *gin.Engine {
	// 设置运行模式
	if global.Conf.System.RunMode == "prd" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 强制启用控制台颜色输出
	gin.ForceConsoleColor()

	// 创建 Gin 引擎（不带默认中间件）
	r := gin.New()

	// 添加 CORS 中间件
	// 允许跨域请求
	r.Use(cors.New(cors.Config{
		// AllowOrigins 允许的来源
		// "*" 表示允许所有来源
		AllowOrigins: []string{"*"},

		// AllowMethods 允许的 HTTP 方法
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},

		// AllowHeaders 允许的请求头
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},

		// ExposeHeaders 暴露给客户端的响应头
		ExposeHeaders: []string{"Content-Length"},

		// AllowCredentials 是否允许携带凭证
		AllowCredentials: true,
	}))

	// 添加请求 ID 中间件
	// 为每个请求生成唯一 ID，便于追踪和调试
	r.Use(requestid.New())

	// 添加结构化日志中间件
	// 使用 slog 记录请求日志
	r.Use(sloggin.New(global.Log))

	// 添加访问日志中间件
	// 记录请求的详细信息
	r.Use(middleware.AccessLog)

	// 添加全局异常处理中间件
	// 捕获 panic 并返回友好的错误信息
	r.Use(middleware.Exception)

	// 注册健康检查接口
	// 用于负载均衡和监控
	r.GET("/heatch_check", api.HeathCheck)
	global.Log.Info("初始化健康检查接口完成")

	// 创建 API 路由组
	// 使用配置文件中的 URL 路径前缀
	apiGroup := r.Group(global.Conf.System.UrlPathPrefix)

	// 创建 v1 版本路由组
	// 方便统一添加路由前缀和版本控制
	v1Group := apiGroup.Group("v1")

	// 注册 RAG 路由
	// 包括知识库索引等接口
	router.InitRagRouter(v1Group)

	// 注册 Agent 路由
	// 包括对话接口等
	router.InitAgentRouter(v1Group)

	global.Log.Info("初始化基础路由完成")

	return r
}
