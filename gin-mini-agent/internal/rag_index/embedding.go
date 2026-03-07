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

package rag_index

import (
	"context"
	"fmt"
	"time"

	"gin-mini-agent/pkg/global"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/samber/lo"
)

func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	config := &openai.EmbeddingConfig{
		BaseURL: global.Conf.AiModel.EmbeddingModel.BaseURL,
		APIKey:  global.Conf.AiModel.EmbeddingModel.APIKey,
		Model:   global.Conf.AiModel.EmbeddingModel.Model,
	}
	eb, err = openai.NewEmbedder(ctx, config)
	if err != nil {
		return nil, err
	}
	return eb, nil
}

type retryEmbedding struct {
	embedder embedding.Embedder
	maxRetry int
	interval time.Duration
}

func newRetryEmbedding(embedder embedding.Embedder, maxRetry int, interval time.Duration) embedding.Embedder {
	return &retryEmbedding{
		embedder: embedder,
		maxRetry: maxRetry,
		interval: interval,
	}
}

func (r *retryEmbedding) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	var result [][]float64
	var lastErr error

	_, _, err := lo.AttemptWithDelay(r.maxRetry, r.interval, func(index int, duration time.Duration) error {
		if index > 0 {
			fmt.Printf("[重试] 第 %d 次重试，等待 %v 后重试...\n", index, duration)
		}

		res, err := r.embedder.EmbedStrings(ctx, texts, opts...)
		if err != nil {
			lastErr = err
			return err
		}

		result = res
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("重试 %d 次后仍然失败: %w", r.maxRetry, lastErr)
	}

	return result, nil
}
