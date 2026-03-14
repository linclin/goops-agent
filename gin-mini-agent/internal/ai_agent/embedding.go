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

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

// newEmbedding 创建嵌入模型实例
//
// 嵌入模型（Embedding Model）用于将文本转换为高维向量表示。
// 这些向量捕获文本的语义信息，使得语义相似的文本在向量空间中距离更近。
//
// 主要用途:
//   1. 向量检索：将查询和文档转换为向量，计算相似度
//   2. 知识库索引：将文档向量化后存储到向量数据库
//   3. 对话历史存储：将对话内容向量化后存储
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - eb: 嵌入模型实例，实现了 Embedder 接口
//   - err: 创建过程中的错误
//
// 配置说明:
//   - BaseURL: API 基础地址，从配置文件读取
//   - APIKey: API 密钥，从配置文件读取
//   - Model: 嵌入模型名称，从配置文件读取
//
// 常用嵌入模型:
//   - text-embedding-ada-002: OpenAI 的经典嵌入模型
//   - text-embedding-3-small: OpenAI 的新一代小模型
//   - text-embedding-3-large: OpenAI 的新一代大模型
//
// 向量维度:
//   - ada-002: 1536 维
//   - text-embedding-3-small: 1536 维
//   - text-embedding-3-large: 3072 维
//
// 配置文件示例:
//
//	ai_model:
//	  embedding_model:
//	    base_url: "https://api.openai.com/v1"
//	    api_key: "sk-xxx"
//	    model: "text-embedding-ada-002"
func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	// 从配置文件读取嵌入模型配置
	config := &openai.EmbeddingConfig{
		// BaseURL API 基础地址
		// 示例:
		// - OpenAI: https://api.openai.com/v1
		// - Azure: https://your-resource.openai.azure.com
		// - 国内代理: https://api.your-proxy.com/v1
		BaseURL: global.Conf.AiModel.EmbeddingModel.BaseURL,

		// APIKey API 访问密钥
		// 用于身份验证，保护 API 资源
		APIKey: global.Conf.AiModel.EmbeddingModel.APIKey,

		// Model 嵌入模型名称
		// 指定使用哪个模型生成向量
		// 常用模型:
		// - text-embedding-ada-002: 经典模型
		// - text-embedding-3-small: 新一代小模型
		// - text-embedding-3-large: 新一代大模型
		Model: global.Conf.AiModel.EmbeddingModel.Model,
	}

	// 创建嵌入模型实例
	// 使用 OpenAI 兼容的 API 接口
	eb, err = openai.NewEmbedder(ctx, config)
	if err != nil {
		return nil, err
	}
	return eb, nil
}
