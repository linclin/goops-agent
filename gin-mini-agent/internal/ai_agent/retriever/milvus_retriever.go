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

// NewMilvusRetriever 创建 Milvus 检索器
//
// Milvus 是一个高性能的分布式向量数据库，专为大规模向量检索设计。
// 适合大规模生产环境部署，支持分布式架构和高可用。
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
//   - milvus.addr: Milvus 服务器地址，如 "localhost:19530"
//   - milvus.username: 用户名（可选）
//   - milvus.password: 密码（可选）
//   - milvus.collection: 集合名称
//
// 前置条件:
//   - Milvus 服务器需要正常运行
//   - 需要先创建集合并建立向量索引
//
// 使用示例:
//
//	retriever, err := NewMilvusRetriever(ctx, embedder)
//	docs, err := retriever.Retrieve(ctx, "什么是机器学习？")
func NewMilvusRetriever(ctx context.Context, embedder embedding.Embedder) (rtr retriever.Retriever, err error) {
	// 创建 Milvus 检索器配置
	config := &milvus2.RetrieverConfig{
		// Milvus 客户端配置
		ClientConfig: &milvusclient.ClientConfig{
			Address:  global.Conf.RAG.Milvus.Addr,
			Username: global.Conf.RAG.Milvus.Username,
			Password: global.Conf.RAG.Milvus.Password,
		},
		// 集合名称
		Collection: global.Conf.RAG.Milvus.Collection,
		// 默认返回 8 个结果
		TopK:       8,
		// 嵌入模型
		Embedding:  embedder,
		// 文档转换函数：将 Milvus 结果转换为 schema.Document
		DocumentConverter: func(ctx context.Context, result milvusclient.ResultSet) ([]*schema.Document, error) {
			// 预分配结果切片
			docs := make([]*schema.Document, 0, result.ResultCount)

			// 遍历结果集
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

				// 提取相似度分数
				if i < len(result.Scores) {
					doc.WithScore(float64(result.Scores[i]))
				}

				docs = append(docs, doc)
			}

			return docs, nil
		},
	}

	// 创建 Milvus 检索器
	rtr, err = milvus2.NewRetriever(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建 milvus 检索器失败: %w", err)
	}

	return rtr, nil
}
