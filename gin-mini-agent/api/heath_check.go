package api

import (
	"gin-mini-agent/models"

	"github.com/gin-gonic/gin"
)

func HeathCheck(c *gin.Context) {
	models.OkWithDetailed("健康检查完成", models.CustomError[models.Ok], c)
}
