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

func NewChromemRetriever(ctx context.Context, embedder embedding.Embedder) (rtr retriever.Retriever, err error) {
	chromemPath := global.Conf.RAG.Chromem.Path
	if chromemPath == "" {
		chromemPath = "./data/chromem"
	}

	collectionName := global.Conf.RAG.Chromem.Collection
	if collectionName == "" {
		collectionName = "rag_collection"
	}

	// 使用持久化数据库
	db, err := chromem.NewPersistentDB(chromemPath, true)
	if err != nil {
		return nil, fmt.Errorf("创建持久化 chromem 数据库失败: %w", err)
	}

	collection, err := db.GetOrCreateCollection(collectionName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("获取或创建集合失败: %w", err)
	}

	return &chromemRetriever{
		collection: collection,
		embedder:   embedder,
	}, nil
}

type chromemRetriever struct {
	collection *chromem.Collection
	embedder   embedding.Embedder
}

func (c *chromemRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	defaultTopK := 8
	options := retriever.GetCommonOptions(&retriever.Options{
		TopK: &defaultTopK,
	}, opts...)
	topK := *options.TopK

	// 生成查询向量
	vec, err := c.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("嵌入查询失败: %w", err)
	}

	if len(vec) == 0 || len(vec[0]) == 0 {
		return nil, fmt.Errorf("生成的查询向量为空")
	}

	// 转换为 float32
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

	// 查询集合
	results, err := c.collection.QueryEmbedding(ctx, float32Vec, topK, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("查询 chromem 失败: %w", err)
	}

	// 转换为 schema.Document
	documents := make([]*schema.Document, 0, len(results))
	for _, result := range results {
		doc := &schema.Document{
			ID:       result.ID,
			Content:  result.Content,
			MetaData: make(map[string]any),
		}
		for k, v := range result.Metadata {
			doc.MetaData[k] = v
		}
		// chromem 返回的是相似度分数（距离），需要转换为 score
		doc.WithScore(float64(1 - result.Similarity))
		documents = append(documents, doc)
	}

	return documents, nil
}
