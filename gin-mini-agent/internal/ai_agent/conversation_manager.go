/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ai_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"time"

	milvusindexer "github.com/cloudwego/eino-ext/components/indexer/milvus2"
	redisindexer "github.com/cloudwego/eino-ext/components/indexer/redis"
	milvusretriever "github.com/cloudwego/eino-ext/components/retriever/milvus2"
	redisretriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/philippgille/chromem-go"
	redisCli "github.com/redis/go-redis/v9"

	"gin-mini-agent/pkg/global"
)

// ConversationManager 对话历史管理器
type ConversationManager struct {
	indexer   indexer.Indexer
	retriever retriever.Retriever
}

// NewConversationManager 创建对话历史管理器
func NewConversationManager(ctx context.Context, embedder embedding.Embedder) (*ConversationManager, error) {
	dbType := global.Conf.RAG.Type
	if dbType == "" {
		dbType = "chromem"
	}

	switch dbType {
	case "chromem":
		return newChromemConversationManager(ctx, embedder)
	case "milvus":
		return newMilvusConversationManager(ctx, embedder)
	case "redis":
		return newRedisConversationManager(ctx, embedder)
	default:
		return newChromemConversationManager(ctx, embedder)
	}
}

// Store 存储对话历史
func (cm *ConversationManager) Store(ctx context.Context, userQuery string, aiResponse string) error {
	content := fmt.Sprintf("用户: %s\n助手: %s", userQuery, aiResponse)
	doc := &schema.Document{
		ID:      fmt.Sprintf("conv_%d", time.Now().UnixNano()),
		Content: content,
		MetaData: map[string]any{
			"type":      "conversation",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
	_, err := cm.indexer.Store(ctx, []*schema.Document{doc})
	return err
}

// Retrieve 检索对话历史
func (cm *ConversationManager) Retrieve(ctx context.Context, query string) ([]*schema.Document, error) {
	return cm.retriever.Retrieve(ctx, query)
}

// ==================== Chromem 实现 ====================

type chromemConversationIndexer struct {
	collection *chromem.Collection
	embedder   embedding.Embedder
}

func newChromemConversationManager(ctx context.Context, embedder embedding.Embedder) (*ConversationManager, error) {
	chromemPath := global.Conf.RAG.Chromem.Path
	if chromemPath == "" {
		chromemPath = "./data/chromem"
	}

	collectionName := global.Conf.RAG.Chromem.Collection + "_conversation"
	if collectionName == "" {
		collectionName = "rag_collection_conversation"
	}

	db, err := chromem.NewPersistentDB(chromemPath, true)
	if err != nil {
		return nil, fmt.Errorf("创建持久化 chromem 数据库失败: %w", err)
	}

	collection, err := db.GetOrCreateCollection(collectionName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("获取或创建对话历史集合失败: %w", err)
	}

	indexer := &chromemConversationIndexer{
		collection: collection,
		embedder:   embedder,
	}

	retriever := &chromemConversationRetriever{
		collection: collection,
		embedder:   embedder,
	}

	return &ConversationManager{
		indexer:   indexer,
		retriever: retriever,
	}, nil
}

func (ci *chromemConversationIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	ids := make([]string, 0, len(docs))

	for _, doc := range docs {
		vecs, err := ci.embedder.EmbedStrings(ctx, []string{doc.Content})
		if err != nil {
			return nil, fmt.Errorf("生成向量失败: %w", err)
		}

		if len(vecs) == 0 || len(vecs[0]) == 0 {
			return nil, fmt.Errorf("生成的向量为空")
		}

		float32Vec := make([]float32, len(vecs[0]))
		for i, v := range vecs[0] {
			float32Vec[i] = float32(v)
		}

		metadata := make(map[string]string)
		for k, v := range doc.MetaData {
			metadata[k] = fmt.Sprintf("%v", v)
		}

		err = ci.collection.AddDocuments(ctx, []chromem.Document{
			{
				ID:        doc.ID,
				Content:   doc.Content,
				Embedding: float32Vec,
				Metadata:  metadata,
			},
		}, runtime.NumCPU())
		if err != nil {
			return nil, fmt.Errorf("存储文档失败: %w", err)
		}

		ids = append(ids, doc.ID)
	}

	return ids, nil
}

type chromemConversationRetriever struct {
	collection *chromem.Collection
	embedder   embedding.Embedder
}

func (cr *chromemConversationRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	defaultTopK := 5
	options := retriever.GetCommonOptions(&retriever.Options{
		TopK: &defaultTopK,
	}, opts...)
	topK := *options.TopK

	vec, err := cr.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("嵌入查询失败: %w", err)
	}

	if len(vec) == 0 || len(vec[0]) == 0 {
		return nil, fmt.Errorf("生成的查询向量为空")
	}

	float32Vec := make([]float32, len(vec[0]))
	for i, v := range vec[0] {
		float32Vec[i] = float32(v)
	}

	count := cr.collection.Count()
	if topK > count {
		topK = count
	}

	if count == 0 {
		return []*schema.Document{}, nil
	}

	results, err := cr.collection.QueryEmbedding(ctx, float32Vec, topK, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("查询对话历史失败: %w", err)
	}

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
		doc.WithScore(float64(1 - result.Similarity))
		documents = append(documents, doc)
	}

	return documents, nil
}

