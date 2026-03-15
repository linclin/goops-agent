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

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"

	"gin-mini-agent/internal/ai_agent/tools"
	"gin-mini-agent/pkg/global"
)

// GetTools 获取所有可用工具列表
//
// 该函数返回 Agent 可以调用的工具列表。
// 工具是 Agent 与外部世界交互的能力扩展，
// 允许 Agent 执行搜索、文件操作、浏览器自动化等任务。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - skillMiddleware: 可选的 Skill 中间件，用于提取 skill 工具
//
// 返回:
//   - []tool.BaseTool: 工具列表
//   - error: 获取过程中的错误
//
// 当前可用工具:
//   - open: 打开文件或 URL，读取内容
//   - fileeditor: 文件编辑工具，支持读写文件
//   - browseruse: 浏览器自动化工具，支持网页交互
//   - skill: 技能加载工具（如果提供了 skillMiddleware）
//
// 工具调用流程:
//  1. Agent 分析用户问题，决定是否需要调用工具
//  2. Agent 选择合适的工具并生成调用参数
//  3. 工具执行并返回结果
//  4. Agent 根据结果继续推理或生成最终答案
//
// 使用示例:
//
//	tools, err := GetTools(ctx, skillMiddleware)
//	for _, t := range tools {
//	    info, _ := t.Info(ctx)
//	    fmt.Println(info.Name, info.Desc)
//	}
func GetTools(ctx context.Context, skillMiddleware adk.AgentMiddleware) ([]tool.BaseTool, error) {
	// 创建打开文件工具（自定义实现）
	// 功能: 打开本地文件或 URL，读取内容
	// 用途: 当用户需要查看文件内容或访问网页时使用
	toolOpen, err := tools.NewOpenFileTool(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 创建文件编辑器工具（自定义实现）
	// 功能: 创建、读取、修改文件
	// 用途: 当用户需要编辑代码或文本文件时使用
	toolFileEditor, err := tools.NewFileEditorTool(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 创建 Python 执行器工具（自定义实现）
	// 功能: 执行 Python 代码
	// 用途: 当用户需要执行 Python 脚本时使用
	toolPyExecutor, err := tools.NewPyExecutorTool(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 创建 HTTP GET 请求工具（官方库实现）
	// 功能: 发送 HTTP GET 请求
	// 用途: 当用户需要获取网页内容或 API 数据时使用
	toolHTTPGet, err := tools.NewHTTPGetTool(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 创建 HTTP POST 请求工具（官方库实现）
	// 功能: 发送 HTTP POST 请求
	// 用途: 当用户需要提交数据或创建资源时使用
	toolHTTPPost, err := tools.NewHTTPPostTool(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 创建 HTTP PUT 请求工具（官方库实现）
	// 功能: 发送 HTTP PUT 请求
	// 用途: 当用户需要更新资源时使用
	toolHTTPPut, err := tools.NewHTTPPutTool(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 创建 HTTP DELETE 请求工具（官方库实现）
	// 功能: 发送 HTTP DELETE 请求
	// 用途: 当用户需要删除资源时使用
	toolHTTPDelete, err := tools.NewHTTPDeleteTool(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 创建浏览器自动化工具（自定义实现）
	// 功能: 控制浏览器进行网页交互
	// 用途: 当用户需要访问网页、填写表单、截图时使用
	toolBrowserUse, err := tools.NewBrowserUseTool(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 创建命令执行工具（自定义实现）
	// 功能: 在终端中执行命令
	// 用途: 当用户需要执行系统命令、脚本时使用
	toolCommand, err := tools.NewCommandTool(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 创建 kubectl 工具（基于 client-go SDK 实现）
	// 功能: 操作 Kubernetes 集群，支持 get、describe、create、delete、apply 等操作
	// 用途: 当用户需要管理 Kubernetes 集群时使用
	toolKubectl, err := tools.NewKubectlTool(ctx, nil)
	if err != nil {
		global.Log.Warn("创建 kubectl 工具失败，可能是因为未配置 kubeconfig", "error", err)
		// 继续执行，不影响其他工具
	}

	// 构建工具列表
	toolList := []tool.BaseTool{
		toolOpen,
		toolFileEditor,
		toolPyExecutor,
		toolHTTPGet,
		toolHTTPPost,
		toolHTTPPut,
		toolHTTPDelete,
		toolBrowserUse,
		toolCommand,
	}

	// 添加 kubectl 工具（如果创建成功）
	if toolKubectl != nil {
		toolList = append(toolList, toolKubectl)
		global.Log.Info("kubectl 工具加载成功")
	}

	// 加载 MCP 工具
	// MCP (Model Context Protocol) 是由 Anthropic 推出的标准协议
	// 用于 LLM 应用和外部数据源或工具之间通信
	global.Log.Info("开始加载 MCP 工具...")
	mcpTools := loadMCPTools(ctx)
	if len(mcpTools) > 0 {
		global.Log.Info("MCP 工具加载成功", "count", len(mcpTools))
		toolList = append(toolList, mcpTools...)
	}

	// 如果提供了 Skill 中间件，提取并添加额外的工具
	// Skill 中间件会提供 skill 工具，用于加载和执行预定义的技能
	if len(skillMiddleware.AdditionalTools) > 0 {
		global.Log.Info("从 Skill Middleware 加载工具", "count", len(skillMiddleware.AdditionalTools))
		for i, t := range skillMiddleware.AdditionalTools {
			info, err := t.Info(ctx)
			if err != nil {
				global.Log.Warn("获取工具信息失败", "index", i, "error", err)
				continue
			}
			global.Log.Info("加载 Skill 工具", "name", info.Name, "desc", info.Desc)
		}
		toolList = append(toolList, skillMiddleware.AdditionalTools...)
	}

	// 返回工具列表
	// Agent 会根据用户问题自动选择合适的工具
	return toolList, nil
}

// loadMCPTools 加载 MCP 工具
//
// 该函数尝试加载各种 MCP 服务器提供的工具。
// MCP 工具需要外部依赖（如 npx、uvx），如果环境不满足会跳过。
//
// 当前加载的 MCP 工具:
//   - filesystem: 文件系统操作（需要 npx）
//   - fetch: 网页内容获取（需要 uvx）
//   - memory: 记忆存储（需要 npx）
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []tool.BaseTool: 成功加载的 MCP 工具列表
func loadMCPTools(ctx context.Context) []tool.BaseTool {
	var mcpTools []tool.BaseTool

	// 加载文件系统 MCP 工具
	// 允许访问项目目录下的文件
	global.Log.Info("尝试加载 MCP 文件系统工具...")
	fsTools, fsClient, err := tools.NewFileSystemMCPTools(ctx, ".")
	if err != nil {
		global.Log.Warn("加载 MCP 文件系统工具失败", "error", err)
	} else {
		global.Log.Info("MCP 文件系统工具加载成功", "count", len(fsTools))
		mcpTools = append(mcpTools, fsTools...)
		// 注意: MCP 客户端需要保持运行，这里不关闭
		_ = fsClient
	}

	// 加载 Fetch MCP 工具
	// 用于获取网页内容
	global.Log.Info("尝试加载 MCP Fetch 工具...")
	fetchTools, fetchClient, err := tools.NewFetchMCPTools(ctx)
	if err != nil {
		global.Log.Warn("加载 MCP Fetch 工具失败", "error", err)
	} else {
		global.Log.Info("MCP Fetch 工具加载成功", "count", len(fetchTools))
		mcpTools = append(mcpTools, fetchTools...)
		_ = fetchClient
	}

	// 加载 Memory MCP 工具
	// 用于存储和检索记忆
	global.Log.Info("尝试加载 MCP Memory 工具...")
	memoryTools, memoryClient, err := tools.NewMemoryMCPTools(ctx)
	if err != nil {
		global.Log.Warn("加载 MCP Memory 工具失败", "error", err)
	} else {
		global.Log.Info("MCP Memory 工具加载成功", "count", len(memoryTools))
		mcpTools = append(mcpTools, memoryTools...)
		_ = memoryClient
	}

	return mcpTools
}
