package indexer

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/milvus-io/milvus/client/v2/milvusclient"

	milvus2 "github.com/cloudwego/eino-ext/components/indexer/milvus2"

	"gin-mini-agent/pkg/global"
)

// NewMilvusIndexer 创建 Milvus 索引器
//
// Milvus 是一个高性能的分布式向量数据库，专为大规模向量检索设计。
// 适合大规模生产环境部署，支持分布式架构和高可用。
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
//   - milvus.addr: Milvus 服务器地址，如 "localhost:19530"
//   - milvus.username: 用户名（可选）
//   - milvus.password: 密码（可选）
//   - milvus.collection: 集合名称
//
// 向量配置:
//   - Dimension: 向量维度，默认 1024
//   - MetricType: 距离度量类型，使用余弦相似度（COSINE）
//   - IndexBuilder: 索引构建器，使用 HNSW 算法
//     - M: HNSW 参数，每个节点的最大连接数，默认 16
//     - EfConstruction: 构建时的搜索范围，默认 200
//
// 使用示例:
//
//	indexer, err := NewMilvusIndexer(ctx, embedder)
//	ids, err := indexer.Store(ctx, []*schema.Document{...})
func NewMilvusIndexer(ctx context.Context, embedder embedding.Embedder) (idr indexer.Indexer, err error) {
	// 创建 Milvus 索引器配置
	config := &milvus2.IndexerConfig{
		// Milvus 客户端配置
		ClientConfig: &milvusclient.ClientConfig{
			Address:  global.Conf.RAG.Milvus.Addr,
			Username: global.Conf.RAG.Milvus.Username,
			Password: global.Conf.RAG.Milvus.Password,
		},
		// 集合名称
		Collection: global.Conf.RAG.Milvus.Collection,
		// 向量配置
		Vector: &milvus2.VectorConfig{
			// 向量维度
			// 需要与嵌入模型输出的维度一致
			// 常见维度:
			// - text-embedding-ada-002: 1536
			// - text-embedding-3-small: 1536
			// - text-embedding-3-large: 3072
			Dimension: 1024,
			// 距离度量类型
			// COSINE: 余弦相似度，适合文本语义相似度计算
			// L2: 欧几里得距离
			// IP: 内积
			MetricType: milvus2.COSINE,
			// 索引构建器
			// HNSW (Hierarchical Navigable Small World) 是一种高效的近似最近邻搜索算法
			// 参数说明:
			// - M: 每个节点的最大连接数，影响召回率和索引大小
			// - EfConstruction: 构建时的搜索范围，影响索引质量和构建时间
			IndexBuilder: milvus2.NewHNSWIndexBuilder().WithM(16).WithEfConstruction(200),
		},
		// 嵌入模型
		Embedding: embedder,
	}

	// 创建 Milvus 索引器
	indexer, err := milvus2.NewIndexer(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建 milvus 索引器失败: %w", err)
	}

	// 返回包装后的索引器实例
	return &milvusIndexer{
		indexer: indexer,
	}, nil
}

// milvusIndexer Milvus 索引器实现
//
// 实现了 indexer.Indexer 接口，提供基于 Milvus 的向量存储能力。
// 包装了官方的 Milvus 索引器，添加了自动 ID 生成功能。
type milvusIndexer struct {
	// indexer 官方 Milvus 索引器实例
	indexer indexer.Indexer
}

// Store 将文档存储到 Milvus
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
//  1. 为每个没有 ID 的文档生成 UUID
//  2. 调用官方索引器的 Store 方法存储文档
//
// 注意事项:
//   - 文档 ID 如果为空，会自动生成 UUID
//   - 向量化由官方索引器自动处理
func (m *milvusIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	// 为没有 ID 的文档生成 UUID
	for _, doc := range docs {
		if doc.ID == "" {
			doc.ID = uuid.New().String()
		}
	}
	// 调用官方索引器的 Store 方法
	return m.indexer.Store(ctx, docs, opts...)
}
