// Package global 提供全局配置和变量
//
// 该包定义了应用程序的全局配置结构和变量。
// 主要功能包括：
//   - 配置文件解析和存储
//   - 全局变量管理
//   - 日志和翻译器初始化
//
// 配置文件格式: YAML
// 配置文件路径: ./config.yaml（默认）
package global

import "log/slog"

// Configuration 系统配置结构
//
// 该结构体定义了应用程序的完整配置。
// 使用 mapstructure 标签映射 YAML 配置文件。
// YAML 文件使用 "-" 分隔单词，转为驼峰命名方便使用。
//
// 配置示例 (config.yaml):
//
//	system:
//	  app-name: "gin-mini-agent"
//	  run-mode: "debug"
//	  port: 8080
//	logs:
//	  level: "info"
//	  path: "./logs/app.log"
//	auth:
//	  user: "admin"
//	  password: "password"
//	rag:
//	  type: "redis"
//	  redis:
//	    addr: "localhost:6379"
//	ai-model:
//	  chat-model:
//	    base-url: "https://api.openai.com/v1"
//	    api-key: "sk-xxx"
//	    model: "gpt-4"
type Configuration struct {
	// System 系统配置
	// 包含应用名称、运行模式、端口等基础配置
	System SystemConfiguration `mapstructure:"system" json:"system"`

	// Logs 日志配置
	// 包含日志级别、路径、轮转策略等
	Logs LogsConfiguration `mapstructure:"logs" json:"logs"`

	// Auth 认证配置
	// 包含 Basic Auth 的用户名和密码
	Auth AuthConfiguration `mapstructure:"auth" json:"auth"`

	// RAG 检索增强生成配置
	// 包含向量数据库类型和连接信息
	RAG RAGConfiguration `mapstructure:"rag" json:"rag"`

	// AiModel AI 模型配置
	// 包含聊天模型和嵌入模型的配置
	AiModel AiModelConfiguration `mapstructure:"ai-model" json:"ai-model"`
}

// SystemConfiguration 系统配置
//
// 包含应用程序的基础配置信息。
type SystemConfiguration struct {
	// AppName 应用名称
	// 用于日志标识和进程管理
	AppName string `mapstructure:"app-name" json:"appName"`

	// RunMode 运行模式
	// 可选值: debug, release, test
	// debug: 开发模式，输出详细日志
	// release: 生产模式，优化性能
	// test: 测试模式
	RunMode string `mapstructure:"run-mode" json:"runMode"`

	// UrlPathPrefix URL 路径前缀
	// 用于 API 版本控制，如 "/api/v1"
	UrlPathPrefix string `mapstructure:"url-path-prefix" json:"urlPathPrefix"`

	// Port 服务端口
	// HTTP 服务监听的端口号
	Port int `mapstructure:"port" json:"port"`

	// BaseApi 基础 API 地址
	// 用于构建完整的 API URL
	BaseApi string `mapstructure:"base-api" json:"baseApi"`

	// Transaction 是否启用事务
	// 数据库事务开关
	Transaction bool `mapstructure:"transaction" json:"transaction"`
}

// LogsConfiguration 日志配置
//
// 配置日志的输出方式和轮转策略。
type LogsConfiguration struct {
	// Level 日志级别
	// 可选值: debug, info, warn, error
	Level slog.Level `mapstructure:"level" json:"level"`

	// Path 日志文件路径
	// 日志文件的存储位置
	Path string `mapstructure:"path" json:"path"`

	// MaxSize 单个日志文件最大大小（MB）
	// 超过此大小会触发日志轮转
	MaxSize int `mapstructure:"max-size" json:"maxSize"`

	// MaxBackups 保留的旧日志文件最大数量
	// 超过此数量会删除最旧的日志文件
	MaxBackups int `mapstructure:"max-backups" json:"maxBackups"`

	// MaxAge 保留旧日志文件的最大天数
	// 超过此天数的日志文件会被删除
	MaxAge int `mapstructure:"max-age" json:"maxAge"`

	// Compress 是否压缩旧日志文件
	// 使用 gzip 压缩，节省磁盘空间
	Compress bool `mapstructure:"compress" json:"compress"`
}

