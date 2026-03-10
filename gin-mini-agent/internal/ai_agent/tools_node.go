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

	"github.com/cloudwego/eino/components/tool"

	tools "gin-mini-agent/internal/ai_agent/tools"
)

func GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	// 创建打开文件工具（自定义）
	toolOpen, err := NewOpenFileTool(ctx)
	if err != nil {
		return nil, err
	}
	// 创建文件编辑器工具（官方库）
	toolFileEditor, err := tools.NewFileEditorTool(ctx, nil)
	if err != nil {
		return nil, err
	}
	// 创建浏览器自动化工具（官方库）
	toolBrowserUse, err := tools.NewBrowserUseTool(ctx, nil)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{
		toolOpen,
		toolFileEditor,
		toolBrowserUse,
	}, nil
}

func NewOpenFileTool(ctx context.Context) (tn tool.BaseTool, err error) {
	return tools.NewOpenFileTool(ctx, nil)
}
