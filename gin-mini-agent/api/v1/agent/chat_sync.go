// Package agent 提供 AI Agent 的 HTTP API 接口
//
// 该包实现了 AI Agent 的 HTTP 接口，支持流式和非流式对话。
// 主要功能包括：
//   - 流式对话（SSE）：实时返回 AI 生成的内容
//   - 非流式对话：一次性返回完整响应
//   - 对话历史存储：自动将对话存储到向量数据库
//
// 接口说明:
//   - POST /api/v1/agent/chat: 流式对话接口
//   - POST /api/v1/agent/chat/sync: 非流式对话接口
package agent

import (
	"github.com/gin-gonic/gin"

	"gin-mini-agent/internal/ai_agent"
	"gin-mini-agent/models"
)

// ChatSync 非流式聊天接口
//
// 该函数处理非流式对话请求，一次性返回完整的 AI 响应。
// 适用于不需要实时显示生成过程的场景。
//
// 请求方法: POST
// 请求路径: /api/v1/agent/chat/sync
// Content-Type: application/json
//
// 请求示例:
//
//	{
//	    "id": "session_123",
//	    "query": "什么是机器学习？",
//	    "history": []
//	}
//
// 响应格式: application/json
// 响应示例:
//
//	{
//	    "code": 200,
//	    "data": {
//	        "id": "session_123",
//	        "content": "机器学习是人工智能的一个分支...",
//	        "role": "assistant"
//	    }
//	}
//
// 对话历史存储:
//   - 响应返回后，异步存储到向量数据库
//   - 存储失败不影响用户体验
func ChatSync(c *gin.Context) {
	// 解析请求参数
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 获取全局 AI Agent 实例
	runnable := ai_agent.GlobalAgent
	if runnable == nil {
		models.FailWithMessage("AI Agent 未初始化", c)
		return
	}

	// 准备输入消息
	userMessage := &ai_agent.UserMessage{
		ID:      req.ID,
		Query:   req.Query,
		History: req.History,
	}

	// 调用 AI Agent 的 Invoke 方法获取完整响应
	resp, err := runnable.Invoke(c.Request.Context(), userMessage)
	if err != nil {
		models.FailWithMessage("调用 AI Agent 失败: "+err.Error(), c)
		return
	}

	// 异步存储对话历史到向量数据库
	if resp.Content != "" && ai_agent.GlobalConversationManager != nil {
		go func(userQuery, response string) {
			if storeErr := ai_agent.GlobalConversationManager.Store(c.Request.Context(), userQuery, response); storeErr != nil {
				// 存储失败只记录日志，不影响用户体验
				// log.Printf("存储对话历史失败: %v", storeErr)
			}
		}(req.Query, resp.Content)
	}

	// 返回响应
	models.OkWithData(gin.H{
		"id":      req.ID,
		"content": resp.Content,
		"role":    resp.Role,
	}, c)
}
