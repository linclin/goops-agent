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
	"time"

	eino_mcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPToolConfig MCP 工具配置
//
// MCP (Model Context Protocol) 是由 Anthropic 推出的用于 LLM 应用
// 和外部数据源或工具之间通信的标准协议。
type MCPToolConfig struct {
	// Command MCP 服务器启动命令
	// 例如: "npx", "uvx", "python"
	Command string

	// Args MCP 服务器启动参数
	// 例如: ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/dir"]
	Args []string

	// Env 环境变量
	Env []string

	// ToolNames 要加载的工具名称列表
	// 如果为空，则加载所有可用工具
	ToolNames []string

	// InitTimeout 初始化超时时间（秒）
	// 默认: 30 秒
	InitTimeout int

	// ClientInfo 客户端信息
	ClientInfo *MCPClientInfo
}

// MCPClientInfo MCP 客户端信息
type MCPClientInfo struct {
	// Name 客户端名称
	Name string
	// Version 客户端版本
	Version string
}

// defaultMCPToolConfig 创建默认配置
func defaultMCPToolConfig() *MCPToolConfig {
	return &MCPToolConfig{
		InitTimeout: 30,
		ClientInfo: &MCPClientInfo{
			Name:    "gin-mini-agent",
			Version: "1.0.0",
		},
	}
}

// MCPClientWrapper MCP 客户端包装器
//
// 该结构体封装了 MCP 客户端和相关配置。
type MCPClientWrapper struct {
	client *client.Client
	config *MCPToolConfig
}

// NewMCPClient 创建 MCP 客户端
//
// 该函数创建一个连接到 MCP 服务器的客户端。
// MCP 服务器可以是本地进程（通过 stdio 通信）或远程服务。
//
// 参数:
//   - ctx: 上下文
//   - config: MCP 工具配置
//
// 返回:
//   - *MCPClientWrapper: MCP 客户端包装器
//   - error: 创建错误
//
// 使用示例:
//
//	config := &MCPToolConfig{
//	    Command: "npx",
//	    Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/path/to/dir"},
//	}
//	mcpClient, err := NewMCPClient(ctx, config)
func NewMCPClient(ctx context.Context, config *MCPToolConfig) (*MCPClientWrapper, error) {
	slog.InfoContext(ctx, "[mcp] 创建 MCP 客户端", "command", config.Command, "args", config.Args)

	if config == nil {
		config = defaultMCPToolConfig()
	}

	if config.Command == "" {
		slog.ErrorContext(ctx, "[mcp] MCP 服务器命令不能为空")
		return nil, fmt.Errorf("MCP 服务器命令不能为空")
	}

	if config.InitTimeout <= 0 {
		config.InitTimeout = 30
	}

	if config.ClientInfo == nil {
		config.ClientInfo = &MCPClientInfo{
			Name:    "gin-mini-agent",
			Version: "1.0.0",
		}
	}

	mcpClient, err := client.NewStdioMCPClient(config.Command, config.Env, config.Args...)
	if err != nil {
		slog.ErrorContext(ctx, "[mcp] 创建 MCP 客户端失败", "error", err)
		return nil, fmt.Errorf("创建 MCP 客户端失败: %w", err)
	}

	initCtx, cancel := context.WithTimeout(ctx, time.Duration(config.InitTimeout)*time.Second)
	defer cancel()

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    config.ClientInfo.Name,
		Version: config.ClientInfo.Version,
	}

	_, err = mcpClient.Initialize(initCtx, initRequest)
	if err != nil {
		slog.ErrorContext(ctx, "[mcp] 初始化 MCP 客户端失败", "error", err)
		return nil, fmt.Errorf("初始化 MCP 客户端失败: %w", err)
	}

	slog.InfoContext(ctx, "[mcp] MCP 客户端创建成功", "command", config.Command)
	return &MCPClientWrapper{
		client: mcpClient,
		config: config,
	}, nil
}

