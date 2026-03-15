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
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

// newReactAgent 创建 React Agent 组件
//
// 该函数是 Graph 中 ReactAgent 节点的初始化函数，
// 负责创建和配置 ReAct（Reasoning + Acting）推理代理。
//
// ReAct Agent 工作原理:
//  1. 推理（Reasoning）: 分析用户问题，决定是否需要调用工具
//  2. 行动（Acting）: 调用工具获取信息
//  3. 观察（Observation）: 观察工具返回的结果
//  4. 循环：重复上述步骤，直到可以生成最终答案
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - skillMiddleware: 可选的 Skill 中间件，用于提取 skill 工具
//
// 返回:
//   - lba: Lambda 实例，支持同步调用和流式调用
//   - err: 创建过程中的错误
//
// 配置说明:
//   - MaxStep: 最大推理步数，防止无限循环
//   - ToolCallingModel: 用于工具调用的聊天模型
//   - ToolsConfig: 可用工具列表
//   - ToolReturnDirectly: 直接返回结果的工具列表
//
// 可用工具:
//   - open: 打开文件或 URL
//   - browseruse: 浏览器自动化
//   - fileeditor: 文件编辑
//   - wikipedia: 维基百科搜索
//   - searxng: 搜索引擎
//   - httprequest: HTTP 请求
//   - commandline: 命令行执行
//   - skill: 技能加载（如果提供了 skillMiddleware）
func newReactAgent(ctx context.Context, skillMiddleware adk.AgentMiddleware) (lba *compose.Lambda, err error) {
	// 创建 ReAct Agent 配置
	config := &react.AgentConfig{
		// MaxStep 最大推理步数
		// 限制 Agent 最多执行 25 步推理循环
		// 防止 Agent 陷入无限循环或消耗过多资源
		MaxStep: 25,

		// ToolReturnDirectly 直接返回结果的工具列表
		// 空列表表示所有工具的结果都需要经过推理处理
		// 如果某个工具的结果应该直接返回给用户，可以将其名称添加到列表中
		ToolReturnDirectly: map[string]struct{}{},
	}

	// 创建聊天模型实例
	// 该模型用于：
	// 1. 工具调用决策：决定是否需要调用工具
	// 2. 工具参数生成：生成工具调用所需的参数
	// 3. 结果推理：分析工具返回的结果
	// 4. 最终答案生成：生成最终的回复
	chatModelIns, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	config.ToolCallingModel = chatModelIns

	// 获取可用工具列表
	// 工具定义在 internal/ai_agent/tools 目录下
	// 每个工具都实现了 tool.InvokableTool 接口
	tools, err := GetTools(ctx, skillMiddleware)
	if err != nil {
		return nil, err
	}
	config.ToolsConfig.Tools = tools

	// 创建 ReAct Agent 实例
	// 该实例实现了 Generate（同步）和 Stream（流式）两个方法
	ins, err := react.NewAgent(ctx, config)
	if err != nil {
		return nil, err
	}

	// 将 Agent 转换为 Lambda
	// AnyLambda 创建一个支持同步和流式调用的 Lambda
	// 参数说明:
	//   - ins.Generate: 同步调用方法
	//   - ins.Stream: 流式调用方法
	//   - nil: 不需要转换输入
	//   - nil: 不需要转换输出
	lba, err = compose.AnyLambda(ins.Generate, ins.Stream, nil, nil)
	if err != nil {
		return nil, err
	}
	return lba, nil
}
