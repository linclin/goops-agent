package retriever

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/milvusclient"

	milvus2 "github.com/cloudwego/eino-ext/components/retriever/milvus2"

	"gin-mini-agent/pkg/global"
)

func NewMilvusRetriever(ctx context.Context, embedder embedding.Embedder) (rtr retriever.Retriever, err error) {
	config := &milvus2.RetrieverConfig{
		ClientConfig: &milvusclient.ClientConfig{
			Address:  global.Conf.RAG.Milvus.Addr,
			Username: global.Conf.RAG.Milvus.Username,
			Password: global.Conf.RAG.Milvus.Password,
		},
		Collection: global.Conf.RAG.Milvus.Collection,
		TopK:       8,
		Embedding:  embedder,
		DocumentConverter: func(ctx context.Context, result milvusclient.ResultSet) ([]*schema.Document, error) {
			docs := make([]*schema.Document, 0, result.ResultCount)

			for i := 0; i < result.ResultCount; i++ {
				doc := &schema.Document{
					MetaData: make(map[string]any),
				}

				// 提取 ID
				if result.IDs != nil {
					if id, err := result.IDs.Get(i); err == nil {
						if idStr, ok := id.(string); ok {
							doc.ID = idStr
						}
					}
				}

				// 提取内容
				if contentCol := result.GetColumn("content"); contentCol != nil {
					if content, err := contentCol.Get(i); err == nil {
						if contentStr, ok := content.(string); ok {
							doc.Content = contentStr
						}
					}
				}

				// 提取元数据
				if metadataCol := result.GetColumn("metadata"); metadataCol != nil {
					if metadata, err := metadataCol.Get(i); err == nil {
						if metadataMap, ok := metadata.(map[string]any); ok {
							for k, v := range metadataMap {
								doc.MetaData[k] = v
							}
						}
					}
				}

				// 提取分数
				if i < len(result.Scores) {
					doc.WithScore(float64(result.Scores[i]))
				}

				docs = append(docs, doc)
			}

			return docs, nil
		},
	}

	rtr, err = milvus2.NewRetriever(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建 milvus 检索器失败: %w", err)
	}

	return rtr, nil
}
