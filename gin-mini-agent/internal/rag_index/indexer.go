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

	einoindexer "github.com/cloudwego/eino/components/indexer"

	ragindexer "gin-mini-agent/internal/rag_index/indexer"
)

// newIndexer 创建索引器
//
// 索引器负责将文档向量化并存储到向量数据库。
// 根据配置选择不同的向量数据库实现。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - dbType: 向量数据库类型（chromem/redis/milvus）
//
// 返回:
//   - einoindexer.Indexer: 索引器实例
//   - error: 创建过程中的错误
//
// 支持的向量数据库:
//   - chromem: 本地文件存储，适合开发和小规模部署
//   - redis: 分布式存储，适合中等规模部署
//   - milvus: 分布式向量数据库，适合大规模部署
//
// 重试机制:
//   - 使用 retryEmbedder 包装嵌入模型
//   - 最大重试次数: 5 次
//   - 重试间隔: 2 秒
//   - 用于处理 API 限流等临时错误
func newIndexer(ctx context.Context, dbType string) (idr einoindexer.Indexer, err error) {
	// 创建嵌入模型
	// 嵌入模型用于将文档内容转换为向量
	embedder, err := newEmbedding(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建嵌入模型失败: %w", err)
	}

	// 创建带重试的嵌入模型
	// 参数:
	//   - embedder: 原始嵌入模型
	//   - maxRetries: 最大重试次数（5 次）
	//   - retryDelay: 重试间隔（2 秒）
	// 用于处理 API 限流、网络抖动等临时错误
	retryEmbedder := newRetryEmbedding(embedder, 5, 2*time.Second)

	// 根据配置选择不同的向量数据库实现
	switch dbType {
	case "chromem":
		// Chromem: 本地文件存储
		// 优点: 无需额外依赖，部署简单
		// 缺点: 不支持分布式，扩展性有限
		return ragindexer.NewChromemIndexer(ctx, retryEmbedder)
	case "milvus":
		// Milvus: 分布式向量数据库
		// 优点: 高性能，支持分布式，适合大规模数据
		// 缺点: 部署复杂，需要额外资源
		return ragindexer.NewMilvusIndexer(ctx, retryEmbedder)
	case "redis":
		// Redis: 分布式缓存数据库
		// 优点: 部署简单，性能好，支持持久化
		// 缺点: 向量检索能力有限
		return ragindexer.NewRedisIndexer(ctx, retryEmbedder)
	default:
		// 默认使用 Redis
		return ragindexer.NewRedisIndexer(ctx, retryEmbedder)
	}
}
