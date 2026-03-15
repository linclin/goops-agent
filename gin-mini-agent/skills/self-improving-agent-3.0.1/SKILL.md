---
name: self-improving-agent
description: "记录学习内容、错误和纠正，实现持续改进。使用场景：(1) 命令或操作意外失败，(2) 用户纠正 AI（'不对...'、'实际上...'），(3) 用户请求不存在的功能，(4) 外部 API 或工具失败，(5) 发现知识过时或不正确，(6) 发现重复任务的更好方法。在主要任务前也应审查学习内容。"
metadata:
  author: gin-mini-agent
  version: 3.0.1
---

# 自我改进技能

将学习内容和错误直接记录到 `rag_docs/self-improving-agent/` 目录，作为 RAG 知识库的一部分。系统在后续对话中可自动检索这些知识，实现持续改进。

## 快速参考

| 情况 | 操作 |
|-----------|--------|
| 命令/操作失败 | 记录到 `rag_docs/self-improving-agent/errors/` 目录 |
| 用户纠正你 | 记录到 `rag_docs/self-improving-agent/learnings/` 目录，标签 `correction` |
| 用户想要缺失的功能 | 记录到 `rag_docs/self-improving-agent/features/` 目录 |
| API/外部工具失败 | 记录到 `rag_docs/self-improving-agent/errors/` 目录，包含集成详情 |
| 知识已过时 | 记录到 `rag_docs/self-improving-agent/learnings/` 目录，标签 `knowledge_gap` |
| 发现更好的方法 | 记录到 `rag_docs/self-improving-agent/learnings/` 目录，标签 `best_practice` |
| 项目约定 | 记录到 `rag_docs/self-improving-agent/conventions/` 目录 |

## 目录结构

```
rag_docs/self-improving-agent/
├── conventions/          # 项目约定（编码规范、工作流程等）
│   ├── coding-style.md
│   └── workflow.md
├── learnings/            # 学习记录（纠正、最佳实践、知识缺口）
│   ├── LRN-20250315-001-slog-logging.md
│   └── LRN-20250315-002-error-handling.md
├── errors/               # 错误记录（失败案例及解决方案）
│   ├── ERR-20250315-001-build-timeout.md
│   └── ERR-20250315-002-api-failure.md
└── features/             # 功能请求（用户需求记录）
    └── FEAT-20250315-001-new-tool.md
```

## 记录格式

### 学习记录

创建文件 `rag_docs/self-improving-agent/learnings/LRN-YYYYMMDD-XXX-brief-title.md`：

```markdown
# [LRN-YYYYMMDD-XXX] 简要标题

**类别**: correction | best_practice | knowledge_gap
**优先级**: low | medium | high | critical
**状态**: pending | resolved | in_progress
**区域**: frontend | backend | infra | tests | docs | config
**相关文件**: path/to/file.ext

## 背景

描述发生了什么，为什么会发现这个学习点。

## 内容

### 错误做法
描述之前错误或不优的做法。

### 正确做法
描述正确的做法或发现的知识。

## 示例

```go
// 代码示例
slog.InfoContext(ctx, "[tool] 操作开始", "param", value)
```

## 相关条目

- 相关条目 ID（如有）
```

### 错误记录

创建文件 `rag_docs/self-improving-agent/errors/ERR-YYYYMMDD-XXX-brief-title.md`：

```markdown
# [ERR-YYYYMMDD-XXX] 简要标题

**工具/命令**: tool_name 或 command
**优先级**: low | medium | high | critical
**状态**: pending | resolved | wont_fix
**区域**: frontend | backend | infra | tests | docs | config
**可复现**: yes | no | unknown
**相关文件**: path/to/file.ext

## 错误描述

简要描述失败的内容。

## 错误信息

```
实际错误消息或输出
```

## 上下文

- 尝试的操作：...
- 使用的参数：...
- 环境信息：...

## 解决方案

描述如何解决这个问题。

## 预防措施

描述如何避免类似问题。

## 相关条目

- 相关错误 ID（如有）
```

### 功能请求

创建文件 `rag_docs/self-improving-agent/features/FEAT-YYYYMMDD-XXX-brief-title.md`：

```markdown
# [FEAT-YYYYMMDD-XXX] 简要标题

**优先级**: low | medium | high | critical
**状态**: pending | in_progress | completed | wont_implement
**复杂度**: simple | medium | complex
**频率**: first_time | recurring

## 需求描述

用户想要实现什么功能。

## 使用场景

为什么需要这个功能，解决什么问题。

## 建议实现

描述可能的实现方案。

## 相关功能

- 相关功能名称（如有）
```

### 项目约定

创建文件 `rag_docs/self-improving-agent/conventions/topic-name.md`：

```markdown
# 约定主题

## 概述

简要描述这个约定的目的。

## 规范

### 规则 1
具体规则描述。

### 规则 2
具体规则描述。

## 示例

```go
// 正确示例
slog.InfoContext(ctx, "message", "key", value)

// 错误示例
fmt.Printf("message: %v\n", value)
```

## 原因

为什么要有这个约定。
```

## ID 生成规则

格式：`TYPE-YYYYMMDD-XXX`
- TYPE: `LRN`（学习）、`ERR`（错误）、`FEAT`（功能）
- YYYYMMDD: 当前日期
- XXX: 顺序号（如 `001`、`002`）

示例：`LRN-20250315-001`、`ERR-20250315-001`、`FEAT-20250315-001`

