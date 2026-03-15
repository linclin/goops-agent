# GoOps Agent 项目

## 项目概述

GoOps Agent 是一个基于 Go 语言和 React 的 AI 代理系统，集成了先进的大语言模型能力和工具调用功能，为用户提供智能对话和自动化操作能力。

## 项目结构

```
goops-agent/
├── antd-x/          # 前端项目（React + Ant Design X）
└── gin-mini-agent/  # 后端项目（Go + Gin + Eino）
```

## 功能特性

### 核心功能
- **智能对话**：基于大语言模型的流式对话能力
- **工具调用**：集成多种工具，支持浏览器自动化、HTTP请求、文件操作、命令执行、python脚本等
- **Skills 集成**：支持通过官方 Skill Middleware 加载和执行预定义技能
- **RAG 增强**：检索增强生成，提升模型回答的准确性和时效性
- **向量检索**：支持 Chromem、Milvus 和 Redis 向量数据库
- **文档处理**：支持文件加载、分割和向量化
- **对话历史**：支持将对话历史存储到向量数据库中进行上下文检索
- **自我提升**：集成 self-improving-agent 技能，支持系统自我学习和知识库更新
- **结构化日志**：使用 slog 进行详细的日志记录，便于调试和监控

### Tools 工具集成
| 工具名称 | 功能说明 | 实现方式 |
|---------|---------|---------|
| command | 在终端中执行命令并返回输出。支持 Windows (PowerShell)、Linux 和 macOS (sh) | 自定义实现 |
| open | 打开文件或URL，读取内容 | 自定义实现 |
| str_replace_editor | 文件编辑器，支持创建、查看、编辑文件 | 官方库 eino-ext  |
| python_execute | 执行 Python 代码字符串 | 官方库 eino-ext  |
| request_get | 发送 HTTP GET 请求 | 官方库 eino-ext |
| request_post | 发送 HTTP POST 请求 | 官方库 eino-ext |
| request_put | 发送 HTTP PUT 请求 | 官方库 eino-ext |
| request_delete | 发送 HTTP DELETE 请求 | 官方库 eino-ext |
| browser_use | 浏览器自动化，支持网页交互和内容提取 | 官方库 eino-ext |
| mcp_filesystem | MCP 文件系统工具，读写项目目录文件 | MCP 官方服务器 |
| mcp_fetch | MCP Fetch 工具，获取网页内容 | MCP 官方服务器 |
| mcp_memory | MCP Memory 工具，存储和检索记忆 | MCP 官方服务器 |
| kubectl | Kubernetes 多集群管理工具，支持 get、describe、create、delete、apply 等操作 | 基于 client-go SDK 实现 |

### MCP 工具说明

MCP (Model Context Protocol) 是由 Anthropic 推出的标准协议，用于 LLM 应用和外部数据源或工具之间通信。

**当前集成的 MCP 工具：**

| MCP 工具 | 功能 | 依赖 | 说明 |
|---------|------|------|------|
| mcp_filesystem | 文件系统操作 | npx | 允许 AI 读写项目目录下的文件 |
| mcp_fetch | 网页内容获取 | uvx | 允许 AI 获取网页内容 |
| mcp_memory | 记忆存储 | npx | 允许 AI 存储和检索记忆信息 |

**环境要求：**
- 需要安装 Node.js 和 npx（用于 filesystem 和 memory 工具）
- 需要安装 Python 和 uvx（用于 fetch 工具）
- 如果环境不满足，MCP 工具会自动跳过，不影响其他工具使用

### kubectl 工具说明

**功能：**
- 基于 client-go SDK 实现的 Kubernetes 集群管理工具
- 支持 get、describe、create、delete、apply 等核心 kubectl 命令
- 支持多集群管理，自动从 ~/.kube/ 目录加载集群配置
- 自动从 kubeconfig 文件或集群内服务账号获取集群配置

**支持的命令：**

| 命令 | 功能 | 示例 |
|------|------|------|
| get | 获取资源信息 | `kubectl get pods` |
| describe | 获取资源详细信息 | `kubectl describe pod my-pod` |
| create | 创建资源 | 提供 YAML/JSON 内容创建资源 |
| delete | 删除资源 | `kubectl delete pod my-pod` |
| apply | 创建或更新资源 | 提供 YAML/JSON 内容应用配置 |

**多集群支持：**
- **配置方式：** 在 ~/.kube/ 目录下创建以集群名称命名的 kubeconfig 文件（如 ST-XXX-XX）
- **使用方式：** 在工具调用时通过 `cluster` 参数指定集群名称 

