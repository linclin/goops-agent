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
- **工具调用**：集成多种工具，支持浏览器自动化、HTTP请求、文件操作、代码执行等
- **RAG 增强**：检索增强生成，提升模型回答的准确性和时效性
- **向量检索**：支持 Chromem、Milvus 和 Redis 向量数据库
- **文档处理**：支持文件加载、分割和向量化
- **对话历史**：支持将对话历史存储到向量数据库中进行上下文检索

### 工具集成
| 工具名称 | 功能说明 | 实现方式 |
|---------|---------|---------|
| open | 打开文件或URL，读取内容 | 自定义实现 |
| str_replace_editor | 文件编辑器，支持创建、查看、编辑文件 | 官方库 eino-ext  |
| python_execute | 执行 Python 代码字符串 | 官方库 eino-ext  |
| request_get | 发送 HTTP GET 请求 | 官方库 eino-ext |
| request_post | 发送 HTTP POST 请求 | 官方库 eino-ext |
| request_put | 发送 HTTP PUT 请求 | 官方库 eino-ext |
| request_delete | 发送 HTTP DELETE 请求 | 官方库 eino-ext |
| browser_use | 浏览器自动化，支持网页交互和内容提取 | 官方库 eino-ext |

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
│   │   │   └── browseruse.go     # 浏览器自动化工具
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

## 许可证

Apache License 2.0

---

**GoOps Agent** - 智能 AI 代理系统，为您提供高效、智能的自动化服务。