## 工作流程

### 记录学习内容

1. **识别学习点**：当发现值得记录的内容时
2. **创建文件**：使用 fileeditor 工具在 `rag_docs/` 对应目录创建文件
3. **填写内容**：按照格式模板填写详细信息
4. **自动生效**：RAG 系统会在下次索引时自动读取

### 解决问题后

1. **更新状态**：将 `状态: pending` 改为 `状态: resolved`
2. **添加解决方案**：在文件中补充解决方案部分
3. **提升为约定**：如果是通用规则，移动到 `conventions/` 目录

### 定期审查

- 每周审查 `rag_docs/` 目录下的待处理条目
- 将已解决的学习内容提炼为项目约定
- 删除过时或不再相关的条目

## 检测触发器

当注意到以下情况时自动记录：

**纠正**（→ `learnings/` 目录，类别 `correction`）：
- "不，那不对..."
- "实际上，应该是..."
- "你错了..."
- "那已经过时了..."

**功能请求**（→ `features/` 目录）：
- "你能不能也..."
- "我希望你可以..."
- "有没有办法..."
- "为什么你不能..."

**知识缺口**（→ `learnings/` 目录，类别 `knowledge_gap`）：
- 用户提供了你不知道的信息
- 文档已过时
- API 行为与预期不同

**错误**（→ `errors/` 目录）：
- 工具调用返回错误
- 命令执行失败
- 意外的输出或行为

## 优先级指南

| 优先级 | 使用场景 |
|--------|----------|
| `critical` | 阻塞核心功能、数据丢失风险、安全问题 |
| `high` | 重大影响、影响常见工作流、重复问题 |
| `medium` | 中等影响、存在变通方案 |
| `low` | 轻微不便、边缘情况、锦上添花 |

## 区域标签

| 区域 | 范围 |
|------|------|
| `frontend` | UI、组件、客户端代码 |
| `backend` | API、服务、服务端代码 |
| `infra` | CI/CD、部署、Docker、云 |
| `tests` | 测试文件、测试工具、覆盖率 |
| `docs` | 文档、注释、README |
| `config` | 配置文件、环境、设置 |

## 最佳实践

1. **立即记录** - 问题发生后立即创建文件
2. **文件名清晰** - 使用 ID + 简要标题
3. **内容完整** - 包含背景、内容、示例
4. **定期整理** - 将已解决的内容提炼为约定
5. **关联引用** - 在相关条目间建立链接

## 示例操作

### 记录一个错误

使用 fileeditor 工具：

```
命令: create
路径: rag_docs/self-improving-agent/errors/ERR-20250315-001-build-timeout.md
内容:
# [ERR-20250315-001] 构建命令超时

**工具/命令**: command
**优先级**: high
**状态**: pending
**区域**: backend
**可复现**: unknown

## 错误描述
执行 go build 命令时超时。

## 错误信息
\`\`\`
Error: command timeout after 30s
\`\`\`

## 上下文
- 尝试的操作：go build ./...
- 环境信息：Windows PowerShell

## 解决方案
（待补充）

## 预防措施
（待补充）
```

### 记录一个学习点

使用 fileeditor 工具：

```
命令: create
路径: rag_docs/self-improving-agent/learnings/LRN-20250315-001-slog-logging.md
内容:
# [LRN-20250315-001] 使用 slog 记录日志

**类别**: best_practice
**优先级**: high
**状态**: resolved
**区域**: backend

## 背景
项目需要统一的日志记录方式，便于调试和监控。

## 内容

### 错误做法
使用 fmt.Printf 或 log.Println 记录日志。

### 正确做法
使用 Go 标准库 slog 进行结构化日志记录。

## 示例

\`\`\`go
// 正确
slog.InfoContext(ctx, "[tool] 操作开始", "param", value)
slog.ErrorContext(ctx, "[tool] 操作失败", "error", err)

// 错误
fmt.Printf("操作开始: %v\\n", value)
\`\`\`
```

### 创建项目约定

使用 fileeditor 工具：

```
命令: create
路径: rag_docs/self-improving-agent/conventions/logging.md
内容:
# 日志记录约定

## 概述
本项目使用 Go 标准库 slog 进行日志记录。

## 规范

### 使用 Context
所有日志调用应使用 *Context 版本，便于追踪。

### 日志级别
- Info: 正常操作信息
- Debug: 调试信息
- Warn: 警告信息
- Error: 错误信息

### 格式规范
日志消息格式：`[模块名] 操作描述`

## 示例

\`\`\`go
slog.InfoContext(ctx, "[fileeditor] 文件创建成功", "path", path)
slog.ErrorContext(ctx, "[command] 命令执行失败", "cmd", cmd, "error", err)
\`\`\`
```

## 与 RAG 系统集成

### 自动索引

当文件创建到 `rag_docs/self-improving-agent/` 目录后：

1. 系统会自动检测新文件
2. 调用 RAG 索引 API 将内容向量化
3. 后续对话中 Agent 可自动检索相关知识

### 知识检索

在对话中，Agent 会：

1. 根据用户问题检索相关知识
2. 从 `rag_docs/self-improving-agent/` 中获取相关文档
3. 应用学习到的知识解决问题

### 持续改进循环

```
发现问题 → 记录到 rag_docs → RAG 索引 → 后续对话检索应用 → 发现新问题
```

这个闭环确保系统不断学习和改进。
