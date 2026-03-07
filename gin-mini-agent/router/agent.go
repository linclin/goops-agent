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

package router

import (
	"gin-mini-agent/api/v1/agent"
	"gin-mini-agent/pkg/global"

	"github.com/gin-gonic/gin"
)

func InitAgentRouter(r *gin.RouterGroup) (R gin.IRoutes) {
	router := r.Group("agent", gin.BasicAuth(gin.Accounts{
		global.Conf.Auth.User: global.Conf.Auth.Password,
	}))
	{
		// SSE 流式聊天接口
		router.POST("/chat", agent.Chat)
		// 非流式聊天接口（兼容）
		router.POST("/chat/non-stream", agent.ChatNonStream)
	}
	return router
}
