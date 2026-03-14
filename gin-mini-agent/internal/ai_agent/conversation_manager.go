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
//
// 该结构体负责管理对话历史的存储和检索，是 AI Agent 记忆能力的核心组件。
// 它使用向量数据库存储对话历史，支持语义检索，能够找到与当前问题相关的历史对话。
//
// 主要功能:
//   - Store: 存储对话历史到向量数据库
//   - Retrieve: 从向量数据库检索相关的对话历史
//
// 支持的向量数据库:
//   - Chromem: 本地文件存储，适合开发和小规模部署
//   - Redis: 分布式存储，适合中等规模部署
//   - Milvus: 分布式向量数据库，适合大规模部署
//
// 对话历史存储格式:
//
//	用户: {用户问题}
//	助手: {AI回复}
//
// 使用场景:
//   - 保持对话连贯性：AI 可以参考之前的对话内容
//   - 上下文理解：理解用户的问题与之前对话的关系
//   - 个性化回答：根据历史对话调整回答风格
type ConversationManager struct {
	// indexer 索引器，用于存储文档到向量数据库
	indexer indexer.Indexer

	// retriever 检索器，用于从向量数据库检索相关文档
	retriever retriever.Retriever
}

// NewConversationManager 创建对话历史管理器
//
// 该函数根据配置文件中的 RAG.Type 选择不同的向量数据库实现。
// 每种向量数据库都有其优缺点，适用于不同的场景。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - embedder: 嵌入模型，用于将文本转换为向量
//
// 返回:
//   - *ConversationManager: 对话历史管理器实例
//   - error: 创建过程中的错误
//
// 配置示例:
//
//	rag:
//	  type: "chromem"  # 可选: chromem, redis, milvus
func NewConversationManager(ctx context.Context, embedder embedding.Embedder) (*ConversationManager, error) {
	// 从配置文件读取向量数据库类型
	// 默认使用 Chromem（本地文件存储）
	dbType := global.Conf.RAG.Type
	if dbType == "" {
		dbType = "chromem"
	}

	// 根据配置选择不同的向量数据库实现
	switch dbType {
	case "chromem":
		// Chromem: 本地文件存储
		// 优点: 无需额外依赖，部署简单
		// 缺点: 不支持分布式，扩展性有限
		return newChromemConversationManager(ctx, embedder)
	case "milvus":
		// Milvus: 分布式向量数据库
		// 优点: 高性能，支持分布式，适合大规模数据
		// 缺点: 部署复杂，需要额外资源
		return newMilvusConversationManager(ctx, embedder)
	case "redis":
		// Redis: 分布式缓存数据库
		// 优点: 部署简单，性能好，支持持久化
		// 缺点: 向量检索能力有限
		return newRedisConversationManager(ctx, embedder)
	default:
		// 默认使用 Chromem
		return newChromemConversationManager(ctx, embedder)
	}
}