// GetTools 从 MCP 服务器获取工具列表
//
// 该函数将 MCP 服务器提供的工具转换为 Eino 工具格式。
//
// 参数:
//   - ctx: 上下文
//   - wrapper: MCP 客户端包装器
//   - toolNames: 要获取的工具名称列表，为空则获取所有
//
// 返回:
//   - []tool.BaseTool: Eino 工具列表
//   - error: 获取错误
func GetTools(ctx context.Context, wrapper *MCPClientWrapper, toolNames []string) ([]tool.BaseTool, error) {
	slog.DebugContext(ctx, "[mcp] 获取工具列表", "toolNames", toolNames)

	if wrapper == nil || wrapper.client == nil {
		slog.ErrorContext(ctx, "[mcp] MCP 客户端未初始化")
		return nil, fmt.Errorf("MCP 客户端未初始化")
	}

	config := &eino_mcp.Config{
		Cli: wrapper.client,
	}

	if len(toolNames) > 0 {
		config.ToolNameList = toolNames
	}

	tools, err := eino_mcp.GetTools(ctx, config)
	if err != nil {
		slog.ErrorContext(ctx, "[mcp] 获取 MCP 工具失败", "error", err)
		return nil, fmt.Errorf("获取 MCP 工具失败: %w", err)
	}

	slog.InfoContext(ctx, "[mcp] 获取工具列表成功", "count", len(tools))
	return tools, nil
}

// NewMCPTools 创建 MCP 工具集
//
// 该函数是一个便捷方法，一次性完成 MCP 客户端创建和工具获取。
//
// 参数:
//   - ctx: 上下文
//   - config: MCP 工具配置
//
// 返回:
//   - []tool.BaseTool: Eino 工具列表
//   - *MCPClientWrapper: MCP 客户端包装器（用于后续管理）
//   - error: 创建错误
//
// 使用示例:
//
//	config := &MCPToolConfig{
//	    Command: "npx",
//	    Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
//	    ToolNames: []string{"read_file", "write_file"},
//	}
//	tools, client, err := NewMCPTools(ctx, config)
func NewMCPTools(ctx context.Context, config *MCPToolConfig) ([]tool.BaseTool, *MCPClientWrapper, error) {
	wrapper, err := NewMCPClient(ctx, config)
	if err != nil {
		return nil, nil, err
	}

	tools, err := GetTools(ctx, wrapper, config.ToolNames)
	if err != nil {
		return nil, wrapper, err
	}

	return tools, wrapper, nil
}

// Close 关闭 MCP 客户端
//
// 该方法关闭 MCP 客户端连接，释放资源。
func (w *MCPClientWrapper) Close() error {
	if w.client != nil {
		return w.client.Close()
	}
	return nil
}

// ListAvailableTools 列出 MCP 服务器上所有可用的工具
//
// 该函数获取 MCP 服务器提供的所有工具信息。
//
// 参数:
//   - ctx: 上下文
//   - wrapper: MCP 客户端包装器
//
// 返回:
//   - []ToolInfo: 工具信息列表
//   - error: 获取错误
func ListAvailableTools(ctx context.Context, wrapper *MCPClientWrapper) ([]ToolInfo, error) {
	if wrapper == nil || wrapper.client == nil {
		return nil, fmt.Errorf("MCP 客户端未初始化")
	}

	result, err := wrapper.client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("列出 MCP 工具失败: %w", err)
	}

	var tools []ToolInfo
	for _, t := range result.Tools {
		tools = append(tools, ToolInfo{
			Name:        t.Name,
			Description: t.Description,
		})
	}

	return tools, nil
}

// ToolInfo 工具信息
type ToolInfo struct {
	// Name 工具名称
	Name string
	// Description 工具描述
	Description string
}

