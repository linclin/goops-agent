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

// Package rag_index 提供 RAG（检索增强生成）索引功能
//
// 该包负责将文档导入知识库，建立向量索引。
// 主要功能包括：
//   - 文档加载：从文件系统加载 Markdown 文档
//   - 文档分割：将长文档分割成适合检索的小块
//   - 向量索引：将文档向量化并存储到向量数据库
//
// 索引流程:
//
//	文件 -> FileLoader -> MarkdownSplitter -> Indexer -> 向量数据库
//
// 支持的向量数据库:
//   - Chromem: 本地文件存储
//   - Redis: 分布式存储
//   - Milvus: 分布式向量数据库
package rag_index

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"

	"gin-mini-agent/pkg/global"
)

// RagIndex 执行 RAG 索引
//
// 该函数从默认目录（./rag_docs/）加载 Markdown 文档，
// 并使用配置文件中指定的向量数据库建立索引。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - error: 索引过程中的错误
//
// 使用示例:
//
//	err := RagIndex(ctx)
func RagIndex(ctx context.Context) error {
	return RagIndexWithType(ctx, "./rag_docs/", global.Conf.RAG.Type)
}

// RagIndexWithType 执行 RAG 索引（指定目录和数据库类型）
//
// 该函数遍历指定目录下的所有 Markdown 文件，
// 对每个文件执行加载、分割和索引操作。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - dir: 文档目录路径
//   - dbType: 向量数据库类型（chromem/redis/milvus）
//
// 返回:
//   - error: 索引过程中的错误
//
// 处理流程:
//  1. 构建索引图（Graph）
//  2. 遍历目录下的所有 .md 文件
//  3. 对每个文件执行索引操作
//  4. 输出索引结果
//
// Graph 结构:
//   - FileLoader: 加载文件内容
//   - MarkdownSplitter: 分割 Markdown 文档
//   - Indexer: 向量化并存储到数据库
//
// 注意事项:
//   - 只处理 .md 后缀的文件
//   - 每个文件处理后会延迟 1 秒，避免 API 限流
func RagIndexWithType(ctx context.Context, dir string, dbType string) error {
	// 定义节点名称常量
	const (
		// FileLoader 文件加载器节点
		FileLoader = "FileLoader"
		// MarkdownSplitter Markdown 分割器节点
		MarkdownSplitter = "MarkdownSplitter"
		// Indexer 索引器节点
		Indexer = "Indexer"
	)

	// 创建索引图
	// 输入: document.Source（文档源）
	// 输出: []string（文档 ID 列表）
	g := compose.NewGraph[document.Source, []string]()

	// 添加文件加载器节点
	fileLoaderKeyOfLoader, err := newLoader(ctx)
	if err != nil {
		return fmt.Errorf("构建加载器失败: %w", err)
	}
	_ = g.AddLoaderNode(FileLoader, fileLoaderKeyOfLoader)

	// 添加文档转换器节点（Markdown 分割器）
	markdownSplitterKeyOfDocumentTransformer, err := newDocumentTransformer(ctx)
	if err != nil {
		return fmt.Errorf("构建转换器失败: %w", err)
	}
	_ = g.AddDocumentTransformerNode(MarkdownSplitter, markdownSplitterKeyOfDocumentTransformer)

	// 添加索引器节点
	indexerKeyOfIndexer, err := newIndexer(ctx, dbType)
	if err != nil {
		return fmt.Errorf("构建索引器失败: %w", err)
	}
	_ = g.AddIndexerNode(Indexer, indexerKeyOfIndexer)

	// 构建图的边
	// START -> FileLoader -> MarkdownSplitter -> Indexer -> END
	_ = g.AddEdge(compose.START, FileLoader)
	_ = g.AddEdge(FileLoader, MarkdownSplitter)
	_ = g.AddEdge(MarkdownSplitter, Indexer)
	_ = g.AddEdge(Indexer, compose.END)

	// 编译图
	// WithNodeTriggerMode(compose.AnyPredecessor): 任一前驱节点完成即可触发
	runner, err := g.Compile(ctx, compose.WithGraphName("KnowledgeIndexing"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return fmt.Errorf("编译图谱失败: %w", err)
	}

	// 遍历目录，处理所有 Markdown 文件
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("遍历目录失败: %w", err)
		}

		// 跳过目录
		if d.IsDir() {
			return nil
		}

		// 只处理 Markdown 文件
		if !strings.HasSuffix(path, ".md") {
			fmt.Printf("[跳过] 非 Markdown 文件: %s\n", path)
			return nil
		}

		fmt.Printf("[开始] 索引文件: %s\n", path)

		// 执行索引操作
		ids, err := runner.Invoke(ctx, document.Source{URI: path})
		if err != nil {
			return fmt.Errorf("调用索引图谱失败: %w", err)
		}

		fmt.Printf("[完成] 索引文件: %s, 分段数量: %d\n", path, len(ids))

		// 添加延迟，避免 API 限流
		time.Sleep(1 * time.Second)
		return nil
	})

	return err
}
