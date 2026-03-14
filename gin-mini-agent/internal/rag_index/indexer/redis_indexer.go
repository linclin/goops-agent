package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
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
)

// InitRedisIndex 初始化 Redis 向量索引
//
// 该函数检查并创建 Redis 向量索引。
// 如果索引已存在，则跳过创建。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - client: Redis 客户端
//
// 返回:
//   - error: 初始化过程中的错误
//
// 索引配置:
//   - 索引类型: FT.CREATE（RedisSearch）
//   - 存储类型: HASH
//   - 向量算法: FLAT（暴力搜索）
//   - 向量维度: 4096
//   - 距离度量: COSINE（余弦相似度）
//
// 前置条件:
//   - Redis 服务器需要安装 RedisSearch 模块
//   - Redis 版本需要支持向量搜索（Redis Stack 7.2+）
func InitRedisIndex(ctx context.Context, client *redisCli.Client) (err error) {
	// 测试 Redis 连接
	if err = client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// 构建索引名称
	indexName := fmt.Sprintf("%s%s", global.Conf.RAG.Redis.Prefix, "vector_index")

	// 检查索引是否已存在
	exists, err := client.Do(ctx, "FT.INFO", indexName).Result()
	if err != nil {
		// 如果错误不是"Unknown index name"，则返回错误
		if !strings.Contains(err.Error(), "Unknown index name") {
			return fmt.Errorf("failed to check if index exists: %w", err)
		}
		err = nil
	} else if exists != nil {
		// 索引已存在，直接返回
		return nil
	}

	// 创建向量索引
	// 参数说明:
	// - FT.CREATE: 创建索引命令
	// - ON HASH: 使用 HASH 存储文档
	// - PREFIX: 键前缀，用于区分不同应用
	// - SCHEMA: 定义字段结构
	//   - content: 文本字段，用于全文搜索
	//   - metadata: 文本字段，存储元数据 JSON
	//   - content_vector: 向量字段，存储文档向量
	//     - FLAT: 使用暴力搜索算法
	//     - TYPE FLOAT32: 向量元素类型
	//     - DIM 4096: 向量维度
	//     - DISTANCE_METRIC COSINE: 使用余弦相似度
	createIndexArgs := []interface{}{
		"FT.CREATE", indexName,
		"ON", "HASH",
		"PREFIX", "1", global.Conf.RAG.Redis.Prefix,
		"SCHEMA",
		ContentField, "TEXT",
		MetadataField, "TEXT",
		VectorField, "VECTOR", "FLAT",
		"6",
		"TYPE", "FLOAT32",
		"DIM", 4096,
		"DISTANCE_METRIC", "COSINE",
	}

	// 执行创建索引命令
	if err = client.Do(ctx, createIndexArgs...).Err(); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// 验证索引是否创建成功
	if _, err = client.Do(ctx, "FT.INFO", indexName).Result(); err != nil {
		return fmt.Errorf("failed to verify index creation: %w", err)
	}

	return nil
}

// NewRedisIndexer 创建 Redis 索引器
//
// Redis 是一个高性能的内存数据库，支持向量检索功能。
// 适合中等规模部署场景，支持分布式和持久化。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - embedder: 嵌入模型，用于将文档内容转换为向量
//
// 返回:
//   - indexer.Indexer: 索引器实例
//   - error: 创建过程中的错误
//
// 配置:
//   - redis.addr: Redis 服务器地址，如 "localhost:6379"
//   - redis.prefix: 键前缀，用于区分不同应用
//
// 前置条件:
//   - Redis 服务器需要安装 RedisSearch 模块
//   - Redis 版本需要支持向量搜索（Redis Stack 7.2+）
//
// 使用示例:
//
//	indexer, err := NewRedisIndexer(ctx, embedder)
//	ids, err := indexer.Store(ctx, []*schema.Document{...})
func NewRedisIndexer(ctx context.Context, embedder embedding.Embedder) (idr indexer.Indexer, err error) {
	// 创建 Redis 客户端
	redisClient := redisCli.NewClient(&redisCli.Options{
		Addr:     global.Conf.RAG.Redis.Addr,
		Protocol: 2, // 使用 RESP2 协议
	})

	// 错误时关闭客户端
	defer func() {
		if err != nil {
			redisClient.Close()
		}
	}()

	// 初始化向量索引
	if err = InitRedisIndex(ctx, redisClient); err != nil {
		return nil, err
	}

	// 创建 Redis 索引器配置
	config := &redis.IndexerConfig{
		// Redis 客户端
		Client: redisClient,
		// 键前缀
		KeyPrefix: global.Conf.RAG.Redis.Prefix,
		// 批处理大小
		// 设置为 1 表示逐个处理文档
		BatchSize: 1,
		// 文档转换函数：将 schema.Document 转换为 Redis Hash
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redis.Hashes, error) {
			// 如果文档没有 ID，生成 UUID
			if doc.ID == "" {
				doc.ID = uuid.New().String()
			}
			key := doc.ID

			// 序列化元数据为 JSON
			metadataBytes, err := json.Marshal(doc.MetaData)
			if err != nil {
				return nil, fmt.Errorf("序列化元数据失败: %w", err)
			}

			// 返回 Redis Hash 结构
			// 字段说明:
			// - content: 文档内容，EmbedKey 指定向量字段
			// - metadata: 元数据 JSON 字符串
			return &redis.Hashes{
				Key: key,
				Field2Value: map[string]redis.FieldValue{
					ContentField:  {Value: doc.Content, EmbedKey: VectorField},
					MetadataField: {Value: metadataBytes},
				},
			}, nil
		},
	}

	// 设置嵌入模型
	config.Embedding = embedder

	// 创建 Redis 索引器
	idr, err = redis.NewIndexer(ctx, config)
	if err != nil {
		return nil, err
	}

	// 返回包装后的索引器实例
	return &redisIndexer{
		indexer: idr,
	}, nil
}

// redisIndexer Redis 索引器实现
//
// 实现了 indexer.Indexer 接口，提供基于 Redis 的向量存储能力。
// 包装了官方的 Redis 索引器。
type redisIndexer struct {
	// indexer 官方 Redis 索引器实例
	indexer indexer.Indexer
}

// Store 将文档存储到 Redis
//
// 该方法实现了 indexer.Indexer 接口，将文档向量化并存储。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - docs: 要存储的文档列表
//   - opts: 可选配置
//
// 返回:
//   - []string: 文档 ID 列表
//   - error: 存储过程中的错误
//
// 处理流程:
//   - 调用官方索引器的 Store 方法存储文档
//   - 文档 ID 生成和向量化由 DocumentToHashes 函数处理
func (r *redisIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	return r.indexer.Store(ctx, docs, opts...)
}
