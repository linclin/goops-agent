package retriever

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	redisCli "github.com/redis/go-redis/v9"

	"gin-mini-agent/pkg/global"
)

// Redis 字段常量定义
const (
	// ContentField 内容字段名
	ContentField = "content"
	// MetadataField 元数据字段名
	MetadataField = "metadata"
	// VectorField 向量字段名
	VectorField = "content_vector"
	// DistanceField 距离字段名
	DistanceField = "distance"
)

// NewRedisRetriever 创建 Redis 检索器
//
// Redis 是一个高性能的内存数据库，支持向量检索功能。
// 适合中等规模部署场景，支持分布式和持久化。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - embedder: 嵌入模型，用于将查询转换为向量
//
// 返回:
//   - retriever.Retriever: 检索器实例
//   - error: 创建过程中的错误
//
// 配置:
//   - redis.addr: Redis 服务器地址，如 "localhost:6379"
//   - redis.prefix: 键前缀，用于区分不同应用
//
// 前置条件:
//   - Redis 服务器需要安装 RedisSearch 模块
//   - 需要先创建向量索引
//
// 使用示例:
//
//	retriever, err := NewRedisRetriever(ctx, embedder)
//	docs, err := retriever.Retrieve(ctx, "什么是机器学习？")
func NewRedisRetriever(ctx context.Context, embedder embedding.Embedder) (rtr retriever.Retriever, err error) {
	// 创建 Redis 客户端
	redisClient := redisCli.NewClient(&redisCli.Options{
		Addr:     global.Conf.RAG.Redis.Addr,
		Protocol: 2, // 使用 RESP2 协议
	})

	// 创建 Redis 检索器配置
	config := &redis.RetrieverConfig{
		// Redis 客户端
		Client: redisClient,
		// 向量索引名称，使用配置的前缀
		Index:        fmt.Sprintf("%svector_index", global.Conf.RAG.Redis.Prefix),
		// RedisSearch 方言版本
		Dialect:      2,
		// 返回的字段列表
		ReturnFields: []string{ContentField, MetadataField, DistanceField},
		// 默认返回 8 个结果
		TopK:         8,
		// 向量字段名
		VectorField:  VectorField,
		// 文档转换函数：将 Redis 文档转换为 schema.Document
		DocumentConverter: func(ctx context.Context, doc redisCli.Document) (*schema.Document, error) {
			resp := &schema.Document{
				ID:       doc.ID,
				Content:  "",
				MetaData: map[string]any{},
			}
			// 遍历文档字段
			for field, val := range doc.Fields {
				if field == ContentField {
					// 内容字段
					resp.Content = val
				} else if field == MetadataField {
					// 元数据字段，解析 JSON
					var metadata map[string]any
					if err := json.Unmarshal([]byte(val), &metadata); err != nil {
						// 解析失败，直接存储原始值
						resp.MetaData[field] = val
					} else {
						// 解析成功，复制元数据
						for k, v := range metadata {
							resp.MetaData[k] = v
						}
					}
				} else if field == "vector_distance" {
					// 向量距离字段，转换为相似度分数
					distance, err := strconv.ParseFloat(val, 64)
					if err != nil {
						continue
					}
					// score = 1 - distance，值越大表示越相似
					resp.WithScore(1 - distance)
				}
			}

			return resp, nil
		},
	}

	// 设置嵌入模型
	config.Embedding = embedder

	// 创建 Redis 检索器
	rtr, err = redis.NewRetriever(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建 redis 检索器失败: %w", err)
	}
	return rtr, nil
}
