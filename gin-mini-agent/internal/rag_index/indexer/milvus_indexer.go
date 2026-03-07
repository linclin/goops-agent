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

func NewMilvusIndexer(ctx context.Context, embedder embedding.Embedder) (idr indexer.Indexer, err error) {
	config := &milvus2.IndexerConfig{
		ClientConfig: &milvusclient.ClientConfig{
			Address:  global.Conf.RAG.Milvus.Addr,
			Username: global.Conf.RAG.Milvus.Username,
			Password: global.Conf.RAG.Milvus.Password,
		},
		Collection: global.Conf.RAG.Milvus.Collection,
		Vector: &milvus2.VectorConfig{
			Dimension:    1024,
			MetricType:   milvus2.COSINE,
			IndexBuilder: milvus2.NewHNSWIndexBuilder().WithM(16).WithEfConstruction(200),
		},
		Embedding: embedder,
	}

	indexer, err := milvus2.NewIndexer(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建 milvus 索引器失败: %w", err)
	}

	return &milvusIndexer{
		indexer: indexer,
	}, nil
}

type milvusIndexer struct {
	indexer indexer.Indexer
}

func (m *milvusIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	for _, doc := range docs {
		if doc.ID == "" {
			doc.ID = uuid.New().String()
		}
	}
	return m.indexer.Store(ctx, docs, opts...)
}
