// Package retriever 提供向量数据库检索器实现
//
// 该包实现了多种向量数据库的检索器，用于从知识库中检索相关文档。
// 检索器是 RAG（检索增强生成）系统的核心组件。
//
// 支持的向量数据库:
//   - Chromem: 本地文件存储，适合开发和小规模部署
//   - Redis: 分布式存储，适合中等规模部署
//   - Milvus: 分布式向量数据库，适合大规模部署
//
// 检索流程:
//  1. 将用户查询转换为向量（使用嵌入模型）
//  2. 在向量数据库中搜索相似的文档向量
//  3. 返回最相关的文档片段
package retriever

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/philippgille/chromem-go"

	"gin-mini-agent/pkg/global"
)

// NewChromemRetriever 创建 Chromem 检索器
//
// Chromem 是一个轻量级的本地向量数据库，数据存储在本地文件中。
// 适合开发环境和小规模部署场景。
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
//   - chromem.path: 数据存储路径，默认 "./data/chromem"
//   - chromem.collection: 集合名称，默认 "rag_collection"
//
// 使用示例:
//
//	retriever, err := NewChromemRetriever(ctx, embedder)
//	docs, err := retriever.Retrieve(ctx, "什么是机器学习？")
func NewChromemRetriever(ctx context.Context, embedder embedding.Embedder) (rtr retriever.Retriever, err error) {
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

	// 返回检索器实例
	return &chromemRetriever{
		collection: collection,
		embedder:   embedder,
	}, nil
}

// chromemRetriever Chromem 检索器实现
//
// 实现了 retriever.Retriever 接口，提供基于 Chromem 的向量检索能力。
type chromemRetriever struct {
	// collection Chromem 集合，存储向量数据
	collection *chromem.Collection

	// embedder 嵌入模型，用于将文本转换为向量
	embedder embedding.Embedder
}

// Retrieve 从 Chromem 中检索相关文档
//
// 该方法实现了 retriever.Retriever 接口，执行向量相似度检索。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - query: 用户查询字符串
//   - opts: 可选配置，如 TopK（返回结果数量）
//
// 返回:
//   - []*schema.Document: 检索到的文档列表，按相似度排序
//   - error: 检索过程中的错误
//
// 检索流程:
//  1. 将查询字符串转换为向量
//  2. 在集合中搜索最相似的文档
//  3. 转换结果为 schema.Document 格式
//
// 注意事项:
//   - 如果集合为空，返回空列表
//   - TopK 不能超过集合中的文档数量
func (c *chromemRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 设置默认 TopK 为 8
	defaultTopK := 8
	options := retriever.GetCommonOptions(&retriever.Options{
		TopK: &defaultTopK,
	}, opts...)
	topK := *options.TopK

	// 使用嵌入模型将查询转换为向量
	vec, err := c.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("嵌入查询失败: %w", err)
	}

	// 验证向量是否有效
	if len(vec) == 0 || len(vec[0]) == 0 {
		return nil, fmt.Errorf("生成的查询向量为空")
	}

	// 将 float64 向量转换为 float32（Chromem 要求）
	float32Vec := make([]float32, len(vec[0]))
	for i, v := range vec[0] {
		float32Vec[i] = float32(v)
	}

	// 获取集合中的文档数量
	count := c.collection.Count()

	// 确保 topK 不超过集合中的文档数量
	if topK > count {
		topK = count
	}

	// 如果集合为空，直接返回空结果
	if count == 0 {
		return []*schema.Document{}, nil
	}

	// 执行向量检索
	// QueryEmbedding 使用向量相似度搜索最相关的文档
	results, err := c.collection.QueryEmbedding(ctx, float32Vec, topK, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("查询 chromem 失败: %w", err)
	}

	// 转换结果为 schema.Document 格式
	documents := make([]*schema.Document, 0, len(results))
	for _, result := range results {
		doc := &schema.Document{
			ID:       result.ID,
			Content:  result.Content,
			MetaData: make(map[string]any),
		}
		// 复制元数据
		for k, v := range result.Metadata {
			doc.MetaData[k] = v
		}
		// chromem 返回的是相似度分数（距离），需要转换为 score
		// score = 1 - distance，值越大表示越相似
		doc.WithScore(float64(1 - result.Similarity))
		documents = append(documents, doc)
	}

	return documents, nil
}
