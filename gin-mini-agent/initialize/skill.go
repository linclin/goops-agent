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
	"path/filepath"

	"gin-mini-agent/pkg/global"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
)

// InitSkillMiddleware 初始化 Skill 后端和中间件
//
// 该函数负责初始化和配置 Skill 后端和中间件，使 Agent 能够动态发现和使用预定义的技能。
// Skill 是包含指令、脚本和资源的文件夹，Agent 可以按需发现和使用这些技能来扩展自身能力。
//
// 功能说明:
//   - 基于本地文件系统加载技能
//   - 支持技能的渐进式展示（Progressive Disclosure）
//   - 允许 Agent 动态发现和激活技能
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//
// 返回:
//   - skill.Backend: Skill 后端实例，用于技能检索
//   - adk.ChatModelAgentMiddleware: Skill 中间件实例，用于提取指令和工具
//   - error: 初始化过程中的错误
//
// 技能目录结构:
//
//	skills/
//	  ├── kubernetes-1.0.1/
//	  │   └── SKILL.md          # 必需：指令 + 元数据
//	  │   ├── scripts/          # 可选：可执行代码
//	  │   └── references/       # 可选：参考文档
//	  ├── kubectl-1.0.0/
//	  │   └── SKILL.md
//	  └── ...
//
// 使用示例:
//
//	skillBackend, skillMiddleware, err := initialize.InitSkillMiddleware(ctx)
func InitSkillMiddleware(ctx context.Context) (skill.Backend, adk.AgentMiddleware, error) {
	// 获取项目根目录
	// 从全局配置中获取或直接使用相对路径
	baseDir := "./skills"

	// 转换为绝对路径
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, adk.AgentMiddleware{}, err
	}

	global.Log.Info("初始化 Skill 后端", "baseDir", absBaseDir)

	// 创建本地文件系统后端
	// LocalBackend 提供对本地文件系统的访问
	// 用于支持技能中的文件操作命令
	be, err := local.NewBackend(ctx, &local.Config{})
	if err != nil {
		return nil, adk.AgentMiddleware{}, err
	}

	// 创建 Skill 后端
	// 基于文件系统加载技能目录下的所有技能
	// 扫描 BaseDir 下的一级子目录，查找每个子目录中的 SKILL.md 文件
	skillBackend, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
		Backend: be,
		BaseDir: absBaseDir,
	})
	if err != nil {
		return nil, adk.AgentMiddleware{}, err
	}

	// 创建 Skill 中间件（使用旧版本以获取指令和工具）
	// 该中间件会自动将相关的技能注入到 ChatModel 的上下文中
	middleware, err := skill.New(ctx, &skill.Config{
		Backend: skillBackend,
	})
	if err != nil {
		global.Log.Error("创建 Skills 中间件失败", "error", err)
		return nil, adk.AgentMiddleware{}, err
	}

	// 列出所有可用技能
	skills, err := skillBackend.List(ctx)
	if err != nil {
		global.Log.Warn("列出技能失败", "error", err)
	} else {
		global.Log.Info("发现可用技能", "count", len(skills))
		for _, s := range skills {
			global.Log.Info("技能", "name", s.Name, "description", s.Description)
		}
	}

	return skillBackend, middleware, nil
}
