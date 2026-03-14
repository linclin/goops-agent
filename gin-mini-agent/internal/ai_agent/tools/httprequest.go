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

package tools

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/tool"

	httprequest_delete "github.com/cloudwego/eino-ext/components/tool/httprequest/delete"
	httprequest_get "github.com/cloudwego/eino-ext/components/tool/httprequest/get"
	httprequest_post "github.com/cloudwego/eino-ext/components/tool/httprequest/post"
	httprequest_put "github.com/cloudwego/eino-ext/components/tool/httprequest/put"
)

// HTTPRequestToolConfig HTTP 请求工具配置
//
// 可配置超时时间和默认请求头。
type HTTPRequestToolConfig struct {
	// Timeout 请求超时时间（秒）
	// 默认: 30 秒
	Timeout int

	// DefaultHeaders 默认请求头
	// 每次请求都会携带这些请求头
	DefaultHeaders map[string]string

	// HttpClient 自定义 HTTP 客户端
	// 如果不提供，将使用默认客户端
	HttpClient *http.Client
}

// defaultHTTPRequestToolConfig 创建默认配置
//
// 返回默认的工具配置实例。
func defaultHTTPRequestToolConfig() *HTTPRequestToolConfig {
	return &HTTPRequestToolConfig{
		Timeout:        30,
		DefaultHeaders: make(map[string]string),
	}
}

// NewHTTPGetTool 创建 HTTP GET 请求工具
//
// 该函数创建一个用于发送 HTTP GET 请求的工具。
// 使用官方 eino-ext/components/tool/httprequest/get 库实现。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - config: 工具配置，可选
//
// 返回:
//   - tool.InvokableTool: 工具实例
//   - error: 创建过程中的错误
func NewHTTPGetTool(ctx context.Context, config *HTTPRequestToolConfig) (tool.InvokableTool, error) {
	if config == nil {
		config = defaultHTTPRequestToolConfig()
	}

	if config.Timeout <= 0 {
		config.Timeout = 30
	}

	if config.DefaultHeaders == nil {
		config.DefaultHeaders = make(map[string]string)
	}

	if config.HttpClient == nil {
		config.HttpClient = &http.Client{
			Timeout:   time.Duration(config.Timeout) * time.Second,
			Transport: &http.Transport{},
		}
	}

	getConfig := &httprequest_get.Config{
		ToolName:   "request_get",
		ToolDesc:   "向指定 URL 发送 GET 请求，获取网页内容或 API 数据。输入应为完整的 URL 地址。",
		Headers:    config.DefaultHeaders,
		HttpClient: config.HttpClient,
	}

	return httprequest_get.NewTool(ctx, getConfig)
}

// NewHTTPPostTool 创建 HTTP POST 请求工具
//
// 该函数创建一个用于发送 HTTP POST 请求的工具。
// 使用官方 eino-ext/components/tool/httprequest/post 库实现。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - config: 工具配置，可选
//
// 返回:
//   - tool.InvokableTool: 工具实例
//   - error: 创建过程中的错误
func NewHTTPPostTool(ctx context.Context, config *HTTPRequestToolConfig) (tool.InvokableTool, error) {
	if config == nil {
		config = defaultHTTPRequestToolConfig()
	}

	if config.Timeout <= 0 {
		config.Timeout = 30
	}

	if config.DefaultHeaders == nil {
		config.DefaultHeaders = make(map[string]string)
	}

	if config.HttpClient == nil {
		config.HttpClient = &http.Client{
			Timeout:   time.Duration(config.Timeout) * time.Second,
			Transport: &http.Transport{},
		}
	}

	postConfig := &httprequest_post.Config{
		ToolName: "request_post",
		ToolDesc: `向指定 URL 发送 POST 请求，用于提交数据或创建资源。
输入应为 JSON 字符串，包含 "url" 和 "body" 两个键。
"url" 的值应为字符串，"body" 的值应为要发送的数据。
注意：JSON 字符串中的字符串必须使用双引号。`,
		Headers:    config.DefaultHeaders,
		HttpClient: config.HttpClient,
	}

	return httprequest_post.NewTool(ctx, postConfig)
}

// NewHTTPPutTool 创建 HTTP PUT 请求工具
//
// 该函数创建一个用于发送 HTTP PUT 请求的工具。
// 使用官方 eino-ext/components/tool/httprequest/put 库实现。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - config: 工具配置，可选
//
// 返回:
//   - tool.InvokableTool: 工具实例
//   - error: 创建过程中的错误
func NewHTTPPutTool(ctx context.Context, config *HTTPRequestToolConfig) (tool.InvokableTool, error) {
	if config == nil {
		config = defaultHTTPRequestToolConfig()
	}

	if config.Timeout <= 0 {
		config.Timeout = 30
	}

	if config.DefaultHeaders == nil {
		config.DefaultHeaders = make(map[string]string)
	}

	if config.HttpClient == nil {
		config.HttpClient = &http.Client{
			Timeout:   time.Duration(config.Timeout) * time.Second,
			Transport: &http.Transport{},
		}
	}

	putConfig := &httprequest_put.Config{
		ToolName: "request_put",
		ToolDesc: `向指定 URL 发送 PUT 请求，用于更新资源。
输入应为 JSON 字符串，包含 "url" 和 "body" 两个键。
"url" 的值应为字符串，"body" 的值应为要发送的数据。
注意：JSON 字符串中的字符串必须使用双引号。`,
		Headers:    config.DefaultHeaders,
		HttpClient: config.HttpClient,
	}

	return httprequest_put.NewTool(ctx, putConfig)
}

// NewHTTPDeleteTool 创建 HTTP DELETE 请求工具
//
// 该函数创建一个用于发送 HTTP DELETE 请求的工具。
// 使用官方 eino-ext/components/tool/httprequest/delete 库实现。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - config: 工具配置，可选
//
// 返回:
//   - tool.InvokableTool: 工具实例
//   - error: 创建过程中的错误
func NewHTTPDeleteTool(ctx context.Context, config *HTTPRequestToolConfig) (tool.InvokableTool, error) {
	if config == nil {
		config = defaultHTTPRequestToolConfig()
	}

	if config.Timeout <= 0 {
		config.Timeout = 30
	}

	if config.DefaultHeaders == nil {
		config.DefaultHeaders = make(map[string]string)
	}

	if config.HttpClient == nil {
		config.HttpClient = &http.Client{
			Timeout:   time.Duration(config.Timeout) * time.Second,
			Transport: &http.Transport{},
		}
	}

	deleteConfig := &httprequest_delete.Config{
		ToolName:   "request_delete",
		ToolDesc:   "向指定 URL 发送 DELETE 请求，用于删除资源。输入应为完整的 URL 地址。",
		Headers:    config.DefaultHeaders,
		HttpClient: config.HttpClient,
	}

	return httprequest_delete.NewTool(ctx, deleteConfig)
}