**环境要求：**
- 需要在 ~/.kube/ 目录下配置 kubeconfig 文件（以集群名称命名）
- 或者在 Kubernetes 集群内运行（使用服务账号）
- 如果未配置 kubeconfig，kubectl 工具会自动跳过，不影响其他工具使用

### Skills 技能列表

| 技能名称 | 版本 | 描述 | 用途 |
|---------|------|------|------|
| self-improving-agent-3.0.1 | 3.0.1 | 自我提升技能，支持系统自动学习、知识库更新、错误处理和持续优化 | 支持系统自我学习和知识库更新 |
| ssh-essentials-1.0.0 | 1.0.0 | 安全远程访问、密钥管理、隧道和文件传输的基本 SSH 命令 | 提供 SSH 远程访问和文件传输功能 |
| ansible-1.0.0 | 1.0.0 | 避免常见的 Ansible 错误 — YAML 语法陷阱、变量优先级、幂等性失败和处理器问题 | 提供 Ansible 配置管理和自动化部署支持 |
| clickhouse-1.0.1 | 1.0.1 | 查询、优化和管理 ClickHouse OLAP 数据库，包括模式设计、性能调优和数据导入模式 | 提供 ClickHouse 数据库管理和分析功能 |
| database-operations-1.0.0 | 1.0.0 | 用于设计数据库架构、编写迁移脚本、优化 SQL 查询、解决 N+1 问题、创建索引、配置 PostgreSQL、设置 EF Core、实现缓存、分区表或任何数据库性能问题 | 提供全面的数据库设计、迁移和优化功能 |
| docker-essentials-1.0.0 | 1.0.0 | Docker 容器管理、镜像操作和调试的基本命令和工作流程 | 提供 Docker 容器管理和镜像操作功能 |
| k8s-fta-skill-0.1.0 | 0.1.0 | 基于FTA故障树分析法的Kubernetes问题定位和修复工具，支持Pod运行异常、服务访问失败、RBAC权限问题、DNS解析失败、OOMKilled等问题的自动排查和修复 | 提供 Kubernetes 故障排查和自动修复功能 |
| k8s-skill-1.0.1 | 1.0.1 | 腾讯云 TKE 容器服务运维专家，支持集群巡检、状态查询、节点池管理、kubeconfig 获取等 | 提供腾讯云 TKE 集群管理和运维功能 |
| kubectl-1.0.0 | 1.0.0 | 通过 kubectl 命令执行和管理 Kubernetes 集群。查询资源、部署应用、调试容器、管理配置和监控集群健康状态 | 提供 Kubernetes 集群管理和资源操作功能 |
| linux-1.0.0 | 1.0.0 | 操作 Linux 系统，避免权限陷阱、静默失败和常见管理错误 | 提供 Linux 系统操作和故障排查功能 |
| monitoring-1.0.0 | 1.0.0 | 为应用程序和基础设施设置可观测性，包括指标、日志、追踪和告警 | 提供系统监控和可观测性功能 |
| mysql-1.0.1 | 1.0.1 | 编写正确的 MySQL 查询，包括适当的字符集、索引、事务和生产模式 | 提供 MySQL 数据库管理和查询优化功能 |
| prometheus-1.1.0 | 1.1.0 | 查询 Prometheus 监控数据以检查服务器指标、资源使用和系统健康状态 | 提供 Prometheus 监控数据查询功能 |
| redis-store-1.0.0 | 1.0.0 | 有效地使用 Redis 进行缓存、队列和数据结构，包括适当的过期和持久化 | 提供 Redis 缓存和数据结构管理功能 |
| sql-toolkit-1.0.0 | 1.0.0 | 查询、设计、迁移和优化 SQL数据库。用于 SQLite、PostgreSQL 或 MySQL — 架构设计、编写查询、创建迁移、索引、备份/恢复和调试慢查询 | 提供 SQL 数据库设计、查询和管理功能 |
| terraform-1.0.0 | 1.0.0 | 避免常见的 Terraform 错误 — 状态损坏、count vs for_each、生命周期陷阱和依赖顺序 | 提供 Terraform 基础设施即代码管理功能 |

## 技术栈

### 前端技术

| 技术/框架 | 版本 | 用途 |
|---------|------|------|
| React | ^18.3.1 | 前端框架 |
| Ant Design | ^5.29.3 | UI 组件库 |
| @ant-design/x | ^1.6.1 | Ant Design X 组件库 |
| @ant-design/x-sdk | ^2.3.0 | Ant Design X SDK |
| @ant-design/x-markdown | ^2.3.0 | Markdown 渲染组件 |
| TypeScript | ^5.9.3 | 类型系统 |
| Vite | ^5.4.21 | 构建工具 |
| antd-style | ^3.7.1 | Ant Design 样式工具 |
| dayjs | ^1.11.19 | 日期处理库 |

