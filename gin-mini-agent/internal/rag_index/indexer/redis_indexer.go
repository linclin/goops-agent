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

const (
	ContentField  = "content"
	MetadataField = "metadata"
	VectorField   = "content_vector"
)

func InitRedisIndex(ctx context.Context, client *redisCli.Client) (err error) {
	if err = client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	indexName := fmt.Sprintf("%s%s", global.Conf.RAG.Redis.Prefix, "vector_index")

	// 检查是否存在索引
	exists, err := client.Do(ctx, "FT.INFO", indexName).Result()
	if err != nil {
		if !strings.Contains(err.Error(), "Unknown index name") {
			return fmt.Errorf("failed to check if index exists: %w", err)
		}
		err = nil
	} else if exists != nil {
		return nil
	}

	// Create new index
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

	if err = client.Do(ctx, createIndexArgs...).Err(); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// 验证索引是否创建成功
	if _, err = client.Do(ctx, "FT.INFO", indexName).Result(); err != nil {
		return fmt.Errorf("failed to verify index creation: %w", err)
	}

	return nil
}

func NewRedisIndexer(ctx context.Context, embedder embedding.Embedder) (idr indexer.Indexer, err error) {
	redisClient := redisCli.NewClient(&redisCli.Options{
		Addr:     global.Conf.RAG.Redis.Addr,
		Protocol: 2,
	})
	defer func() {
		if err != nil {
			redisClient.Close()
		}
	}()
	if err = InitRedisIndex(ctx, redisClient); err != nil {
		return nil, err
	}
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
