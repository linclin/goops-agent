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
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/cloudwego/eino-ext/components/tool/browseruse"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// BrowserUseToolImpl 浏览器自动化工具实现
//
// 该工具封装了官方的 browseruse 工具，提供浏览器自动化能力。
// 基于 chromedp 实现，支持无头浏览器操作。
//
// 使用场景:
//   - 访问网页并提取内容
//   - 填写表单
//   - 截图
//   - 执行 JavaScript
//   - 网页数据抓取
//
// 官方文档:
// https://github.com/cloudwego/eino-ext/tree/main/components/tool/browseruse
type BrowserUseToolImpl struct {
	// Param 浏览器工具参数
	Param *browseruse.Param
}

// chromePaths 不同操作系统的 Chrome 可执行文件路径
var chromePaths = map[string][]string{
	"windows": {
		// Windows 常见 Chrome 安装路径
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		// 用户目录下的 Chrome
		os.Getenv("LOCALAPPDATA") + `\Google\Chrome\Application\chrome.exe`,
		// Edge 浏览器（基于 Chromium）
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
	},
	"darwin": {
		// macOS Chrome 路径
		`/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`,
		`/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge`,
	},
	"linux": {
		// Linux Chrome 路径
		`/usr/bin/google-chrome`,
		`/usr/bin/google-chrome-stable`,
		`/usr/bin/chromium`,
		`/usr/bin/chromium-browser`,
	},
}

// findChromePath 查找系统中的 Chrome 可执行文件路径
//
// 该函数根据操作系统自动查找 Chrome 浏览器的安装路径。
// 支持 Windows、macOS 和 Linux 系统。
//
// 返回:
//   - string: Chrome 可执行文件路径，如果未找到则返回空字符串
func findChromePath() string {
	paths, ok := chromePaths[runtime.GOOS]
	if !ok {
		return ""
	}

	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// NewBrowserUseTool 创建浏览器自动化工具
//
// 该函数创建一个浏览器自动化工具实例。
// 使用官方的 browseruse 库实现，支持无头浏览器操作。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - param: 工具参数，可选。如果为 nil，使用默认参数
//
// 返回:
//   - tool.BaseTool: 工具实例
//   - error: 创建过程中的错误
//
// 配置说明:
//   - 自动检测 Chrome 浏览器路径
//   - 默认使用无头模式
//   - 支持自定义浏览器选项
//
// 使用示例:
//
//	tool, err := NewBrowserUseTool(ctx, nil)
//	result, err := tool.Invoke(ctx, browseruse.Param{
//	    URL: "https://example.com",
//	    Action: "extract",
//	})
func NewBrowserUseTool(ctx context.Context, Param *browseruse.Param) (tool.BaseTool, error) {
	if Param == nil {
		Param = &browseruse.Param{}
	}

	t := &BrowserUseToolImpl{Param: Param}

	return t.ToEinoTool()
}

// ToEinoTool 实现 tool.BaseTool 接口
//
// 该方法将浏览器工具转换为 Eino 框架的工具接口。
// 使用 utils.InferTool 自动推断工具的参数和返回值类型。
//
// 返回:
//   - tool.InvokableTool: Eino 工具实例
//   - error: 转换错误
func (b *BrowserUseToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	return utils.InferTool("browser_use", "浏览器自动化工具，访问网页、提取内容、执行网页操作", b.Invoke)
}

// Invoke 调用浏览器自动化工具
//
// 该方法是工具的核心实现，负责执行浏览器操作。
// 每次调用都会创建新的浏览器实例，执行完成后自动清理。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - req: 浏览器操作参数，包含 URL 和操作类型
//
// 返回:
//   - browseruse.ToolResult: 操作结果，包含提取的内容或状态信息
//   - error: 执行错误
//
// 错误处理:
//   - 工具调用失败时返回包含错误信息的 ToolResult，而不是抛出错误
//   - 这样可以让 Agent 了解失败原因并尝试其他方案
//
// 支持的操作:
//   - 访问网页: 打开指定的 URL
//   - 提取内容: 获取网页文本内容
//   - 截图: 对网页进行截图
//   - 执行脚本: 在页面中执行 JavaScript
func (b *BrowserUseToolImpl) Invoke(ctx context.Context, req browseruse.Param) (browseruse.ToolResult, error) {
	slog.InfoContext(ctx, "[browser_use] 工具调用开始", "url", req.URL)

	chromePath := findChromePath()
	if chromePath == "" {
		slog.ErrorContext(ctx, "[browser_use] 未找到 Chrome 浏览器")
		return browseruse.ToolResult{
			Error: "浏览器自动化工具调用失败: 未找到 Chrome 浏览器。请安装 Google Chrome 或 Microsoft Edge 浏览器后重试。",
		}, nil
	}

	slog.DebugContext(ctx, "[browser_use] 找到 Chrome 浏览器", "path", chromePath)

	but, err := browseruse.NewBrowserUseTool(ctx, &browseruse.Config{
		Headless:           true,
		ChromeInstancePath: chromePath,
	})
	if err != nil {
		slog.ErrorContext(ctx, "[browser_use] 初始化失败", "error", err)
		return browseruse.ToolResult{
			Error: fmt.Sprintf("浏览器自动化工具初始化失败: %v。请确保 Chrome 浏览器已正确安装。", err),
		}, nil
	}

	result, err := but.Execute(&req)
	if err != nil {
		slog.ErrorContext(ctx, "[browser_use] 执行失败", "error", err)
		return browseruse.ToolResult{
			Error: fmt.Sprintf("浏览器自动化工具执行失败: %v", err),
		}, nil
	}

	but.Cleanup()

	slog.InfoContext(ctx, "[browser_use] 执行成功", "url", req.URL)
	return *result, nil
}
