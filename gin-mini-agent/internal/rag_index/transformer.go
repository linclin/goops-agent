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

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino/components/document"
)

// newDocumentTransformer 创建文档转换器
//
// 文档转换器负责将长文档分割成适合检索的小块。
// 它是 RAG 索引流程的第二个节点。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - document.Transformer: 文档转换器实例
//   - error: 创建过程中的错误
//
// 功能说明:
//   - 使用 Markdown 标题分割器
//   - 根据标题层级（#、##、### 等）分割文档
//   - 保留标题信息作为元数据
//
// 分割策略:
//   - 一级标题（#）: 提取为 "title" 元数据
//   - 其他层级: 按段落自然分割
//
// 使用示例:
//
//	transformer, err := newDocumentTransformer(ctx)
//	docs, err := transformer.Transform(ctx, []*schema.Document{...})
func newDocumentTransformer(ctx context.Context) (tfr document.Transformer, err error) {
	// 创建 Markdown 标题分割器配置
	config := &markdown.HeaderConfig{
		// Headers 定义标题层级与元数据字段的映射
		// 键: Markdown 标题符号（#、##、### 等）
		// 值: 对应的元数据字段名
		Headers: map[string]string{
			"#": "title", // 一级标题映射为 "title" 字段
		},
		// TrimHeaders 是否从内容中移除标题
		// false: 保留标题在内容中
		// true: 从内容中移除标题，只保留在元数据中
		TrimHeaders: false,
	}

	// 创建 Markdown 标题分割器实例
	tfr, err = markdown.NewHeaderSplitter(ctx, config)
	if err != nil {
		return nil, err
	}
	return tfr, nil
}
