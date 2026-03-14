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

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// systemPrompt 系统提示词模板
//
// 该模板定义了 AI 助手的行为规范和响应策略。
// 模板使用 FString 格式，支持以下占位符：
//   - {date}: 当前日期时间
//   - {conversation_history}: 从向量数据库检索的对话历史
//   - {documents}: 从知识库检索的相关文档
//
// 提示词结构:
//   - 角色定义: AI 智能助手
//   - 核心能力: 知识库检索、通用对话、工具调用
//   - 交互指南: 回应规则、帮助方式、限制说明
//   - 知识库使用规则: 优先级和来源说明
//   - 对话历史使用规则: 上下文连贯性
//   - 上下文信息: 动态填充的检索结果
var systemPrompt = `
# 角色: AI 智能助手

## 核心能力
- 结合内部知识库和网络文档回答用户问题
- 当知识库中没有相关内容时，基于通用知识回答问题
- 记住之前的对话内容，提供连贯的对话体验
- 网络搜索、文件/URL 打开

## 交互指南
- 回应前，请确保：
  • 完全理解用户的请求和需求，如有歧义，请向用户澄清
  • 考虑最合适的解决方案

- 提供帮助时：
  • 清晰简洁
  • 相关时包含实用示例
  • 必要时参考文档
  • 适当时建议改进或后续步骤

- 如果请求超出您的能力范围：
  • 明确传达您的限制，尽可能建议替代方法

- 如果问题复杂或复合，您需要逐步思考，避免直接给出低质量答案。

## 知识库使用规则
- 如果"相关文档"部分有内容，请优先基于知识库内容回答问题
- 如果"相关文档"部分为空或没有相关内容，请基于您的通用知识回答问题
- 回答时要明确说明信息来源（知识库或通用知识）

## 对话历史使用规则
- 如果"对话历史"部分有内容，请参考之前的对话内容
- 保持对话的连贯性和上下文理解
- 如果用户的问题与之前的对话相关，请结合历史信息回答

## 上下文信息
- 当前日期: {date}
- 对话历史: |-
==== conversation start ====
  {conversation_history}
==== conversation end ====
- 相关文档: |-
==== doc start ====
  {documents}
==== doc end ====
`

// ChatTemplateConfig 聊天模板配置
//
// 定义了提示词模板的格式和消息模板列表。
type ChatTemplateConfig struct {
	// FormatType 模板格式类型
	// FString: 使用 Python 风格的格式化字符串，如 {name}
	FormatType schema.FormatType

	// Templates 消息模板列表
	// 按顺序定义系统消息、历史消息占位符和用户消息
	Templates []schema.MessagesTemplate
}

// newChatTemplate 创建聊天模板组件
//
// 该函数是 Graph 中 ChatTemplate 节点的初始化函数，
// 负责创建和配置提示词模板。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - ctp: 聊天模板实例，用于渲染最终的提示词
//   - err: 创建过程中的错误
//
// 模板结构:
//
//  1. SystemMessage: 系统提示词，定义 AI 的角色和行为规范
//  2. MessagesPlaceholder: 历史消息占位符，用于插入对话历史
//  3. UserMessage: 用户消息模板，包含 {content} 占位符
//
// 变量来源:
//   - {date}: 由 inputToHistory 节点提供
//   - {history}: 由 inputToHistory 节点提供（前端传入的历史）
//   - {content}: 由 inputToHistory 节点提供
//   - {documents}: 由 Retriever 节点提供（知识库检索结果）
//   - {conversation_history}: 由 ConversationRetriever 节点提供（向量数据库中的历史）
func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	// 创建模板配置
	config := &ChatTemplateConfig{
		// 使用 FString 格式，支持 {variable} 风格的占位符
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			// 系统消息：定义 AI 的角色和行为规范
			schema.SystemMessage(systemPrompt),
			// 历史消息占位符：插入前端传入的对话历史
			// 第二个参数 true 表示历史消息可选（可以为空）
			schema.MessagesPlaceholder("history", true),
			// 用户消息：插入当前用户的问题
			// {content} 将被替换为用户查询
			schema.UserMessage("{content}"),
		},
	}

	// 从消息模板创建聊天模板
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}
