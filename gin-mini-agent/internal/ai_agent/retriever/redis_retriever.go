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

const (
	ContentField  = "content"
	MetadataField = "metadata"
	VectorField   = "content_vector"
	DistanceField = "distance"
)

func NewRedisRetriever(ctx context.Context, embedder embedding.Embedder) (rtr retriever.Retriever, err error) {
	redisClient := redisCli.NewClient(&redisCli.Options{
		Addr:     global.Conf.RAG.Redis.Addr,
		Protocol: 2,
	})

	config := &redis.RetrieverConfig{
		Client:       redisClient,
		Index:        fmt.Sprintf("%svector_index", global.Conf.RAG.Redis.Prefix),
		Dialect:      2,
		ReturnFields: []string{ContentField, MetadataField, DistanceField},
		TopK:         8,
		VectorField:  VectorField,
		DocumentConverter: func(ctx context.Context, doc redisCli.Document) (*schema.Document, error) {
			resp := &schema.Document{
				ID:       doc.ID,
				Content:  "",
				MetaData: map[string]any{},
			}
			for field, val := range doc.Fields {
				if field == ContentField {
					resp.Content = val
				} else if field == MetadataField {
					var metadata map[string]any
					if err := json.Unmarshal([]byte(val), &metadata); err != nil {
						resp.MetaData[field] = val
					} else {
						for k, v := range metadata {
							resp.MetaData[k] = v
						}
					}
				} else if field == "vector_distance" {
					distance, err := strconv.ParseFloat(val, 64)
					if err != nil {
						continue
					}
					resp.WithScore(1 - distance)
				}
			}

			return resp, nil
		},
	}
	config.Embedding = embedder
	rtr, err = redis.NewRetriever(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建 redis 检索器失败: %w", err)
	}
	return rtr, nil
}
