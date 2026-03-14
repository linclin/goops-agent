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

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
)

// newLoader 创建文件加载器
//
// 文件加载器负责从文件系统加载文档内容。
// 它是 RAG 索引流程的第一个节点。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - document.Loader: 文档加载器实例
//   - error: 创建过程中的错误
//
// 功能说明:
//   - 支持加载各种格式的文件
//   - 自动检测文件编码
//   - 提取文件内容和元数据
//
// 使用示例:
//
//	loader, err := newLoader(ctx)
//	docs, err := loader.Load(ctx, document.Source{URI: "/path/to/file.md"})
func newLoader(ctx context.Context) (ldr document.Loader, err error) {
	// 创建文件加载器配置
	// 当前使用默认配置，可根据需要添加自定义配置
	config := &file.FileLoaderConfig{}

	// 创建文件加载器实例
	ldr, err = file.NewFileLoader(ctx, config)
	if err != nil {
		return nil, err
	}
	return ldr, nil
}
