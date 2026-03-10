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

package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/tool/browseruse"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type BrowserUseToolImpl struct {
	config *browseruse.Config
}

// NewBrowserUseTool 创建浏览器自动化工具
// 参照官方示例：https://github.com/cloudwego/eino-ext/tree/main/components/tool/browseruse
func NewBrowserUseTool(ctx context.Context, config *browseruse.Config) (tool.BaseTool, error) {
	if config == nil {
		config = &browseruse.Config{}
	}
	t := &BrowserUseToolImpl{config: config}
	return t.ToEinoTool()
}

// ToEinoTool 实现 tool.BaseTool 接口
func (b *BrowserUseToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	return utils.InferTool("browser_use", "浏览器自动化工具，访问网页、提取内容、执行网页操作", b.Invoke)
}

// Invoke 调用浏览器自动化工具
func (b *BrowserUseToolImpl) Invoke(ctx context.Context, req browseruse.Param) (browseruse.ToolResult, error) {
	but, err := browseruse.NewBrowserUseTool(ctx, &browseruse.Config{})
	if err != nil {
		return browseruse.ToolResult{}, err
	}
	result, err := but.Execute(&req)
	if err != nil {
		return browseruse.ToolResult{}, err
	}
	but.Cleanup()
	return *result, nil
}
