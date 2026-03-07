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
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"

	"gin-mini-agent/pkg/global"
)

func RagIndex(ctx context.Context) error {
	return RagIndexWithType(ctx, "./rag_docs/", global.Conf.RAG.Type)
}

func RagIndexWithType(ctx context.Context, dir string, dbType string) error {
	const (
		FileLoader       = "FileLoader"
		MarkdownSplitter = "MarkdownSplitter"
		Indexer          = "Indexer"
	)
	g := compose.NewGraph[document.Source, []string]()
	fileLoaderKeyOfLoader, err := newLoader(ctx)
	if err != nil {
		return fmt.Errorf("构建加载器失败: %w", err)
	}
	_ = g.AddLoaderNode(FileLoader, fileLoaderKeyOfLoader)
	markdownSplitterKeyOfDocumentTransformer, err := newDocumentTransformer(ctx)
	if err != nil {
		return fmt.Errorf("构建转换器失败: %w", err)
	}
	_ = g.AddDocumentTransformerNode(MarkdownSplitter, markdownSplitterKeyOfDocumentTransformer)
	indexerKeyOfIndexer, err := newIndexer(ctx, dbType)
	if err != nil {
		return fmt.Errorf("构建索引器失败: %w", err)
	}
	_ = g.AddIndexerNode(Indexer, indexerKeyOfIndexer)
	_ = g.AddEdge(compose.START, FileLoader)
	_ = g.AddEdge(FileLoader, MarkdownSplitter)
	_ = g.AddEdge(MarkdownSplitter, Indexer)
	_ = g.AddEdge(Indexer, compose.END)

	runner, err := g.Compile(ctx, compose.WithGraphName("KnowledgeIndexing"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return fmt.Errorf("编译图谱失败: %w", err)
	}

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("遍历目录失败: %w", err)
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			fmt.Printf("[跳过] 非 Markdown 文件: %s\n", path)
			return nil
		}

		fmt.Printf("[开始] 索引文件: %s\n", path)

		ids, err := runner.Invoke(ctx, document.Source{URI: path})
		if err != nil {
			return fmt.Errorf("调用索引图谱失败: %w", err)
		}

		fmt.Printf("[完成] 索引文件: %s, 分段数量: %d\n", path, len(ids))
		// 在处理每个文件后添加延迟
		time.Sleep(1 * time.Second)
		return nil
	})

	return err
}
