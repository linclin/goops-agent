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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"

	"gin-mini-agent/internal/ai_agent"
	"gin-mini-agent/models"
)

var CallbacksHandler callbacks.Handler

// ChatRequest 聊天请求参数
//
// 该结构体定义了聊天接口的请求格式。
//
// 字段说明:
//   - ID: 会话唯一标识符，用于追踪对话会话
//   - Query: 用户输入的问题或请求
//   - History: 对话历史，包含之前的对话记录
type ChatRequest struct {
	// ID 会话唯一标识符
	// 用于追踪对话会话，便于日志记录和问题排查
	ID string `json:"id" binding:"required"`

	// Query 用户输入的问题或请求
	// 这是用户的核心输入，会被 AI Agent 处理
	Query string `json:"query" binding:"required"`

	// History 对话历史
	// 包含当前会话中之前的对话记录
	// 用于保持对话的连贯性和上下文理解
	History []*schema.Message `json:"history"`
}

// SSEEvent SSE 事件结构
//
// 该结构体定义了 Server-Sent Events (SSE) 的事件格式。
// 用于流式传输 AI 生成的内容。
//
// 事件类型:
//   - start: 对话开始
//   - message: 消息内容（逐字发送）
//   - done: 对话完成
//   - error: 发生错误
type SSEEvent struct {
	// Event 事件类型
	// 可选值: start, message, done, error
	Event string `json:"event"`

	// Data 事件数据
	// 对于 message 事件，包含单个字符
	// 对于 error 事件，包含错误信息
	Data string `json:"data"`
}

// sseEvent 生成 SSE 格式的事件字符串
//
// 该函数将事件名和数据格式化为标准的 SSE (Server-Sent Events) 格式。
// SSE 标准格式: event: <event>\ndata: <data>\n\n
//
// 参数:
//   - name: 事件名称
//   - data: 事件数据
//
// 返回:
//   - string: SSE 格式的字符串
//
// 注意:
//   - Gin 的 c.SSEvent 方法是直接写入 Context 响应的，不返回值
//   - 在 c.Stream 回调中需要写入 io.Writer，因此需要手动格式化
func sseEvent(name string, data interface{}) string {
	return fmt.Sprintf("event: %s\ndata: %v\n\n", name, data)
}

// Chat 执行 AI Agent 对话（支持 SSE 流式输出）
//
// 该函数处理流式对话请求，使用 Server-Sent Events (SSE) 技术
// 实时返回 AI 生成的内容。
//
// 请求方法: POST
// 请求路径: /api/v1/agent/chat
// Content-Type: application/json
//
// 请求示例:
//
//	{
//	    "id": "session_123",
//	    "query": "什么是机器学习？",
//	    "history": [
//	        {"role": "user", "content": "你好"},
//	        {"role": "assistant", "content": "你好！有什么可以帮助你的？"}
//	    ]
//	}
//
// 响应格式: text/event-stream
// 事件序列:
//  1. start: 对话开始
//  2. message: 逐字发送内容（多个事件）
//  3. done: 对话完成
//
// 对话历史存储:
//   - 对话完成后，异步存储到向量数据库
//   - 存储失败不影响用户体验
func ChatSse(c *gin.Context) {
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
	cbLogConfig := &LogCallbackConfig{
		Detail: true,
		Debug:  true,
	}
	// this is for invoke option of WithCallback
	CallbacksHandler = LogCallback(cbLogConfig)
	// 调用 AI Agent 的 Stream 方法获取流式响应
	streamReader, err := runnable.Stream(c.Request.Context(), userMessage, compose.WithCallbacks(CallbacksHandler))
	if err != nil {
		models.FailWithMessage("调用 AI Agent 失败: "+err.Error(), c)
		return
	}
	defer streamReader.Close()

	// 用于收集 AI 回复内容（用于存储对话历史）
	var aiResponse string

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 开始流式传输
	// c.Stream 的回调函数返回 bool：
	//   - true: 继续流式传输
	//   - false: 结束流式传输
	c.Stream(func(w io.Writer) bool {
		// 发送开始事件
		io.WriteString(w, sseEvent("start", ""))

		// 循环接收流式数据
		for {
			// 从 StreamReader 接收数据
			resp, err := streamReader.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					// 流结束，发送完成事件
					io.WriteString(w, sseEvent("done", ""))

					// 异步存储对话历史到向量数据库
					// 在后台执行，不阻塞响应
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
					io.WriteString(w, sseEvent("error", string(data)))
					break
				}
			}

			// 处理接收到的消息
			if resp != nil && resp.Content != "" {
				// 收集 AI 回复内容（用于存储对话历史）
				aiResponse += resp.Content

				// 逐字发送消息内容
				// 每个字符作为单独的 SSE 事件发送
				for _, char := range resp.Content {
					data, _ := json.Marshal(SSEEvent{
						Event: "message",
						Data:  string(char),
					})
					io.WriteString(w, sseEvent("message", string(data)))
				}
			}
		}

		return false // 结束流
	})
}

type LogCallbackConfig struct {
	Detail bool
	Debug  bool
}

func LogCallback(config *LogCallbackConfig) callbacks.Handler {
	if config == nil {
		config = &LogCallbackConfig{
			Detail: true,
		}
	}
	builder := callbacks.NewHandlerBuilder()
	builder.OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
		slog.InfoContext(ctx, "[callback] 节点开始执行",
			"component", info.Component,
			"type", info.Type,
			"name", info.Name)

		// 监控工具调用，特别是 skill 工具
		if info.Component == "tool" {
			slog.InfoContext(ctx, "[callback] 工具调用开始",
				"tool_name", info.Name,
				"tool_type", info.Type)

			// 检查是否是 skill 工具
			if info.Name == "skill" {
				slog.InfoContext(ctx, "[callback] 🎯 Skill 工具被调用！")
			}
		}

		if config.Detail {
			var b []byte
			if config.Debug {
				b, _ = json.MarshalIndent(input, "", "  ")
			} else {
				b, _ = json.Marshal(input)
			}
			slog.DebugContext(ctx, "[callback] 输入数据", "input", string(b))
		}
		return ctx
	})
	builder.OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		slog.InfoContext(ctx, "[callback] 节点执行完成",
			"component", info.Component,
			"type", info.Type,
			"name", info.Name)

		// 监控工具调用完成
		if info.Component == "tool" && info.Name == "skill" {
			slog.InfoContext(ctx, "[callback] 🎯 Skill 工具调用完成！")
		}

		if config.Detail {
			var b []byte
			if config.Debug {
				b, _ = json.MarshalIndent(output, "", "  ")
			} else {
				b, _ = json.Marshal(output)
			}
			slog.DebugContext(ctx, "[callback] 输出数据", "output", string(b))
		}
		return ctx
	})
	builder.OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
		slog.ErrorContext(ctx, "[callback] 节点执行错误",
			"component", info.Component,
			"type", info.Type,
			"name", info.Name,
			"error", err)
		return ctx
	})
	return builder.Build()
}
