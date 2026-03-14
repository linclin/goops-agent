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
	"fmt"

	"github.com/cloudwego/eino/components/retriever"

	ragretriever "gin-mini-agent/internal/ai_agent/retriever"
	"gin-mini-agent/pkg/global"
)

// newRetriever 创建知识库检索器实例
//
// 检索器（Retriever）是 RAG（检索增强生成）系统的核心组件，
// 负责根据用户查询从知识库中检索相关的文档片段。
//
// 工作原理:
//  1. 将用户查询转换为向量（使用嵌入模型）
//  2. 在向量数据库中搜索相似的文档向量
//  3. 返回最相关的文档片段
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - rtr: 检索器实例，实现了 Retriever 接口
//   - err: 创建过程中的错误
//
// 支持的向量数据库:
//   - Chromem: 本地文件存储，适合开发和小规模部署
//   - Redis: 分布式存储，适合中等规模部署
//   - Milvus: 分布式向量数据库，适合大规模部署
//
// 配置文件示例:
//
//	rag:
//	  type: "chromem"  # 可选: chromem, redis, milvus
//	  chromem:
//	    persist_dir: "./data/chromem"
//	  redis:
//	    addr: "localhost:6379"
//	  milvus:
//	    addr: "localhost:19530"
//
// 使用示例:
//
//	retriever, err := newRetriever(ctx)
//	docs, err := retriever.Retrieve(ctx, "什么是机器学习？")
func newRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	// 创建嵌入模型实例
	// 嵌入模型用于将查询文本转换为向量
	// 检索时使用相同的嵌入模型确保向量空间一致
	embedder, err := newEmbedding(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建嵌入模型失败: %w", err)
	}

	// 从配置文件读取向量数据库类型
	// 默认使用 Redis
	dbType := global.Conf.RAG.Type
	if dbType == "" {
		dbType = "redis"
	}

	// 根据配置选择不同的向量数据库实现
	switch dbType {
	case "chromem":
		// Chromem: 本地文件存储
		// 优点: 无需额外依赖，部署简单
		// 缺点: 不支持分布式，扩展性有限
		return ragretriever.NewChromemRetriever(ctx, embedder)
	case "milvus":
		// Milvus: 分布式向量数据库
		// 优点: 高性能，支持分布式，适合大规模数据
		// 缺点: 部署复杂，需要额外资源
		return ragretriever.NewMilvusRetriever(ctx, embedder)
	case "redis":
		// Redis: 分布式缓存数据库
		// 优点: 部署简单，性能好，支持持久化
		// 缺点: 向量检索能力有限
		return ragretriever.NewRedisRetriever(ctx, embedder)
	default:
		// 默认使用 Redis
		return ragretriever.NewRedisRetriever(ctx, embedder)
	}
}
