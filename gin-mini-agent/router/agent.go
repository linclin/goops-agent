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

// Package router 提供 HTTP 路由注册功能
//
// 该包负责注册应用程序的 HTTP 路由。
// 主要功能包括：
//   - Agent 路由注册
//   - RAG 路由注册
//
// 路由结构:
//   - /api/v1/agent/*: Agent 相关路由
//   - /api/v1/rag/*: RAG 相关路由
package router

import (
	"github.com/gin-gonic/gin"

	"gin-mini-agent/api/v1/agent"
	"gin-mini-agent/pkg/global"
)

// InitAgentRouter 初始化 Agent 路由
//
// 该函数注册 Agent 相关的 HTTP 路由。
// 所有路由都需要 Basic Auth 认证。
//
// 参数:
//   - r: Gin 路由组（通常是 /api/v1）
//
// 返回:
//   - gin.IRoutes: 注册的路由
//
// 路由列表:
//   - POST /agent/chat: SSE 流式聊天接口
//   - POST /agent/chat/non-stream: 非流式聊天接口（兼容）
//
// 认证配置:
//   - 使用 Basic Auth
//   - 用户名和密码从配置文件读取
//
// 使用示例:
//
//	v1Group := r.Group("v1")
//	router.InitAgentRouter(v1Group)
func InitAgentRouter(r *gin.RouterGroup) (R gin.IRoutes) {
	// 创建 Agent 路由组
	// 添加 Basic Auth 认证中间件
	router := r.Group("agent", gin.BasicAuth(gin.Accounts{
		global.Conf.Auth.User: global.Conf.Auth.Password,
	}))
	{
		// SSE 流式聊天接口
		// 支持 Server-Sent Events 流式输出
		router.POST("/chat/sse", agent.ChatSse)

		// 非流式聊天接口（兼容）
		// 返回完整的 JSON 响应
		router.POST("/chat/sync", agent.ChatSync)
	}
	return router
}
