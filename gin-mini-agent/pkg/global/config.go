package global

import "log/slog"

// Configuration 系统配置, 配置字段可参见yml注释
// viper内置了mapstructure, yml文件用"-"区分单词, 转为驼峰方便
type Configuration struct {
	System  SystemConfiguration  `mapstructure:"system" json:"system"`
	Logs    LogsConfiguration    `mapstructure:"logs" json:"logs"`
	Auth    AuthConfiguration    `mapstructure:"auth" json:"auth"`
	RAG     RAGConfiguration     `mapstructure:"rag" json:"rag"`
	AiModel AiModelConfiguration `mapstructure:"ai-model" json:"ai-model"`
}

type SystemConfiguration struct {
	AppName       string `mapstructure:"app-name" json:"appName"`
	RunMode       string `mapstructure:"run-mode" json:"runMode"`
	UrlPathPrefix string `mapstructure:"url-path-prefix" json:"urlPathPrefix"`
	Port          int    `mapstructure:"port" json:"port"`
	BaseApi       string `mapstructure:"base-api" json:"baseApi"`
	Transaction   bool   `mapstructure:"transaction" json:"transaction"`
}

type LogsConfiguration struct {
	Level      slog.Level `mapstructure:"level" json:"level"`
	Path       string     `mapstructure:"path" json:"path"`
	MaxSize    int        `mapstructure:"max-size" json:"maxSize"`
	MaxBackups int        `mapstructure:"max-backups" json:"maxBackups"`
	MaxAge     int        `mapstructure:"max-age" json:"maxAge"`
	Compress   bool       `mapstructure:"compress" json:"compress"`
}

type AuthConfiguration struct {
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
}

type RAGConfiguration struct {
	Type    string               `mapstructure:"type" json:"type"`
	Redis   RedisConfiguration   `mapstructure:"redis" json:"redis"`
	Chromem ChromemConfiguration `mapstructure:"chromem" json:"chromem"`
	Milvus  MilvusConfiguration  `mapstructure:"milvus" json:"milvus"`
}

type RedisConfiguration struct {
	Addr   string `mapstructure:"addr" json:"addr"`
	Prefix string `mapstructure:"prefix" json:"prefix"`
}

type ChromemConfiguration struct {
	Path       string `mapstructure:"path" json:"path"`
	Collection string `mapstructure:"collection" json:"collection"`
}

type MilvusConfiguration struct {
	Addr       string `mapstructure:"addr" json:"addr"`
	Username   string `mapstructure:"username" json:"username"`
	Password   string `mapstructure:"password" json:"password"`
	Collection string `mapstructure:"collection" json:"collection"`
}

type AiModelConfiguration struct {
	EmbeddingModel ChatConfiguration `mapstructure:"embedding-model" json:"embeddingModel"`
	ChatModel      ChatConfiguration `mapstructure:"chat-model" json:"chatModel"`
}
type ChatConfiguration struct {
	BaseURL string `mapstructure:"base-url" json:"baseURL"`
	APIKey  string `mapstructure:"api-key" json:"apiKey"`
	Model   string `mapstructure:"model" json:"model"`
}
