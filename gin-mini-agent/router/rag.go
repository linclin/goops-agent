package router

import (
	"gin-mini-agent/api/v1/rag"
	"gin-mini-agent/pkg/global"

	"github.com/gin-gonic/gin"
)

func InitRagRouter(r *gin.RouterGroup) (R gin.IRoutes) {
	router := r.Group("rag", gin.BasicAuth(gin.Accounts{
		global.Conf.Auth.User: global.Conf.Auth.Password,
	}))
	{
		router.POST("/index", rag.RagIndex)
	}
	return router
}