### 后端技术

| 技术/框架 | 版本 | 用途 |
|---------|------|------|
| Go | 1.26 | 后端开发语言 |
| Gin | ^1.12.0 | Web 框架 |
| CloudWeGo Eino | ^0.8.0 | AI 框架 |
| eino-ext | latest | Eino 扩展组件（HTTP请求、浏览器自动化等） |
| Milvus | ^2.6.2 | 向量数据库 |
| Redis | ^9.18.0 | 缓存和向量存储 |
| Chromem | ^0.7.0 | 向量检索 |
| OpenAI | - | 模型和嵌入 |
| Viper | ^1.21.0 | 配置管理 |
| Cron | ^3.0.1 | 定时任务 |
| slog | 内置 | 结构化日志 |

## 快速开始

### 前端启动

```bash
# 进入前端目录
cd antd-x

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端应用将在 `http://localhost:3000` 运行。

### 后端启动

```bash
# 进入后端目录
cd gin-mini-agent

# 安装依赖
go mod tidy

# 启动后端服务器
go run main.go
```

后端服务将在 `http://localhost:8080` 运行。

## API 接口

### 流式对话接口

- **URL**: `/api/v1/agent/chat/sse`
- **方法**: `POST`
- **Content-Type**: `application/json`
- **请求体**:
  ```json
  {
    "message": "你好，请帮我查询天气"
  }
  ```
- **响应**: SSE (Server-Sent Events) 流式返回

### 同步对话接口

- **URL**: `/api/v1/agent/chat/sync`
- **方法**: `POST`
- **Content-Type**: `application/json`
- **请求体**:
  ```json
  {
    "message": "你好，请帮我查询天气"
  }
  ```
- **响应**: JSON 格式完整响应

### RAG 索引接口

- **URL**: `/api/v1/rag/index`
- **方法**: `POST`
- **Content-Type**: `multipart/form-data`
- **响应**: 索引状态

## 配置说明

后端配置文件位于 `gin-mini-agent/conf/` 目录：

- `config.prd.yml` - 生产环境配置
- `config.st.yml` - 测试环境配置
- `config.se.yml` - 开发环境配置

主要配置项：
- **模型配置**：OpenAI API 密钥、模型名称、Base URL
- **向量数据库**：Chromem/Milvus/Redis 连接信息
- **RAG 配置**：文档分割、嵌入模型配置
- **工具配置**：浏览器、HTTP请求等工具配置

## 项目架构

### 前端架构

```
antd-x/
├── src/
│   ├── pages/          # 页面组件
│   ├── _utils/         # 工具函数
│   ├── App.tsx         # 主应用组件
│   ├── main.tsx        # 应用入口
│   └── index.css       # 全局样式
├── package.json        # 依赖配置
└── vite.config.ts      # Vite 配置
```

### 后端架构

```
gin-mini-agent/
├── api/               # API 接口
│   └── v1/            # API 版本
│       └── agent/     # Agent 相关接口
├── internal/          # 内部实现
│   ├── ai_agent/      # AI 代理实现
│   │   ├── tools/     # 工具定义
│   │   │   ├── open.go           # 文件打开工具
│   │   │   ├── fileeditor.go     # 文件编辑工具
│   │   │   ├── pyexecutor.go     # Python执行工具
│   │   │   ├── httprequest.go    # HTTP请求工具
│   │   │   ├── browseruse.go     # 浏览器自动化工具
│   │   │   └── skill.go          # 技能加载工具
│   │   ├── retriever/ # 向量检索器实现
│   │   │   ├── chromem_retriever.go
│   │   │   ├── redis_retriever.go
│   │   │   └── milvus_retriever.go
│   │   ├── ai_agent.go           # Agent 主逻辑
│   │   ├── tools_node.go         # 工具注册
│   │   ├── chat_template.go      # 提示词模板
│   │   └── conversation_manager.go # 对话历史管理
│   └── rag_index/     # RAG 索引实现
├── middleware/        # 中间件
├── initialize/        # 初始化
├── conf/              # 配置文件
├── router/            # 路由定义
└── main.go            # 应用入口
```

## 工具使用示例

### 文件打开工具

```json
{
  "path": "/path/to/file.txt"
}
```

### 文件编辑工具

```json
{
  "command": "view",
  "path": "/path/to/file.txt",
  "view_range": [1, 100]
}
```

