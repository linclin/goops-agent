package rag

import (
	"gin-mini-agent/internal/rag_index"
	"gin-mini-agent/models"

	"github.com/gin-gonic/gin"
)

// RagIndexRequest 索引请求参数
type RagIndexRequest struct {
	Dir    string `json:"dir" binding:"required"`
	DbType string `json:"dbType"`
}

// RagIndex 执行 RAG 索引
func RagIndex(c *gin.Context) {
	var req RagIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	var err error
	if req.DbType != "" {
		err = rag_index.RagIndexWithType(c.Request.Context(), req.Dir, req.DbType)
	} else {
		err = rag_index.RagIndex(c.Request.Context())
	}

	if err != nil {
		models.FailWithMessage("索引失败: "+err.Error(), c)
		return
	}

	models.OkWithData(gin.H{
		"message": "索引成功",
		"dir":     req.Dir,
		"dbType":  req.DbType,
	}, c)
}