// NewFileSystemMCPTools 创建文件系统 MCP 工具
//
// 该函数创建一个连接到文件系统 MCP 服务器的工具集。
// 文件系统 MCP 服务器允许 AI 读写指定目录下的文件。
//
// 参数:
//   - ctx: 上下文
//   - allowedDir: 允许访问的目录路径
//
// 返回:
//   - []tool.BaseTool: Eino 工具列表
//   - *MCPClientWrapper: MCP 客户端包装器
//   - error: 创建错误
//
// 使用示例:
//
//	tools, client, err := NewFileSystemMCPTools(ctx, "/tmp")
func NewFileSystemMCPTools(ctx context.Context, allowedDir string) ([]tool.BaseTool, *MCPClientWrapper, error) {
	config := &MCPToolConfig{
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-filesystem", allowedDir},
		InitTimeout: 30,
	}
	return NewMCPTools(ctx, config)
}

// NewAirbnbMCPTools 创建 Airbnb MCP 工具
//
// 该函数创建一个连接到 Airbnb MCP 服务器的工具集。
// Airbnb MCP 服务器允许 AI 搜索 Airbnb 上的住宿信息。
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []tool.BaseTool: Eino 工具列表
//   - *MCPClientWrapper: MCP 客户端包装器
//   - error: 创建错误
//
// 使用示例:
//
//	tools, client, err := NewAirbnbMCPTools(ctx)
func NewAirbnbMCPTools(ctx context.Context) ([]tool.BaseTool, *MCPClientWrapper, error) {
	config := &MCPToolConfig{
		Command:     "npx",
		Args:        []string{"-y", "@openbnb/mcp-server-airbnb", "--ignore-robots-txt"},
		InitTimeout: 30,
	}
	return NewMCPTools(ctx, config)
}

// NewGitHubMCPTools 创建 GitHub MCP 工具
//
// 该函数创建一个连接到 GitHub MCP 服务器的工具集。
// GitHub MCP 服务器允许 AI 操作 GitHub 仓库。
//
// 参数:
//   - ctx: 上下文
//   - token: GitHub 个人访问令牌
//
// 返回:
//   - []tool.BaseTool: Eino 工具列表
//   - *MCPClientWrapper: MCP 客户端包装器
//   - error: 创建错误
func NewGitHubMCPTools(ctx context.Context, token string) ([]tool.BaseTool, *MCPClientWrapper, error) {
	config := &MCPToolConfig{
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-github"},
		InitTimeout: 30,
		Env:         []string{fmt.Sprintf("GITHUB_PERSONAL_ACCESS_TOKEN=%s", token)},
	}
	return NewMCPTools(ctx, config)
}

// NewFetchMCPTools 创建 Fetch MCP 工具
//
// 该函数创建一个连接到 Fetch MCP 服务器的工具集。
// Fetch MCP 服务器允许 AI 获取网页内容。
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []tool.BaseTool: Eino 工具列表
//   - *MCPClientWrapper: MCP 客户端包装器
//   - error: 创建错误
func NewFetchMCPTools(ctx context.Context) ([]tool.BaseTool, *MCPClientWrapper, error) {
	config := &MCPToolConfig{
		Command:     "uvx",
		Args:        []string{"mcp-server-fetch"},
		InitTimeout: 30,
	}
	return NewMCPTools(ctx, config)
}

// NewMemoryMCPTools 创建 Memory MCP 工具
//
// 该函数创建一个连接到 Memory MCP 服务器的工具集。
// Memory MCP 服务器允许 AI 存储和检索记忆信息。
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []tool.BaseTool: Eino 工具列表
//   - *MCPClientWrapper: MCP 客户端包装器
//   - error: 创建错误
func NewMemoryMCPTools(ctx context.Context) ([]tool.BaseTool, *MCPClientWrapper, error) {
	config := &MCPToolConfig{
		Command:     "npx",
		Args:        []string{"-y", "@modelcontextprotocol/server-memory"},
		InitTimeout: 30,
	}
	return NewMCPTools(ctx, config)
}
