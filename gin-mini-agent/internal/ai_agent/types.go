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

package ai_agent

import "github.com/cloudwego/eino/schema"

// UserMessage 用户消息结构体
//
// 该结构体定义了 AI Agent 的输入格式，包含用户查询和相关上下文信息。
// 它是 Graph 的入口数据类型，会被各个节点处理和转换。
//
// 字段说明:
//   - ID: 会话唯一标识符，用于追踪和关联对话
//   - Query: 用户输入的问题或请求，是检索和推理的核心输入
//   - History: 前端传入的对话历史，用于保持对话连贯性
//
// 使用示例:
//
//	userMessage := &UserMessage{
//	    ID:      "session_123",
//	    Query:   "什么是机器学习？",
//	    History: []*schema.Message{
//	        {Role: "user", Content: "你好"},
//	        {Role: "assistant", Content: "你好！有什么可以帮助你的？"},
//	    },
//	}
//
// 数据流向:
//
//	UserMessage -> InputToQuery (提取 Query) -> Retriever/ConversationRetriever
//	UserMessage -> InputToHistory (转换为模板变量) -> ChatTemplate
type UserMessage struct {
	// ID 会话唯一标识符
	// 用于追踪对话会话，便于日志记录和问题排查
	ID string `json:"id"`

	// Query 用户输入的问题或请求
	// 这是用户的核心输入，会被用于：
	// 1. 知识库检索：从向量数据库中查找相关文档
	// 2. 对话历史检索：从历史对话中查找相关上下文
	// 3. 提示词模板：作为用户消息传入模板
	Query string `json:"query"`

	// History 前端传入的对话历史
	// 包含当前会话中之前的对话记录
	// 用于保持对话的连贯性和上下文理解
	// 注意：这是前端维护的历史，与向量数据库中存储的历史不同
	History []*schema.Message `json:"history"`
}
