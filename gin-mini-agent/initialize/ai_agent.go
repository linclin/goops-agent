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

package initialize

import (
	"context"

	"gin-mini-agent/internal/ai_agent"
	"gin-mini-agent/pkg/global"
)

// InitAiAgent 初始化 AI Agent
//
// 该函数负责初始化全局 AI Agent 实例。
// AI Agent 是应用程序的核心组件，负责处理用户对话请求。
//
// 初始化流程:
//  1. 创建上下文
//  2. 构建 AI Agent（包括 Graph、工具、模型等）
//  3. 设置全局变量
//
// 注意事项:
//   - 如果初始化失败，会触发 panic
//   - 初始化成功后会记录日志
//
// 使用示例:
//
//	initialize.InitAiAgent()
func InitAiAgent() {
	// 创建上下文
	ctx := context.Background()

	// 构建 AI Agent
	// BuildAiAgent 会创建完整的 Graph 工作流
	agent, err := ai_agent.BuildAiAgent(ctx)
	if err != nil {
		panic("初始化 AI Agent 失败: " + err.Error())
	}

	// 设置全局变量
	ai_agent.GlobalAgent = agent

	// 记录初始化完成日志
	global.Log.Info("初始化 AI Agent 完成")
}
