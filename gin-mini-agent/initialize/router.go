package initialize

import (
	"gin-mini-agent/api"
	"gin-mini-agent/middleware"
	"gin-mini-agent/pkg/global"
	"gin-mini-agent/router"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
)

// 初始化总路由
func Routers() *gin.Engine {
	if global.Conf.System.RunMode == "prd" {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.ForceConsoleColor()
	// 创建带有默认中间件的路由:
	// 日志与恢复中间件
	// r := gin.Default()
	// 创建不带中间件的路由:
	r := gin.New()
	// 初始化Trace中间件
	r.Use(requestid.New())
	// slog日志
	r.Use(sloggin.New(global.Log))
	// 添加访问记录
	r.Use(middleware.AccessLog)
	// 添加全局异常处理中间件
	r.Use(middleware.Exception)
	// 初始化健康检查接口
	r.GET("/heatch_check", api.HeathCheck)
	global.Log.Info("初始化健康检查接口完成")
	// 初始化API路由
	apiGroup := r.Group(global.Conf.System.UrlPathPrefix)
	// 方便统一添加路由前缀
	v1Group := apiGroup.Group("v1")
	router.InitRagRouter(v1Group) // 注册 RAG 路由
	global.Log.Info("初始化基础路由完成")
	return r
}
