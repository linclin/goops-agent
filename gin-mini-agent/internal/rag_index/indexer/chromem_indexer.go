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

func NewChromemIndexer(ctx context.Context, embedder embedding.Embedder) (idr indexer.Indexer, err error) {
	chromemPath := global.Conf.RAG.Chromem.Path
	if chromemPath == "" {
		chromemPath = "./data/chromem"
	}

	collectionName := global.Conf.RAG.Chromem.Collection
	if collectionName == "" {
		collectionName = "rag_collection"
	}

	// 使用持久化数据库，压缩存储
	db, err := chromem.NewPersistentDB(chromemPath, true)
	if err != nil {
		return nil, fmt.Errorf("创建持久化 chromem 数据库失败: %w", err)
	}

	collection, err := db.GetOrCreateCollection(collectionName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("获取或创建集合失败: %w", err)
	}

	return &chromemIndexer{
		db:         db,
		collection: collection,
		embedder:   embedder,
	}, nil
}

type chromemIndexer struct {
	db         *chromem.DB
	collection *chromem.Collection
	embedder   embedding.Embedder
}

func (c *chromemIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	var ids []string
	var documents []string
	var metadatas []map[string]string
	var embeddings [][]float32

	for _, doc := range docs {
		if doc.ID == "" {
			doc.ID = uuid.New().String()
		}
		ids = append(ids, doc.ID)
		documents = append(documents, doc.Content)

		metadata := make(map[string]string)
		for k, v := range doc.MetaData {
			metadata[k] = fmt.Sprintf("%v", v)
		}
		metadatas = append(metadatas, metadata)

		vec, err := c.embedder.EmbedStrings(ctx, []string{doc.Content})
		if err != nil {
			return nil, fmt.Errorf("嵌入文档失败: %w", err)
		}
		if len(vec) > 0 {
			float32Vec := make([]float32, len(vec[0]))
			for i, v := range vec[0] {
				float32Vec[i] = float32(v)
			}
			embeddings = append(embeddings, float32Vec)
		}
	}

	err := c.collection.Add(ctx, ids, embeddings, metadatas, documents)
	if err != nil {
		return nil, fmt.Errorf("添加文档到 chromem 失败: %w", err)
	}

	return ids, nil
}
