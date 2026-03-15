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

// Package tools 提供 AI Agent 可调用的工具集合
//
// 该包定义了多种工具，扩展 AI Agent 的能力边界。
// 工具是 Agent 与外部世界交互的桥梁，允许 Agent 执行文件操作、
// 浏览器自动化、网络请求等任务。
//
// 当前可用工具:
//   - open: 打开文件/目录/网页链接
//   - browseruse: 浏览器自动化工具
//   - fileeditor: 文件编辑器
//
// 工具开发指南:
//  1. 定义工具结构体，包含配置信息
//  2. 实现 ToEinoTool 方法，返回 tool.InvokableTool
//  3. 实现 Invoke 方法，执行具体的工具逻辑
//  4. 使用 utils.InferTool 自动生成工具信息
package tools

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// OpenFileToolImpl 打开文件工具实现
//
// 该工具用于在系统默认应用中打开文件、目录或网页链接。
// 支持跨平台操作（Windows、macOS、Linux）。
//
// 使用场景:
//   - 用户要求打开某个文件查看
//   - 用户要求打开某个网页
//   - 用户要求打开某个目录
//
// 示例:
//   - 打开文件: file:///path/to/file.txt
//   - 打开网页: https://example.com
//   - 打开目录: /path/to/directory
type OpenFileToolImpl struct {
	// config 工具配置
	config *OpenFileToolConfig
}

// OpenFileToolConfig 打开文件工具配置
//
// 当前配置为空，保留用于未来扩展。
type OpenFileToolConfig struct {
}

// defaultOpenFileToolConfig 创建默认配置
//
// 返回默认的工具配置实例。
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - *OpenFileToolConfig: 配置实例
//   - error: 错误信息
func defaultOpenFileToolConfig(ctx context.Context) (*OpenFileToolConfig, error) {
	config := &OpenFileToolConfig{}
	return config, nil
}

// NewOpenFileTool 创建打开文件工具实例
//
// 该函数创建一个用于打开文件/目录/网页的工具。
// 如果未提供配置，将使用默认配置。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - config: 工具配置，可选
//
// 返回:
//   - tool.BaseTool: 工具实例
//   - error: 创建过程中的错误
//
// 使用示例:
//
//	tool, err := NewOpenFileTool(ctx, nil)
//	result, err := tool.Invoke(ctx, OpenReq{URI: "https://example.com"})
func NewOpenFileTool(ctx context.Context, config *OpenFileToolConfig) (tn tool.BaseTool, err error) {
	// 如果配置为空，使用默认配置
	if config == nil {
		config, err = defaultOpenFileToolConfig(ctx)
		if err != nil {
			return nil, err
		}
	}

	// 创建工具实例
	t := &OpenFileToolImpl{config: config}

	// 转换为 Eino 工具
	tn, err = t.ToEinoTool()
	if err != nil {
		return nil, err
	}
	return tn, nil
}

// ToEinoTool 转换为 Eino 工具接口
//
// 该方法将工具实现转换为 Eino 框架的工具接口。
// 使用 utils.InferTool 自动推断工具的参数和返回值类型。
//
// 返回:
//   - tool.InvokableTool: Eino 工具实例
//   - error: 转换错误
func (of *OpenFileToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	// 使用 InferTool 自动生成工具信息
	// 参数:
	//   - name: 工具名称，用于 Agent 识别
	//   - description: 工具描述，帮助 Agent 理解工具用途
	//   - invoke: 工具调用函数
	return utils.InferTool("open", "在系统默认应用中打开文件/目录/网页链接", of.Invoke)
}

// Invoke 执行打开操作
//
// 该方法是工具的核心实现，负责打开指定的 URI。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - req: 打开请求，包含要打开的 URI
//
// 返回:
//   - OpenRes: 打开结果，包含操作消息
//   - error: 执行错误
//
// 支持的 URI 类型:
//   - 文件路径: file:///path/to/file 或直接使用路径
//   - 网页链接: http:// 或 https://
//   - 目录路径: /path/to/directory
//
// 跨平台支持:
//   - Windows: 使用 rundll32 打开
//   - macOS: 使用 open 命令
//   - Linux: 使用 xdg-open 命令
func (of *OpenFileToolImpl) Invoke(ctx context.Context, req OpenReq) (res OpenRes, err error) {
	slog.InfoContext(ctx, "[open] 工具调用开始", "uri", req.URI)

	if req.URI == "" {
		slog.WarnContext(ctx, "[open] URI 为空")
		res.Message = "URI 不能为空"
		return res, nil
	}

	if isFilePath(req.URI) {
		req.URI = strings.TrimPrefix(req.URI, "file:///")
		if _, err := os.Stat(req.URI); err != nil {
			slog.ErrorContext(ctx, "[open] 文件不存在", "uri", req.URI, "error", err)
			res.Message = fmt.Sprintf("文件不存在: %s", req.URI)
			return res, nil
		}
	}

	err = openURI(req.URI)
	if err != nil {
		slog.ErrorContext(ctx, "[open] 打开失败", "uri", req.URI, "error", err)
		res.Message = fmt.Sprintf("打开失败 %s: %s", req.URI, err.Error())
		return res, nil
	}

	slog.InfoContext(ctx, "[open] 工具调用成功", "uri", req.URI)
	res.Message = fmt.Sprintf("成功，已打开 %s", req.URI)
	return res, nil
}

// OpenReq 打开请求结构体
//
// 定义了打开工具的输入参数。
type OpenReq struct {
	// URI 要打开的资源标识符
	// 支持文件路径、目录路径、网页链接
	// 示例:
	//   - 文件: file:///path/to/file.txt 或 /path/to/file.txt
	//   - 目录: /path/to/directory
	//   - 网页: https://example.com
	URI string `json:"uri" jsonschema_description:"要打开的文件/目录/网页链接的 URI"`
}

// OpenRes 打开响应结构体
//
// 定义了打开工具的输出结果。
type OpenRes struct {
	// Message 操作结果消息
	// 包含操作成功或失败的详细信息
	Message string `json:"message" jsonschema_description:"操作消息"`
}

// openURI 使用系统默认应用打开 URI
//
// 该函数根据操作系统选择合适的命令打开 URI。
//
// 参数:
//   - uri: 要打开的 URI
//
// 返回:
//   - error: 打开错误
//
// 支持的操作系统:
//   - Windows: rundll32 url.dll,FileProtocolHandler
//   - macOS: open
//   - Linux: xdg-open
func openURI(uri string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Windows 使用 rundll32 打开
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", uri)
	case "darwin":
		// macOS 使用 open 命令
		cmd = exec.Command("open", uri)
	case "linux":
		// Linux 使用 xdg-open 命令
		cmd = exec.Command("xdg-open", uri)
	default:
		return fmt.Errorf("不支持的平台")
	}
	return cmd.Run()
}

// isFilePath 判断路径是否为文件路径
//
// 该函数检查给定的路径是否使用 file:// 协议。
//
// 参数:
//   - path: 要检查的路径
//
// 返回:
//   - bool: 如果是文件路径返回 true
func isFilePath(path string) bool {
	s, err := url.Parse(path)
	// 检查协议是否为 file 且路径不为空
	return err == nil && s.Scheme == "file" && s.Path != ""
}