// AuthConfiguration 认证配置
//
// 配置 Basic HTTP 认证的用户名和密码。
type AuthConfiguration struct {
	// User 用户名
	// Basic Auth 认证用户名
	User string `mapstructure:"user" json:"user"`

	// Password 密码
	// Basic Auth 认证密码
	Password string `mapstructure:"password" json:"password"`
}

// RAGConfiguration 检索增强生成配置
//
// 配置向量数据库的类型和连接信息。
type RAGConfiguration struct {
	// Type 向量数据库类型
	// 可选值: chromem, redis, milvus
	// chromem: 本地文件存储，适合开发
	// redis: 分布式存储，适合中等规模
	// milvus: 分布式向量数据库，适合大规模
	Type string `mapstructure:"type" json:"type"`

	// Redis Redis 配置
	// 当 type 为 redis 时使用
	Redis RedisConfiguration `mapstructure:"redis" json:"redis"`

	// Chromem Chromem 配置
	// 当 type 为 chromem 时使用
	Chromem ChromemConfiguration `mapstructure:"chromem" json:"chromem"`

	// Milvus Milvus 配置
	// 当 type 为 milvus 时使用
	Milvus MilvusConfiguration `mapstructure:"milvus" json:"milvus"`
}

// RedisConfiguration Redis 配置
//
// 配置 Redis 数据库连接信息。
type RedisConfiguration struct {
	// Addr Redis 服务器地址
	// 格式: host:port，如 "localhost:6379"
	Addr string `mapstructure:"addr" json:"addr"`

	// Prefix 键前缀
	// 用于区分不同应用的数据
	// 示例: "myapp:"
	Prefix string `mapstructure:"prefix" json:"prefix"`
}

// ChromemConfiguration Chromem 配置
//
// 配置 Chromem 本地向量数据库。
type ChromemConfiguration struct {
	// Path 数据存储路径
	// 向量数据的持久化存储位置
	// 默认: "./data/chromem"
	Path string `mapstructure:"path" json:"path"`

	// Collection 集合名称
	// 向量数据的集合名称
	// 默认: "rag_collection"
	Collection string `mapstructure:"collection" json:"collection"`
}

// MilvusConfiguration Milvus 配置
//
// 配置 Milvus 分布式向量数据库。
type MilvusConfiguration struct {
	// Addr Milvus 服务器地址
	// 格式: host:port，如 "localhost:19530"
	Addr string `mapstructure:"addr" json:"addr"`

	// Username 用户名
	// Milvus 认证用户名（可选）
	Username string `mapstructure:"username" json:"username"`

	// Password 密码
	// Milvus 认证密码（可选）
	Password string `mapstructure:"password" json:"password"`

	// Collection 集合名称
	// 向量数据的集合名称
	Collection string `mapstructure:"collection" json:"collection"`
}

// AiModelConfiguration AI 模型配置
//
// 配置聊天模型和嵌入模型。
type AiModelConfiguration struct {
	// EmbeddingModel 嵌入模型配置
	// 用于将文本转换为向量
	EmbeddingModel ChatConfiguration `mapstructure:"embedding-model" json:"embeddingModel"`

	// ChatModel 聊天模型配置
	// 用于生成对话响应
	ChatModel ChatConfiguration `mapstructure:"chat-model" json:"chatModel"`
}

// ChatConfiguration 聊天/嵌入模型配置
//
// 配置 OpenAI 兼容的 API 接口。
type ChatConfiguration struct {
	// BaseURL API 基础地址
	// 示例:
	// - OpenAI: https://api.openai.com/v1
	// - Azure: https://your-resource.openai.azure.com
	// - 国内代理: https://api.your-proxy.com/v1
	BaseURL string `mapstructure:"base-url" json:"baseURL"`

	// APIKey API 访问密钥
	// 用于身份验证
	APIKey string `mapstructure:"api-key" json:"apiKey"`

	// Model 模型名称
	// 聊天模型示例: gpt-4, gpt-3.5-turbo, claude-3
	// 嵌入模型示例: text-embedding-ada-002, text-embedding-3-small
	Model string `mapstructure:"model" json:"model"`
}
