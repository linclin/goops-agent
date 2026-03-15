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
	"gin-mini-agent/pkg/global"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// newChatModel 创建聊天模型实例
//
// 该函数创建一个支持工具调用的聊天模型，用于：
//  1. 生成对话响应
//  2. 决定是否调用工具
//  3. 生成工具调用参数
//  4. 分析工具返回结果
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - middleware: 可选的中间件，用于增强 ChatModel 功能（如 Skill 中间件）
//
// 返回:
//   - cm: 聊天模型实例，实现了 ToolCallingChatModel 接口
//   - err: 创建过程中的错误
//
// 配置说明:
//   - BaseURL: API 基础地址，从配置文件读取
//   - APIKey: API 密钥，从配置文件读取
//   - Model: 模型名称，从配置文件读取
//
// 支持的模型:
//   - OpenAI: gpt-4, gpt-3.5-turbo 等
//   - 兼容 OpenAI API 的模型: Claude、Llama 等
//
// 配置文件示例:
//
//	ai_model:
//	  chat_model:
//	    base_url: "https://api.openai.com/v1"
//	    api_key: "sk-xxx"
//	    model: "gpt-4"
func newChatModel(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	// 从配置文件读取聊天模型配置
	config := &openai.ChatModelConfig{
		// BaseURL API 基础地址
		BaseURL: global.Conf.AiModel.ChatModel.BaseURL,

		// APIKey API 访问密钥
		// 用于身份验证，保护 API 资源
		APIKey: global.Conf.AiModel.ChatModel.APIKey,

		// Model 模型名称
		// 指定使用哪个模型进行推理
		Model: global.Conf.AiModel.ChatModel.Model,
	}

	// 创建聊天模型实例
	// 使用 OpenAI 兼容的 API 接口
	cm, err = openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}

	return cm, nil
}
