/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package agent

import (
	"encoding/json"
	"io"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"

	"gin-mini-agent/internal/ai_agent"
	"gin-mini-agent/models"
)

// ChatRequest 聊天请求参数
type ChatRequest struct {
	ID      string            `json:"id" binding:"required"`
	Query   string            `json:"query" binding:"required"`
	History []*schema.Message `json:"history"`
}

// SSEEvent SSE 事件结构
type SSEEvent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// Chat 执行 AI Agent 对话（支持 SSE 流式输出）
func Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 使用全局 AI Agent
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

	// 调用 AI Agent 的 Stream 方法获取流式响应
	streamReader, err := runnable.Stream(c.Request.Context(), userMessage)
	if err != nil {
		models.FailWithMessage("调用 AI Agent 失败: "+err.Error(), c)
		return
	}
	defer streamReader.Close()

	// 用于收集 AI 回复内容
	var aiResponse string

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Stream(func(w io.Writer) bool {
		// 发送开始事件
		sse.Encode(w, sse.Event{
			Event: "start",
			Data:  "",
		})

		// 循环接收流式数据
		for {
			// 从 StreamReader 接收数据
			resp, err := streamReader.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					// 流结束，发送完成事件
					sse.Encode(w, sse.Event{
						Event: "done",
						Data:  "",
					})

					// 异步存储对话历史到向量数据库
					if aiResponse != "" && ai_agent.GlobalConversationManager != nil {
						go func(userQuery, response string) {
							// 使用后台上下文存储，不阻塞响应
							storeCtx := c.Request.Context()
							if storeErr := ai_agent.GlobalConversationManager.Store(storeCtx, userQuery, response); storeErr != nil {
								// 存储失败只记录日志，不影响用户体验
								// log.Printf("存储对话历史失败: %v", storeErr)
							}
						}(req.Query, aiResponse)
					}
					break
				} else {
					// 发生错误，发送错误事件
					data, _ := json.Marshal(SSEEvent{
						Event: "error",
						Data:  err.Error(),
					})
					sse.Encode(w, sse.Event{
						Event: "error",
						Data:  string(data),
					})
					break
				}
			}

			// 处理接收到的消息
			if resp != nil && resp.Content != "" {
				// 收集 AI 回复内容
				aiResponse += resp.Content

				// 逐字发送消息内容
				for _, char := range resp.Content {
					data, _ := json.Marshal(SSEEvent{
						Event: "message",
						Data:  string(char),
					})
					sse.Encode(w, sse.Event{
						Event: "message",
						Data:  string(data),
					})
				}
			}
		}

		return false // 结束流
	})
}

// ChatNonStream 非流式聊天接口（用于兼容）
func ChatNonStream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 使用全局 AI Agent
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

	// 调用 AI Agent
	resp, err := runnable.Invoke(c.Request.Context(), userMessage)
	if err != nil {
		models.FailWithMessage("调用 AI Agent 失败: "+err.Error(), c)
		return
	}

	// 存储对话历史到向量数据库
	if resp.Content != "" && ai_agent.GlobalConversationManager != nil {
		go func(userQuery, response string) {
			if storeErr := ai_agent.GlobalConversationManager.Store(c.Request.Context(), userQuery, response); storeErr != nil {
				// 存储失败只记录日志，不影响用户体验
				// log.Printf("存储对话历史失败: %v", storeErr)
			}
		}(req.Query, resp.Content)
	}

	models.OkWithData(gin.H{
		"id":      req.ID,
		"content": resp.Content,
		"role":    resp.Role,
	}, c)
}
