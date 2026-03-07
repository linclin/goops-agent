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

var systemPrompt = `
# 角色: DevOps专家助手

## 核心能力
- 结合内部知识库和网络文档回答用户问题
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

## 上下文信息
- 当前日期: {date}
- 相关文档: |-
==== doc start ====
  {documents}
==== doc end ====
`

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// newChatTemplate component initialization function of node 'ChatTemplate' in graph 'EinoAgent'
func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	// TODO Modify component configuration here.
	config := &ChatTemplateConfig{
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(systemPrompt),
			schema.MessagesPlaceholder("history", true),
			schema.UserMessage("{content}"),
		},
	}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}
