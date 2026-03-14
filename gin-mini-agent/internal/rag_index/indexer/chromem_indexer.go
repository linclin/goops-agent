// Package indexer 提供向量数据库索引器实现
//
// 该包实现了多种向量数据库的索引器，用于将文档向量化并存储。
// 索引器是 RAG（检索增强生成）系统的核心组件。
//
// 支持的向量数据库:
//   - Chromem: 本地文件存储，适合开发和小规模部署
//   - Redis: 分布式存储，适合中等规模部署
//   - Milvus: 分布式向量数据库，适合大规模部署
//
// 索引流程:
//  1. 接收文档列表
//  2. 为每个文档生成唯一 ID
//  3. 使用嵌入模型将文档内容转换为向量
//  4. 将文档 ID、内容、向量和元数据存储到向量数据库
package indexer

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/philippgille/chromem-go"

	"gin-mini-agent/pkg/global"
)

// NewChromemIndexer 创建 Chromem 索引器
//
// Chromem 是一个轻量级的本地向量数据库，数据存储在本地文件中。
// 适合开发环境和小规模部署场景。
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
//   - chromem.path: 数据存储路径，默认 "./data/chromem"
//   - chromem.collection: 集合名称，默认 "rag_collection"
//
// 使用示例:
//
//	indexer, err := NewChromemIndexer(ctx, embedder)
//	ids, err := indexer.Store(ctx, []*schema.Document{...})
func NewChromemIndexer(ctx context.Context, embedder embedding.Embedder) (idr indexer.Indexer, err error) {
	// 获取 Chromem 数据存储路径
	chromemPath := global.Conf.RAG.Chromem.Path
	if chromemPath == "" {
		chromemPath = "./data/chromem"
	}

	// 获取集合名称
	collectionName := global.Conf.RAG.Chromem.Collection
	if collectionName == "" {
		collectionName = "rag_collection"
	}

	// 创建持久化数据库
	// 第二个参数 true 表示压缩存储，节省磁盘空间
	db, err := chromem.NewPersistentDB(chromemPath, true)
	if err != nil {
		return nil, fmt.Errorf("创建持久化 chromem 数据库失败: %w", err)
	}

	// 获取或创建集合
	// 如果集合不存在，会自动创建
	collection, err := db.GetOrCreateCollection(collectionName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("获取或创建集合失败: %w", err)
	}

	// 返回索引器实例
	return &chromemIndexer{
		db:         db,
		collection: collection,
		embedder:   embedder,
	}, nil
}

// chromemIndexer Chromem 索引器实现
//
// 实现了 indexer.Indexer 接口，提供基于 Chromem 的向量存储能力。
type chromemIndexer struct {
	// db Chromem 数据库实例
	db *chromem.DB

	// collection Chromem 集合，存储向量数据
	collection *chromem.Collection

	// embedder 嵌入模型，用于将文本转换为向量
	embedder embedding.Embedder
}

// Store 将文档存储到 Chromem
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
//  1. 为每个文档生成唯一 ID（如果没有）
//  2. 将文档内容转换为向量
//  3. 将文档 ID、向量、元数据和内容添加到集合
//
// 注意事项:
//   - 文档 ID 如果为空，会自动生成 UUID
//   - 元数据中的值会被转换为字符串格式
func (c *chromemIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	var ids []string
	var documents []string
	var metadatas []map[string]string
	var embeddings [][]float32

	// 遍历文档，准备存储数据
	for _, doc := range docs {
		// 如果文档没有 ID，生成 UUID
		if doc.ID == "" {
			doc.ID = uuid.New().String()
		}
		ids = append(ids, doc.ID)
		documents = append(documents, doc.Content)

		// 转换元数据为字符串格式
		// Chromem 要求元数据值为字符串类型
		metadata := make(map[string]string)
		for k, v := range doc.MetaData {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		metadatas = append(metadatas, metadata)

		// 使用嵌入模型将文档内容转换为向量
		vec, err := c.embedder.EmbedStrings(ctx, []string{doc.Content})
		if err != nil {
			return nil, fmt.Errorf("嵌入文档失败: %w", err)
		}

		// 将 float64 向量转换为 float32（Chromem 要求）
		if len(vec) > 0 {
			float32Vec := make([]float32, len(vec[0]))
			for i, v := range vec[0] {
				float32Vec[i] = float32(v)
			}
			embeddings = append(embeddings, float32Vec)
		}
	}

	// 将文档添加到集合
	// 参数: 上下文、ID 列表、向量列表、元数据列表、内容列表
	err := c.collection.Add(ctx, ids, embeddings, metadatas, documents)
	if err != nil {
		return nil, fmt.Errorf("添加文档到 chromem 失败: %w", err)
	}

	return ids, nil
}
