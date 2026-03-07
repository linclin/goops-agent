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
	"fmt"

	"github.com/cloudwego/eino/components/retriever"

	ragretriever "gin-mini-agent/internal/ai_agent/retriever"
	"gin-mini-agent/pkg/global"
)

func newRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	embedder, err := newEmbedding(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建嵌入模型失败: %w", err)
	}

	dbType := global.Conf.RAG.Type
	if dbType == "" {
		dbType = "redis"
	}

	switch dbType {
	case "chromem":
		return ragretriever.NewChromemRetriever(ctx, embedder)
	case "milvus":
		return ragretriever.NewMilvusRetriever(ctx, embedder)
	case "redis":
		return ragretriever.NewRedisRetriever(ctx, embedder)
	default:
		return ragretriever.NewRedisRetriever(ctx, embedder)
	}
}