// Store 存储对话历史
//
// 将用户问题和 AI 回复组合后存储到向量数据库。
// 存储的内容会被向量化，以便后续进行语义检索。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - userQuery: 用户问题
//   - aiResponse: AI 回复
//
// 返回:
//   - error: 存储过程中的错误
//
// 存储格式:
//
//	用户: {用户问题}
//	助手: {AI回复}
//
// 元数据:
//   - type: "conversation" - 标识这是对话历史
//   - timestamp: 存储时间
//
// 使用示例:
//
//	err := cm.Store(ctx, "什么是机器学习？", "机器学习是人工智能的一个分支...")
func (cm *ConversationManager) Store(ctx context.Context, userQuery string, aiResponse string) error {
	// 组合用户问题和 AI 回复
	// 这种格式便于后续检索时理解对话上下文
	content := fmt.Sprintf("用户: %s\n助手: %s", userQuery, aiResponse)

	// 创建文档对象
	doc := &schema.Document{
		// ID: 使用时间戳生成唯一标识
		ID:      fmt.Sprintf("conv_%d", time.Now().UnixNano()),
		Content: content,
		// 元数据: 存储对话类型和时间戳
		MetaData: map[string]any{
			"type":      "conversation",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	// 调用索引器存储文档
	_, err := cm.indexer.Store(ctx, []*schema.Document{doc})
	return err
}

// Retrieve 检索对话历史
//
// 根据用户查询从向量数据库中检索相关的历史对话。
// 使用语义相似度进行检索，能够找到语义相关的历史对话。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - query: 用户查询
//
// 返回:
//   - []*schema.Document: 检索到的对话历史文档列表
//   - error: 检索过程中的错误
//
// 使用示例:
//
//	docs, err := cm.Retrieve(ctx, "机器学习")
//	for _, doc := range docs {
//	    fmt.Println(doc.Content)
//	}
func (cm *ConversationManager) Retrieve(ctx context.Context, query string) ([]*schema.Document, error) {
	return cm.retriever.Retrieve(ctx, query)
}

// ==================== Chromem 实现 ====================

// chromemConversationIndexer Chromem 对话历史索引器
//
// 使用 Chromem 本地向量数据库存储对话历史。
// Chromem 是一个轻量级的向量数据库，数据存储在本地文件中。
type chromemConversationIndexer struct {
	// collection Chromem 集合，用于存储向量数据
	collection *chromem.Collection

	// embedder 嵌入模型，用于将文本转换为向量
	embedder embedding.Embedder
}

// newChromemConversationManager 创建 Chromem 对话历史管理器
//
// 该函数创建一个使用 Chromem 作为存储后端的对话历史管理器。
// Chromem 是一个轻量级的本地向量数据库，适合开发和小规模部署。
//
// 参数:
//   - ctx: 上下文
//   - embedder: 嵌入模型
//
// 返回:
//   - *ConversationManager: 对话历史管理器
//   - error: 创建错误
//
// 配置:
//   - chromem.path: 数据存储路径，默认 "./data/chromem"
//   - chromem.collection: 集合名称，默认 "rag_collection"
func newChromemConversationManager(ctx context.Context, embedder embedding.Embedder) (*ConversationManager, error) {
	// 获取 Chromem 数据存储路径
	chromemPath := global.Conf.RAG.Chromem.Path
	if chromemPath == "" {
		chromemPath = "./data/chromem"
	}

	// 获取集合名称，添加 "_conversation" 后缀区分对话历史
	collectionName := global.Conf.RAG.Chromem.Collection + "_conversation"
	if collectionName == "" {
		collectionName = "rag_collection_conversation"
	}

	// 创建持久化数据库
	// 第二个参数 true 表示压缩存储
	db, err := chromem.NewPersistentDB(chromemPath, true)
	if err != nil {
		return nil, fmt.Errorf("创建持久化 chromem 数据库失败: %w", err)
	}

	// 获取或创建集合
	// 如果集合不存在，会自动创建
	collection, err := db.GetOrCreateCollection(collectionName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("获取或创建对话历史集合失败: %w", err)
	}

	// 创建索引器和检索器
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

// Store 存储文档到 Chromem
//
// 将文档向量化后存储到 Chromem 集合中。
//
// 参数:
//   - ctx: 上下文
//   - docs: 要存储的文档列表
//   - opts: 可选配置
//
// 返回:
//   - []string: 存储的文档 ID 列表
//   - error: 存储错误
func (ci *chromemConversationIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	ids := make([]string, 0, len(docs))

	for _, doc := range docs {
		// 使用嵌入模型将文档内容转换为向量
		vecs, err := ci.embedder.EmbedStrings(ctx, []string{doc.Content})
		if err != nil {
			return nil, fmt.Errorf("生成向量失败: %w", err)
		}

		// 验证向量是否有效
		if len(vecs) == 0 || len(vecs[0]) == 0 {
			return nil, fmt.Errorf("生成的向量为空")
		}

		// 将 float64 向量转换为 float32（Chromem 要求）
		float32Vec := make([]float32, len(vecs[0]))
		for i, v := range vecs[0] {
			float32Vec[i] = float32(v)
		}

		// 将元数据转换为字符串格式
		metadata := make(map[string]string)
		for k, v := range doc.MetaData {
			metadata[k] = fmt.Sprintf("%v", v)
		}

		// 添加文档到集合
		// runtime.NumCPU() 表示使用所有 CPU 核心并行处理
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

// chromemConversationRetriever Chromem 对话历史检索器
type chromemConversationRetriever struct {
	collection *chromem.Collection
	embedder   embedding.Embedder
}

// Retrieve 从 Chromem 检索相关文档
//
// 将查询向量化后，在 Chromem 集合中搜索相似的文档。
//
// 参数:
//   - ctx: 上下文
//   - query: 查询字符串
//   - opts: 可选配置（如 TopK）
//
// 返回:
//   - []*schema.Document: 检索到的文档列表
//   - error: 检索错误
func (cr *chromemConversationRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 设置默认 TopK 为 5
	defaultTopK := 5
	options := retriever.GetCommonOptions(&retriever.Options{
		TopK: &defaultTopK,
	}, opts...)
	topK := *options.TopK

	// 将查询转换为向量
	vec, err := cr.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("嵌入查询失败: %w", err)
	}

	// 验证向量是否有效
	if len(vec) == 0 || len(vec[0]) == 0 {
		return nil, fmt.Errorf("生成的查询向量为空")
	}

	// 转换为 float32
	float32Vec := make([]float32, len(vec[0]))
	for i, v := range vec[0] {
		float32Vec[i] = float32(v)
	}

	// 检查集合中的文档数量
	count := cr.collection.Count()
	// 如果 TopK 大于文档数量，调整 TopK
	if topK > count {
		topK = count
	}

	// 如果集合为空，返回空结果
	if count == 0 {
		return []*schema.Document{}, nil
	}

	// 执行向量检索
	results, err := cr.collection.QueryEmbedding(ctx, float32Vec, topK, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("查询对话历史失败: %w", err)
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
		// 设置相似度分数（1 - 距离 = 相似度）
		doc.WithScore(float64(1 - result.Similarity))
		documents = append(documents, doc)
	}

	return documents, nil
}

// ==================== Redis 实现 ====================

// Redis 字段常量定义
const (
	// redisContentField 内容字段名
	redisContentField = "content"
	// redisMetadataField 元数据字段名
	redisMetadataField = "metadata"
	// redisVectorField 向量字段名
	redisVectorField = "content_vector"
	// redisDistanceField 距离字段名
	redisDistanceField = "distance"
)

// newRedisConversationManager 创建 Redis 对话历史管理器
//
// 该函数创建一个使用 Redis 作为存储后端的对话历史管理器。
// Redis 是一个高性能的内存数据库，支持向量检索。
//
// 参数:
//   - ctx: 上下文
//   - embedder: 嵌入模型
//
// 返回:
//   - *ConversationManager: 对话历史管理器
//   - error: 创建错误
//
// 配置:
//   - redis.addr: Redis 地址，如 "localhost:6379"
//   - redis.prefix: 键前缀，用于区分不同应用
func newRedisConversationManager(ctx context.Context, embedder embedding.Embedder) (*ConversationManager, error) {
	// 创建 Redis 客户端
	redisClient := redisCli.NewClient(&redisCli.Options{
		Addr:     global.Conf.RAG.Redis.Addr,
		Protocol: 2, // 使用 RESP2 协议
	})

	// 对话历史键前缀，添加 "conv_" 后缀区分
	conversationPrefix := global.Conf.RAG.Redis.Prefix + "conv_"

	// 创建 Redis 索引器配置
	indexerConfig := &redisindexer.IndexerConfig{
		Client:    redisClient,
		KeyPrefix: conversationPrefix,
		// 文档到 Redis Hash 的转换函数
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redisindexer.Hashes, error) {
			if doc.ID == "" {
				return nil, fmt.Errorf("文档 ID 不能为空")
			}

			field2Value := map[string]redisindexer.FieldValue{
				// 内容字段：存储对话内容，并生成向量
				redisContentField: {
					Value:     doc.Content,
					EmbedKey:  redisVectorField, // 指定向量字段名
					Stringify: nil,              // 内容直接存储，不需要转换
				},
				// 元数据字段：存储 JSON 格式的元数据
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
		BatchSize: 10, // 批量处理大小
		Embedding: embedder,
	}

	// 创建 Redis 索引器
	redisIndexer, err := redisindexer.NewIndexer(ctx, indexerConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Redis 对话历史索引器失败: %w", err)
	}

	// 创建 Redis 检索器配置
	retrieverConfig := &redisretriever.RetrieverConfig{
		Client:       redisClient,
		Index:        fmt.Sprintf("%svector_index", conversationPrefix), // 向量索引名称
		Dialect:      2,                                                 // 使用 RedisSearch 2.x 语法
		ReturnFields: []string{redisContentField, redisMetadataField, redisDistanceField},
		TopK:         5,
		VectorField:  redisVectorField,
		// Redis 文档到 schema.Document 的转换函数
		DocumentConverter: func(ctx context.Context, doc redisCli.Document) (*schema.Document, error) {
			resp := &schema.Document{
				ID:       doc.ID,
				Content:  "",
				MetaData: map[string]any{},
			}
			// 遍历文档字段
			for field, val := range doc.Fields {
				if field == redisContentField {
					resp.Content = val
				} else if field == redisMetadataField {
					// 解析 JSON 元数据
					var metadata map[string]any
					if err := json.Unmarshal([]byte(val), &metadata); err != nil {
						resp.MetaData[field] = val
					} else {
						for k, v := range metadata {
							resp.MetaData[k] = v
						}
					}
				} else if field == "vector_distance" {
					// 计算相似度分数
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

	// 创建 Redis 检索器
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

// Milvus 字段常量定义
const (
	// milvusIDField ID 字段名
	milvusIDField = "id"
	// milvusContentField 内容字段名
	milvusContentField = "content"
	// milvusVectorField 向量字段名
	milvusVectorField = "vector"
	// milvusMetadataField 元数据字段名
	milvusMetadataField = "metadata"
)

// newMilvusConversationManager 创建 Milvus 对话历史管理器
//
// 该函数创建一个使用 Milvus 作为存储后端的对话历史管理器。
// Milvus 是一个高性能的分布式向量数据库，适合大规模数据场景。
//
// 参数:
//   - ctx: 上下文
//   - embedder: 嵌入模型
//
// 返回:
//   - *ConversationManager: 对话历史管理器
//   - error: 创建错误
//
// 配置:
//   - milvus.addr: Milvus 地址，如 "localhost:19530"
//   - milvus.username: 用户名（可选）
//   - milvus.password: 密码（可选）
//   - milvus.collection: 集合名称
func newMilvusConversationManager(ctx context.Context, embedder embedding.Embedder) (*ConversationManager, error) {
	// 获取集合名称，添加 "_conversation" 后缀
	collectionName := global.Conf.RAG.Milvus.Collection + "_conversation"
	if collectionName == "" {
		collectionName = "rag_collection_conversation"
	}

	// 获取向量维度
	// 通过嵌入一个测试字符串来获取向量维度
	vecs, err := embedder.EmbedStrings(ctx, []string{"test"})
	if err != nil {
		return nil, fmt.Errorf("获取向量维度失败: %w", err)
	}
	if len(vecs) == 0 || len(vecs[0]) == 0 {
		return nil, fmt.Errorf("无法获取向量维度")
	}
	dimension := int64(len(vecs[0]))

	// 创建 Milvus 索引器配置
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
			MetricType:   milvusindexer.COSINE, // 使用余弦相似度
			VectorField:  milvusVectorField,
			IndexBuilder: milvusindexer.NewHNSWIndexBuilder(), // 使用 HNSW 索引
		},
		Embedding: embedder,
		// 文档到 Milvus 列的转换函数
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

				// 转换为 float32
				vec := make([]float32, len(sourceVec))
				for i, v := range sourceVec {
					vec[i] = float32(v)
				}
				vecs = append(vecs, vec)

				// 序列化元数据
				metadata, err := json.Marshal(doc.MetaData)
				if err != nil {
					return nil, fmt.Errorf("序列化元数据失败: %w", err)
				}
				metadatas = append(metadatas, metadata)
			}

			// 返回 Milvus 列
			return []column.Column{
				column.NewColumnVarChar(milvusIDField, ids),
				column.NewColumnVarChar(milvusContentField, contents),
				column.NewColumnFloatVector(milvusVectorField, int(dimension), vecs),
				column.NewColumnJSONBytes(milvusMetadataField, metadatas),
			}, nil
		},
	}

	// 创建 Milvus 索引器
	milvusIndexer, err := milvusindexer.NewIndexer(ctx, indexerConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Milvus 对话历史索引器失败: %w", err)
	}

	// 创建 Milvus 检索器配置
	retrieverConfig := &milvusretriever.RetrieverConfig{
		ClientConfig: &milvusclient.ClientConfig{
			Address:  global.Conf.RAG.Milvus.Addr,
			Username: global.Conf.RAG.Milvus.Username,
			Password: global.Conf.RAG.Milvus.Password,
		},
		Collection: collectionName,
		TopK:       5,
		Embedding:  embedder,
		// Milvus 结果到 schema.Document 的转换函数
		DocumentConverter: func(ctx context.Context, result milvusclient.ResultSet) ([]*schema.Document, error) {
			docs := make([]*schema.Document, 0, result.ResultCount)

			for i := 0; i < result.ResultCount; i++ {
				doc := &schema.Document{
					MetaData: make(map[string]any),
				}

				// 获取 ID
				if result.IDs != nil {
					if id, err := result.IDs.Get(i); err == nil {
						if idStr, ok := id.(string); ok {
							doc.ID = idStr
						}
					}
				}

				// 获取内容
				if contentCol := result.GetColumn(milvusContentField); contentCol != nil {
					if content, err := contentCol.Get(i); err == nil {
						if contentStr, ok := content.(string); ok {
							doc.Content = contentStr
						}
					}
				}

				// 获取元数据
				if metadataCol := result.GetColumn(milvusMetadataField); metadataCol != nil {
					if metadata, err := metadataCol.Get(i); err == nil {
						if metadataMap, ok := metadata.(map[string]any); ok {
							for k, v := range metadataMap {
								doc.MetaData[k] = v
							}
						}
					}
				}

				// 设置相似度分数
				if i < len(result.Scores) {
					doc.WithScore(float64(result.Scores[i]))
				}

				docs = append(docs, doc)
			}

			return docs, nil
		},
	}

	// 创建 Milvus 检索器
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
//
// 该函数检查指定的集合是否存在，如果不存在则创建。
// 创建时会设置字段结构和向量索引。
//
// 参数:
//   - ctx: 上下文
//   - client: Milvus 客户端
//   - collectionName: 集合名称
//   - dimension: 向量维度
//
// 返回:
//   - error: 创建错误
//
// 集合结构:
//   - id: 主键，VarChar 类型
//   - content: 内容，VarChar 类型
//   - vector: 向量，FloatVector 类型
//   - metadata: 元数据，JSON 类型
func createMilvusCollection(ctx context.Context, client *milvusclient.Client, collectionName string, dimension int64) error {
	// 检查集合是否存在
	hasCollection, err := client.HasCollection(ctx, milvusclient.NewHasCollectionOption(collectionName))
	if err != nil {
		return fmt.Errorf("检查集合失败: %w", err)
	}

	// 如果集合已存在，直接返回
	if !hasCollection {
		// 定义集合 Schema
		schema := entity.NewSchema().
			// ID 字段：主键
			WithField(entity.NewField().
				WithName(milvusIDField).
				WithDataType(entity.FieldTypeVarChar).
				WithMaxLength(256).
				WithIsPrimaryKey(true)).
			// 内容字段
			WithField(entity.NewField().
				WithName(milvusContentField).
				WithDataType(entity.FieldTypeVarChar).
				WithMaxLength(65535)).
			// 向量字段
			WithField(entity.NewField().
				WithName(milvusVectorField).
				WithDataType(entity.FieldTypeFloatVector).
				WithDim(dimension)).
			// 元数据字段
			WithField(entity.NewField().
				WithName(milvusMetadataField).
				WithDataType(entity.FieldTypeJSON)).
			// 启用动态字段
			WithDynamicFieldEnabled(true)

		// 创建集合
		err = client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(collectionName, schema))
		if err != nil {
			return fmt.Errorf("创建集合失败: %w", err)
		}

		// 创建向量索引
		// HNSW 是一种高效的近似最近邻搜索算法
		// 参数: M=16（连接数），efConstruction=256（构建时的搜索范围）
		idx := index.NewHNSWIndex(entity.COSINE, 16, 256)
		_, err = client.CreateIndex(ctx, milvusclient.NewCreateIndexOption(collectionName, milvusVectorField, idx))
		if err != nil {
			return fmt.Errorf("创建索引失败: %w", err)
		}

		// 加载集合到内存
		loadTask, err := client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collectionName))
		if err != nil {
			return fmt.Errorf("加载集合失败: %w", err)
		}
		// 等待加载完成
		if err := loadTask.Await(ctx); err != nil {
			return fmt.Errorf("等待集合加载失败: %w", err)
		}
	}

	return nil
}