// ==================== Redis 实现 ====================

const (
	redisContentField  = "content"
	redisMetadataField = "metadata"
	redisVectorField   = "content_vector"
	redisDistanceField = "distance"
)

func newRedisConversationManager(ctx context.Context, embedder embedding.Embedder) (*ConversationManager, error) {
	redisClient := redisCli.NewClient(&redisCli.Options{
		Addr:     global.Conf.RAG.Redis.Addr,
		Protocol: 2,
	})

	conversationPrefix := global.Conf.RAG.Redis.Prefix + "conv_"

	indexerConfig := &redisindexer.IndexerConfig{
		Client:    redisClient,
		KeyPrefix: conversationPrefix,
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redisindexer.Hashes, error) {
			if doc.ID == "" {
				return nil, fmt.Errorf("文档 ID 不能为空")
			}

			field2Value := map[string]redisindexer.FieldValue{
				redisContentField: {
					Value:     doc.Content,
					EmbedKey:  redisVectorField,
					Stringify: nil,
				},
				redisMetadataField: {
					Value: doc.MetaData,
					Stringify: func(val any) (string, error) {
						b, err := json.Marshal(val)
						if err != nil {
							return "", err
						}
						return string(b), nil
					},
				},
			}

			return &redisindexer.Hashes{
				Key:         doc.ID,
				Field2Value: field2Value,
			}, nil
		},
		BatchSize: 10,
		Embedding: embedder,
	}

	redisIndexer, err := redisindexer.NewIndexer(ctx, indexerConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Redis 对话历史索引器失败: %w", err)
	}

	retrieverConfig := &redisretriever.RetrieverConfig{
		Client:       redisClient,
		Index:        fmt.Sprintf("%svector_index", conversationPrefix),
		Dialect:      2,
		ReturnFields: []string{redisContentField, redisMetadataField, redisDistanceField},
		TopK:         5,
		VectorField:  redisVectorField,
		DocumentConverter: func(ctx context.Context, doc redisCli.Document) (*schema.Document, error) {
			resp := &schema.Document{
				ID:       doc.ID,
				Content:  "",
				MetaData: map[string]any{},
			}
			for field, val := range doc.Fields {
				if field == redisContentField {
					resp.Content = val
				} else if field == redisMetadataField {
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
		Embedding: embedder,
	}

	redisRetriever, err := redisretriever.NewRetriever(ctx, retrieverConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Redis 对话历史检索器失败: %w", err)
	}

	return &ConversationManager{
		indexer:   redisIndexer,
		retriever: redisRetriever,
	}, nil
}

// ==================== Milvus 实现 ====================

const (
	milvusIDField       = "id"
	milvusContentField  = "content"
	milvusVectorField   = "vector"
	milvusMetadataField = "metadata"
)

func newMilvusConversationManager(ctx context.Context, embedder embedding.Embedder) (*ConversationManager, error) {
	collectionName := global.Conf.RAG.Milvus.Collection + "_conversation"
	if collectionName == "" {
		collectionName = "rag_collection_conversation"
	}

	// 获取向量维度
	vecs, err := embedder.EmbedStrings(ctx, []string{"test"})
	if err != nil {
		return nil, fmt.Errorf("获取向量维度失败: %w", err)
	}
	if len(vecs) == 0 || len(vecs[0]) == 0 {
		return nil, fmt.Errorf("无法获取向量维度")
	}
	dimension := int64(len(vecs[0]))

	// 创建 Milvus indexer
	indexerConfig := &milvusindexer.IndexerConfig{
		ClientConfig: &milvusclient.ClientConfig{
			Address:  global.Conf.RAG.Milvus.Addr,
			Username: global.Conf.RAG.Milvus.Username,
			Password: global.Conf.RAG.Milvus.Password,
		},
		Collection:          collectionName,
		Description:         "对话历史集合",
		ConsistencyLevel:    milvusindexer.ConsistencyLevelBounded,
		EnableDynamicSchema: true,
		Vector: &milvusindexer.VectorConfig{
			Dimension:    dimension,
			MetricType:   milvusindexer.COSINE,
			VectorField:  milvusVectorField,
			IndexBuilder: milvusindexer.NewHNSWIndexBuilder(),
		},
		Embedding: embedder,
		DocumentConverter: func(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]column.Column, error) {
			ids := make([]string, 0, len(docs))
			contents := make([]string, 0, len(docs))
			vecs := make([][]float32, 0, len(docs))
			metadatas := make([][]byte, 0, len(docs))

			for idx, doc := range docs {
				ids = append(ids, doc.ID)
				contents = append(contents, doc.Content)

				var sourceVec []float64
				if len(vectors) == len(docs) {
					sourceVec = vectors[idx]
				}

				if len(sourceVec) == 0 {
					return nil, fmt.Errorf("向量数据缺失，文档 ID: %s", doc.ID)
				}

				vec := make([]float32, len(sourceVec))
				for i, v := range sourceVec {
					vec[i] = float32(v)
				}
				vecs = append(vecs, vec)

				metadata, err := json.Marshal(doc.MetaData)
				if err != nil {
					return nil, fmt.Errorf("序列化元数据失败: %w", err)
				}
				metadatas = append(metadatas, metadata)
			}

			return []column.Column{
				column.NewColumnVarChar(milvusIDField, ids),
				column.NewColumnVarChar(milvusContentField, contents),
				column.NewColumnFloatVector(milvusVectorField, int(dimension), vecs),
				column.NewColumnJSONBytes(milvusMetadataField, metadatas),
			}, nil
		},
	}

	milvusIndexer, err := milvusindexer.NewIndexer(ctx, indexerConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Milvus 对话历史索引器失败: %w", err)
	}

	// 创建 Milvus retriever
	retrieverConfig := &milvusretriever.RetrieverConfig{
		ClientConfig: &milvusclient.ClientConfig{
			Address:  global.Conf.RAG.Milvus.Addr,
			Username: global.Conf.RAG.Milvus.Username,
			Password: global.Conf.RAG.Milvus.Password,
		},
		Collection: collectionName,
		TopK:       5,
		Embedding:  embedder,
		DocumentConverter: func(ctx context.Context, result milvusclient.ResultSet) ([]*schema.Document, error) {
			docs := make([]*schema.Document, 0, result.ResultCount)

			for i := 0; i < result.ResultCount; i++ {
				doc := &schema.Document{
					MetaData: make(map[string]any),
				}

				if result.IDs != nil {
					if id, err := result.IDs.Get(i); err == nil {
						if idStr, ok := id.(string); ok {
							doc.ID = idStr
						}
					}
				}

				if contentCol := result.GetColumn(milvusContentField); contentCol != nil {
					if content, err := contentCol.Get(i); err == nil {
						if contentStr, ok := content.(string); ok {
							doc.Content = contentStr
						}
					}
				}

				if metadataCol := result.GetColumn(milvusMetadataField); metadataCol != nil {
					if metadata, err := metadataCol.Get(i); err == nil {
						if metadataMap, ok := metadata.(map[string]any); ok {
							for k, v := range metadataMap {
								doc.MetaData[k] = v
							}
						}
					}
				}

				if i < len(result.Scores) {
					doc.WithScore(float64(result.Scores[i]))
				}

				docs = append(docs, doc)
			}

			return docs, nil
		},
	}

	milvusRetriever, err := milvusretriever.NewRetriever(ctx, retrieverConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Milvus 对话历史检索器失败: %w", err)
	}

	return &ConversationManager{
		indexer:   milvusIndexer,
		retriever: milvusRetriever,
	}, nil
}

// ==================== 辅助函数 ====================

// createMilvusCollection 创建 Milvus 集合（如果不存在）
func createMilvusCollection(ctx context.Context, client *milvusclient.Client, collectionName string, dimension int64) error {
	hasCollection, err := client.HasCollection(ctx, milvusclient.NewHasCollectionOption(collectionName))
	if err != nil {
		return fmt.Errorf("检查集合失败: %w", err)
	}

	if !hasCollection {
		schema := entity.NewSchema().
			WithField(entity.NewField().
				WithName(milvusIDField).
				WithDataType(entity.FieldTypeVarChar).
				WithMaxLength(256).
				WithIsPrimaryKey(true)).
			WithField(entity.NewField().
				WithName(milvusContentField).
				WithDataType(entity.FieldTypeVarChar).
				WithMaxLength(65535)).
			WithField(entity.NewField().
				WithName(milvusVectorField).
				WithDataType(entity.FieldTypeFloatVector).
				WithDim(dimension)).
			WithField(entity.NewField().
				WithName(milvusMetadataField).
				WithDataType(entity.FieldTypeJSON)).
			WithDynamicFieldEnabled(true)

		err = client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(collectionName, schema))
		if err != nil {
			return fmt.Errorf("创建集合失败: %w", err)
		}

		// 创建向量索引
		idx := index.NewHNSWIndex(entity.COSINE, 16, 256)
		_, err = client.CreateIndex(ctx, milvusclient.NewCreateIndexOption(collectionName, milvusVectorField, idx))
		if err != nil {
			return fmt.Errorf("创建索引失败: %w", err)
		}

		// 加载集合
		loadTask, err := client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collectionName))
		if err != nil {
			return fmt.Errorf("加载集合失败: %w", err)
		}
		if err := loadTask.Await(ctx); err != nil {
			return fmt.Errorf("等待集合加载失败: %w", err)
		}
	}

	return nil
}