```json
{
  "command": "str_replace",
  "path": "/path/to/file.txt",
  "old_str": "旧内容",
  "new_str": "新内容"
}
```

### Python 执行工具

```json
{
  "code": "print('Hello, World!')"
}
```

### HTTP 请求工具

GET 请求：
```json
"https://api.example.com/data"
```

POST 请求：
```json
{
  "url": "https://api.example.com/create",
  "body": {"name": "test", "value": 123}
}
```

### 浏览器自动化工具

```json
{
  "action": "go_to_url",
  "url": "https://www.example.com"
}
```

### 技能加载工具

```json
{
  "skill": "kubernetes-1.0.1"
}
```

## 技能管理

### 技能目录结构

```
skills/
  ├── kubernetes-1.0.1/      # 技能目录（版本化）
  │   └── SKILL.md          # 技能指令和元数据
  │   ├── scripts/          # 可选：可执行脚本
  │   └── references/       # 可选：参考文档
  ├── self-improving-agent/  # 自我提升技能
  │   └── SKILL.md
  └── ...
```

### 技能格式

每个技能需要在 `SKILL.md` 文件中定义：

```markdown
---
name: kubernetes-1.0.1
description: Kubernetes 操作指南
version: 1.0.1
author: GoOps Team
---

# Kubernetes 技能

## 功能说明
- 提供 Kubernetes 集群管理命令
- 支持 Pod、Service、Deployment 等资源操作
- 包含常见故障排查指南

## 使用示例
- 如何查看 Pod 日志
- 如何部署应用
- 如何扩展集群
```

## RAG 文档目录

### 文档存储结构

```
rag_docs/
  ├── conventions/     # 项目约定和规范
  ├── learnings/       # 系统学习的知识
  ├── errors/          # 错误处理和解决方案
  └── features/        # 功能特性文档
```

### 文档管理

- **conventions/**: 存储项目约定、编码规范、架构设计等文档
- **learnings/**: 存储系统通过自我提升学习到的知识
- **errors/**: 存储常见错误和解决方案
- **features/**: 存储功能特性的详细说明

## 部署说明

### 前端部署

```bash
# 构建生产版本
npm run build

# 部署到静态文件服务器
```

### 后端部署

```bash
# 构建可执行文件
go build -o goops-agent main.go

# 运行服务
./goops-agent
```

## 依赖管理

### 前端依赖

使用 npm 管理前端依赖：

```bash
# 安装依赖
npm install

# 更新依赖
npm update
```

### 后端依赖

使用 Go Modules 管理后端依赖：

```bash
# 安装依赖
go mod tidy

# 更新依赖
go get -u
```

## 开发指南

### 前端开发

1. 确保 Node.js >= 16.0
2. 安装依赖：`npm install`
3. 启动开发服务器：`npm run dev`
4. 构建生产版本：`npm run build`

### 后端开发

1. 确保 Go >= 1.26
2. 安装依赖：`go mod tidy`
3. 启动开发服务器：`go run main.go`
4. 构建可执行文件：`go build`

### 添加新工具

1. 在 `internal/ai_agent/tools/` 目录下创建新工具文件
2. 实现 `ToEinoTool()` 方法转换为 Eino 工具接口
3. 实现 `Invoke()` 方法处理工具调用逻辑
4. 在 `tools_node.go` 的 `GetTools()` 函数中注册新工具

### 添加新技能

1. 在 `skills/` 目录下创建新技能目录（建议使用版本号）
2. 在技能目录中创建 `SKILL.md` 文件
3. 按照技能格式定义技能内容
4. 重启应用后技能会自动加载

## 日志记录

### 日志配置

系统使用 Go 标准库的 `slog` 进行结构化日志记录：

- **INFO 级别**：记录正常的系统运行信息
- **DEBUG 级别**：记录详细的调试信息
- **ERROR 级别**：记录错误信息

### 日志输出

- 控制台输出：实时查看系统运行状态
- 技能调用日志：记录技能的加载和调用过程
- 工具调用日志：记录工具的执行情况
- 错误日志：记录系统错误和异常

## 自我提升

系统集成了 `self-improving-agent` 技能，支持：

- **自动学习**：从用户交互中学习新知识
- **知识库更新**：将学习到的知识存储到 `rag_docs/learnings/` 目录
- **错误处理**：记录常见错误和解决方案到 `rag_docs/errors/` 目录
- **持续优化**：基于用户反馈不断改进系统性能

## 许可证

Apache License 2.0

---

**GoOps Agent** - 智能 AI 代理系统，为您提供高效、智能的自动化服务。
