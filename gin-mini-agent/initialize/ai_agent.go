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

// 初始化 AI Agent
func InitAiAgent() {
	ctx := context.Background()
	agent, err := ai_agent.BuildAiAgent(ctx)
	if err != nil {
		panic("初始化 AI Agent 失败: " + err.Error())
	}
	ai_agent.GlobalAgent = agent
	global.Log.Info("初始化 AI Agent 完成")
}
