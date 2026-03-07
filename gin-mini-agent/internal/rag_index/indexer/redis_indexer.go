package indexer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	redisCli "github.com/redis/go-redis/v9"

	"gin-mini-agent/pkg/global"
)

const (
	ContentField  = "content"
	MetadataField = "metadata"
	VectorField   = "vector"
)

func NewRedisIndexer(ctx context.Context, embedder embedding.Embedder) (idr indexer.Indexer, err error) {
	redisClient := redisCli.NewClient(&redisCli.Options{
		Addr:     global.Conf.RAG.Redis.Addr,
		Protocol: 2,
	})

	config := &redis.IndexerConfig{
		Client:    redisClient,
		KeyPrefix: global.Conf.RAG.Redis.Prefix,
		BatchSize: 1,
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redis.Hashes, error) {
			if doc.ID == "" {
				doc.ID = uuid.New().String()
			}
			key := doc.ID

			metadataBytes, err := json.Marshal(doc.MetaData)
			if err != nil {
				return nil, fmt.Errorf("序列化元数据失败: %w", err)
			}

			return &redis.Hashes{
				Key: key,
				Field2Value: map[string]redis.FieldValue{
					ContentField:  {Value: doc.Content, EmbedKey: VectorField},
					MetadataField: {Value: metadataBytes},
				},
			}, nil
		},
	}

	config.Embedding = embedder
	idr, err = redis.NewIndexer(ctx, config)
	if err != nil {
		return nil, err
	}
	return &redisIndexer{
		indexer: idr,
	}, nil
}

type redisIndexer struct {
	indexer indexer.Indexer
}

func (r *redisIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	return r.indexer.Store(ctx, docs, opts...)
}
