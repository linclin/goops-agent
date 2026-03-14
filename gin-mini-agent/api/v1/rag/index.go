// Package rag 提供 RAG（检索增强生成）相关的 HTTP API
//
// 该包定义 RAG 相关的 HTTP 接口处理函数。
// 主要功能包括：
//   - 知识库索引接口
//
// RAG 简介:
//   RAG（Retrieval-Augmented Generation）是一种结合检索和生成的技术，
//   通过从知识库中检索相关文档来增强大模型的回答能力。
//
// API 路径:
//   - POST /api/v1/rag/index: 知识库索引接口
package rag

import (
	"github.com/gin-gonic/gin"

	"gin-mini-agent/internal/rag_index"
	"gin-mini-agent/models"
)

// RagIndexRequest 索引请求参数
//
// 该结构体定义了知识库索引接口的请求参数。
//
// 字段说明:
//   - Dir: 要索引的文档目录路径（必填）
//   - DbType: 向量数据库类型（可选，默认使用配置文件中的设置）
//
// 请求示例:
//
//	{
//	    "dir": "./rag_docs",
//	    "dbType": "chromem"
//	}
type RagIndexRequest struct {
	// Dir 要索引的文档目录路径
	// 支持相对路径和绝对路径
	// 目录中的所有文档将被读取并索引到向量数据库
	Dir string `json:"dir" binding:"required"`

	// DbType 向量数据库类型
	// 可选值: chromem, redis, milvus
	// 如果不指定，使用配置文件中的 rag.type
	DbType string `json:"dbType"`
}

// RagIndex 执行 RAG 索引
//
// 该函数处理知识库索引请求，将指定目录的文档索引到向量数据库。
//
// 请求方法: POST
// 请求路径: /api/v1/rag/index
// Content-Type: application/json
//
// 请求示例:
//
//	{
//	    "dir": "./rag_docs",
//	    "dbType": "chromem"
//	}
//
// 响应示例（成功）:
//
//	{
//	    "request_id": "abc123",
//	    "success": true,
//	    "data": {
//	        "message": "索引成功",
//	        "dir": "./rag_docs",
//	        "dbType": "chromem"
//	    },
//	    "msg": "操作成功",
//	    "total": 0
//	}
//
// 响应示例（失败）:
//
//	{
//	    "request_id": "abc123",
//	    "success": false,
//	    "data": {},
//	    "msg": "参数错误: ...",
//	    "total": 0
//	}
//
// 索引流程:
//  1. 验证请求参数
//  2. 读取指定目录的文档
//  3. 将文档分割成小块
//  4. 生成文档向量
//  5. 存储到向量数据库
//
// 支持的文档类型:
//   - Markdown (.md)
//   - 文本文件 (.txt)
//   - 其他纯文本格式
func RagIndex(c *gin.Context) {
	// 解析请求参数
	var req RagIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 执行索引
	var err error
	if req.DbType != "" {
		// 使用指定的向量数据库类型
		err = rag_index.RagIndexWithType(c.Request.Context(), req.Dir, req.DbType)
	} else {
		// 使用配置文件中的默认向量数据库
		err = rag_index.RagIndex(c.Request.Context())
	}

	// 处理错误
	if err != nil {
		models.FailWithMessage("索引失败: "+err.Error(), c)
		return
	}

	// 返回成功响应
	models.OkWithData(gin.H{
		"message": "索引成功",
		"dir":     req.Dir,
		"dbType":  req.DbType,
	}, c)
}
