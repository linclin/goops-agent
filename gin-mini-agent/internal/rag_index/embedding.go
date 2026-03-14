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

package rag_index

import (
	"context"
	"fmt"
	"time"

	"gin-mini-agent/pkg/global"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/samber/lo"
)

// newEmbedding 创建嵌入模型实例
//
// 嵌入模型（Embedding Model）用于将文本转换为高维向量表示。
// 这些向量捕获文本的语义信息，使得语义相似的文本在向量空间中距离更近。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - embedding.Embedder: 嵌入模型实例
//   - error: 创建过程中的错误
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

// retryEmbedding 带重试机制的嵌入模型
//
// 该结构体包装了原始嵌入模型，添加了自动重试功能。
// 用于处理 API 限流、网络抖动等临时错误。
//
// 重试策略:
//   - 最大重试次数: 可配置
//   - 重试间隔: 固定间隔
//   - 重试条件: 所有错误都会触发重试
type retryEmbedding struct {
	// embedder 原始嵌入模型
	embedder embedding.Embedder

	// maxRetry 最大重试次数
	maxRetry int

	// interval 重试间隔
	interval time.Duration
}

// newRetryEmbedding 创建带重试机制的嵌入模型
//
// 该函数包装原始嵌入模型，添加自动重试功能。
//
// 参数:
//   - embedder: 原始嵌入模型
//   - maxRetry: 最大重试次数
//   - interval: 重试间隔
//
// 返回:
//   - embedding.Embedder: 带重试功能的嵌入模型
//
// 使用场景:
//   - API 限流（Rate Limiting）
//   - 网络抖动
//   - 临时服务不可用
//
// 使用示例:
//
//	embedder, _ := newEmbedding(ctx)
//	retryEmbedder := newRetryEmbedding(embedder, 5, 2*time.Second)
func newRetryEmbedding(embedder embedding.Embedder, maxRetry int, interval time.Duration) embedding.Embedder {
	return &retryEmbedding{
		embedder: embedder,
		maxRetry: maxRetry,
		interval: interval,
	}
}

// EmbedStrings 将文本列表转换为向量（带重试）
//
// 该方法实现了 embedding.Embedder 接口。
// 在调用原始嵌入模型失败时，会自动重试。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - texts: 要转换的文本列表
//   - opts: 可选配置
//
// 返回:
//   - [][]float64: 向量列表，每个向量对应一个输入文本
//   - error: 转换错误（重试次数用尽后返回）
//
// 重试行为:
//   - 每次重试前会等待指定的间隔时间
//   - 重试时会打印日志，方便调试
//   - 所有重试失败后，返回最后一次的错误
func (r *retryEmbedding) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	var result [][]float64
	var lastErr error

	// 使用 lo 库的 AttemptWithDelay 进行带延迟的重试
	// 参数:
	//   - r.maxRetry: 最大重试次数
	//   - r.interval: 重试间隔
	//   - 回调函数: 执行实际的嵌入操作
	_, _, err := lo.AttemptWithDelay(r.maxRetry, r.interval, func(index int, duration time.Duration) error {
		// 第一次尝试（index=0）不打印日志
		if index > 0 {
			fmt.Printf("[重试] 第 %d 次重试，等待 %v 后重试...\n", index, duration)
		}

		// 调用原始嵌入模型
		res, err := r.embedder.EmbedStrings(ctx, texts, opts...)
		if err != nil {
			lastErr = err
			return err
		}

		// 成功，保存结果
		result = res
		return nil
	})

	// 所有重试都失败
	if err != nil {
		return nil, fmt.Errorf("重试 %d 次后仍然失败: %w", r.maxRetry, lastErr)
	}

	return result, nil
}
