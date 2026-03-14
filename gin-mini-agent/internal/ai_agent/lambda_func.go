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

import (
	"context"
	"time"
)

// inputToQuery 将用户消息转换为查询字符串
//
// 该函数是 Graph 中 InputToQuery 节点的处理函数，
// 负责从 UserMessage 结构体中提取 Query 字段。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - input: 用户消息结构体，包含 ID、Query 和 History
//   - opts: 可选参数，当前未使用
//
// 返回:
//   - output: 查询字符串，用于后续的知识库检索和对话历史检索
//   - err: 错误信息，当前始终为 nil
//
// 数据流:
//
//	UserMessage{Query: "什么是机器学习？"} -> "什么是机器学习？"
//
// 该查询字符串将被传递给:
//   - Retriever 节点：用于知识库检索
//   - ConversationRetriever 节点：用于对话历史检索
func inputToQuery(ctx context.Context, input *UserMessage, opts ...any) (output string, err error) {
	return input.Query, nil
}

// inputToHistory 将用户消息转换为模板变量
//
// 该函数是 Graph 中 InputToHistory 节点的处理函数，
// 负责将 UserMessage 转换为 ChatTemplate 所需的变量格式。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - input: 用户消息结构体，包含 ID、Query 和 History
//   - opts: 可选参数，当前未使用
//
// 返回:
//   - output: 模板变量映射，包含以下键:
//   - content: 用户查询字符串
//   - history: 前端传入的对话历史
//   - date: 当前时间，格式为 "2006-01-02 15:04:05"
//   - err: 错误信息，当前始终为 nil
//
// 数据流:
//
//	UserMessage -> map[string]any{
//	    "content": "什么是机器学习？",
//	    "history": []*schema.Message{...},
//	    "date": "2024-01-15 10:30:00",
//	}
//
// 这些变量将被传递给 ChatTemplate 节点，用于渲染提示词模板。
// 模板中的占位符 {content}、{history}、{date} 将被替换为对应的值。
func inputToHistory(ctx context.Context, input *UserMessage, opts ...any) (output map[string]any, err error) {
	return map[string]any{
		// content 用户查询内容
		// 对应模板中的 {content} 占位符
		"content": input.Query,
		// history 对话历史
		// 对应模板中的 {history} 占位符
		// 用于 MessagesPlaceholder 渲染历史消息
		"history": input.History,
		// date 当前时间
		// 对应模板中的 {date} 占位符
		// 用于让 AI 了解当前时间，便于回答时间相关问题
		"date": time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}
