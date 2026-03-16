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

// Package tools 提供 AI Agent 可调用的工具集合
//
// 该包定义了多种工具，扩展 AI Agent 的能力边界。
// 工具是 Agent 与外部世界交互的桥梁，允许 Agent 执行文件操作、
// 浏览器自动化、网络请求等任务。
//
// 当前可用工具:
//   - redis: Redis 操作工具，支持各种 Redis 命令
//
// 工具开发指南:
//  1. 定义工具结构体，包含配置信息
//  2. 实现 ToEinoTool 方法，返回 tool.InvokableTool
//  3. 实现 Invoke 方法，执行具体的工具逻辑
//  4. 使用 utils.InferTool 自动生成工具信息
package tools

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	redisCli "github.com/redis/go-redis/v9"
)

// RedisToolImpl Redis 工具实现
//
// 该工具用于操作 Redis 数据库，支持各种 Redis 命令。
// 支持字符串、哈希、列表、集合、有序集合等数据结构的操作。
//
// 使用场景:
//   - 用户要求操作 Redis 数据库
//   - 用户要求存储或获取缓存数据
//   - 用户要求管理 Redis 键值对
//
// 示例:
//   - 设置键值: SET key value
//   - 获取键值: GET key
//   - 哈希操作: HSET hash field value
//   - 列表操作: LPUSH list value
//   - 集合操作: SADD set member
//   - 有序集合操作: ZADD zset score member
type RedisToolImpl struct {
	// config 工具配置
	config *RedisToolConfig
	// client Redis 客户端
	client *redisCli.Client
}

// RedisToolConfig Redis 工具配置
//
// 定义了 Redis 工具的配置选项，包括连接信息等。
type RedisToolConfig struct {
	// Host Redis 主机地址
	Host string `json:"host" jsonschema_description:"Redis 主机地址"`

	// Port Redis 端口
	Port int `json:"port" jsonschema_description:"Redis 端口"`

	// Password Redis 密码
	Password string `json:"password" jsonschema_description:"Redis 密码"`

	// DB 数据库编号
	DB int `json:"db" jsonschema_description:"数据库编号"`

	// TLS 是否使用 TLS
	TLS bool `json:"tls" jsonschema_description:"是否使用 TLS"`

	// MaxRetries 最大重试次数
	MaxRetries int `json:"max_retries" jsonschema_description:"最大重试次数"`

	// DialTimeout 连接超时时间（秒）
	DialTimeout int `json:"dial_timeout" jsonschema_description:"连接超时时间（秒）"`

	// ReadTimeout 读取超时时间（秒）
	ReadTimeout int `json:"read_timeout" jsonschema_description:"读取超时时间（秒）"`

	// WriteTimeout 写入超时时间（秒）
	WriteTimeout int `json:"write_timeout" jsonschema_description:"写入超时时间（秒）"`
}

// defaultRedisToolConfig 创建默认配置
//
// 返回默认的工具配置实例。
//
// 返回:
//   - *RedisToolConfig: 配置实例
func defaultRedisToolConfig() *RedisToolConfig {
	return &RedisToolConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		TLS:          false,
		MaxRetries:   3,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
	}
}

// NewRedisTool 创建 Redis 工具实例
//
// 该函数创建一个用于操作 Redis 的工具。
// 如果未提供配置，将使用默认配置，但默认配置仅作为 fallback，
// 实际使用时应通过请求参数提供完整的 Redis 连接信息。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - config: 工具配置，可选
//
// 返回:
//   - tool.BaseTool: 工具实例
//   - error: 创建过程中的错误
//
// 使用示例:
//
//	tool, err := NewRedisTool(ctx, nil)
//	// 调用时通过请求参数提供连接信息
func NewRedisTool(ctx context.Context, config *RedisToolConfig) (tool.BaseTool, error) {
	slog.InfoContext(ctx, "[redis] 创建 Redis 工具")

	// 如果配置为空，使用默认配置
	if config == nil {
		config = defaultRedisToolConfig()
	}

	// 注意：不再在创建时连接 Redis，而是在调用时根据请求参数连接
	// 这样可以让大模型在调用时提供完整的连接信息

	slog.InfoContext(ctx, "[redis] Redis 工具创建成功，将在调用时根据请求参数连接 Redis")

	// 创建工具实例
	t := &RedisToolImpl{
		config: config,
		client: nil, // 初始化为 nil，在调用时根据请求参数连接
	}

	// 转换为 Eino 工具
	tn, err := t.ToEinoTool()
	if err != nil {
		return nil, err
	}
	return tn, nil
}

// ToEinoTool 转换为 Eino 工具接口
//
// 该方法将工具实现转换为 Eino 框架的工具接口。
// 使用 utils.InferTool 自动推断工具的参数和返回值类型。
//
// 返回:
//   - tool.InvokableTool: Eino 工具实例
//   - error: 转换错误
func (r *RedisToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	// 使用 InferTool 自动生成工具信息
	// 参数:
	//   - name: 工具名称，用于 Agent 识别
	//   - description: 工具描述，帮助 Agent 理解工具用途
	//   - invoke: 工具调用函数
	return utils.InferTool("redis", "Redis 操作工具，支持各种 Redis 命令，如 SET、GET、HSET、LPUSH、SADD、ZADD 等", r.Invoke)
}

// RedisReq Redis 请求结构体
//
// 定义了 Redis 工具的输入参数。
type RedisReq struct {
	// Command Redis 命令
	// 支持的命令: SET, GET, HSET, HGET, LPUSH, LPOP, SADD, SMEMBERS, ZADD, ZRANGE 等
	Command string `json:"command" jsonschema_description:"Redis 命令，如 SET、GET、HSET、LPUSH、SADD、ZADD 等"`

	// Key 键名
	Key string `json:"key" jsonschema_description:"键名"`

	// Value 值
	// 对于 SET 命令，是要设置的值
	// 对于 HSET 命令，是字段值
	// 对于 LPUSH 命令，是要推入的值
	// 对于 SADD 命令，是要添加的成员
	// 对于 ZADD 命令，是要添加的成员
	Value string `json:"value" jsonschema_description:"值，根据命令不同有不同含义"`

	// Field 字段名
	// 对于 HSET、HGET 等哈希命令
	Field string `json:"field" jsonschema_description:"字段名，用于哈希命令"`

	// Score 分数
	// 对于 ZADD 命令
	Score float64 `json:"score" jsonschema_description:"分数，用于有序集合命令"`

	// Count 数量
	// 对于 LPOP、RPOP 等命令
	Count int `json:"count" jsonschema_description:"数量，用于列表命令"`

	// Start 起始索引
	// 对于 ZRANGE 等命令
	Start int `json:"start" jsonschema_description:"起始索引，用于范围命令"`

	// Stop 结束索引
	// 对于 ZRANGE 等命令
	Stop int `json:"stop" jsonschema_description:"结束索引，用于范围命令"`

	// Expire 过期时间（秒）
	// 对于 SET 命令，设置键的过期时间
	Expire int `json:"expire" jsonschema_description:"过期时间（秒），用于 SET 命令"`

	// Host Redis 主机地址
	// 可选，覆盖配置中的主机地址
	Host string `json:"host" jsonschema_description:"Redis 主机地址"`

	// Port Redis 端口
	// 可选，覆盖配置中的端口
	Port int `json:"port" jsonschema_description:"Redis 端口"`

	// Password Redis 密码
	// 可选，覆盖配置中的密码
	Password string `json:"password" jsonschema_description:"Redis 密码"`

	// DB 数据库编号
	// 可选，覆盖配置中的数据库编号
	DB int `json:"db" jsonschema_description:"数据库编号"`
}

// RedisRes Redis 响应结构体
//
// 定义了 Redis 工具的输出结果。
type RedisRes struct {
	// Result 执行结果
	// 包含命令执行的结果
	Result string `json:"result" jsonschema_description:"执行结果，包含命令执行的结果"`

	// Success 是否执行成功
	// true 表示执行成功，false 表示执行失败
	Success bool `json:"success" jsonschema_description:"是否执行成功，true 表示执行成功，false 表示执行失败"`

	// Error 错误信息
	// 如果执行失败，包含错误信息
	Error string `json:"error" jsonschema_description:"错误信息，如果执行失败，包含错误信息"`
}

// Invoke 执行 Redis 命令
//
// 该方法是工具的核心实现，负责执行指定的 Redis 命令。
// 支持通过请求参数提供完整的 Redis 连接信息，
// 这样大模型可以在调用时动态提供连接参数。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - req: Redis 请求，包含命令和参数
//
// 返回:
//   - RedisRes: 执行结果，包含执行状态、结果内容等
//   - error: 执行错误
func (r *RedisToolImpl) Invoke(ctx context.Context, req RedisReq) (res RedisRes, err error) {
	slog.InfoContext(ctx, "[redis] 调用 Redis 工具", "command", req.Command, "key", req.Key)

	// 验证命令参数
	if req.Command == "" {
		slog.WarnContext(ctx, "[redis] 缺少命令参数")
		res.Result = "缺少命令参数"
		res.Success = false
		res.Error = "缺少命令参数"
		return res, nil
	}

	// 检查是否提供了连接信息
	if req.Host == "" && req.Port == 0 && req.Password == "" && req.DB == 0 {
		slog.WarnContext(ctx, "[redis] 未提供 Redis 连接信息，将使用默认配置")
		// 使用默认配置
		client := redisCli.NewClient(&redisCli.Options{
			Addr:         fmt.Sprintf("%s:%d", r.config.Host, r.config.Port),
			Password:     r.config.Password,
			DB:           r.config.DB,
			TLSConfig:    nil,
			MaxRetries:   r.config.MaxRetries,
			DialTimeout:  time.Duration(r.config.DialTimeout) * time.Second,
			ReadTimeout:  time.Duration(r.config.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(r.config.WriteTimeout) * time.Second,
		})

		// 测试连接
		_, err := client.Ping(ctx).Result()
		if err != nil {
			slog.ErrorContext(ctx, "[redis] 使用默认配置连接 Redis 失败", "error", err)
			res.Result = "使用默认配置连接 Redis 失败，请在请求中提供完整的 Redis 连接信息"
			res.Success = false
			res.Error = err.Error()
			return res, nil
		}

		// 执行 Redis 命令
		var result interface{}
		var errMsg error

		switch req.Command {
		case "SET":
			if req.Key == "" || req.Value == "" {
				res.Result = "SET 命令需要 key 和 value 参数"
				res.Success = false
				res.Error = "SET 命令需要 key 和 value 参数"
				return res, nil
			}
			if req.Expire > 0 {
				result, errMsg = client.Set(ctx, req.Key, req.Value, time.Duration(req.Expire)*time.Second).Result()
			} else {
				result, errMsg = client.Set(ctx, req.Key, req.Value, 0).Result()
			}
		case "GET":
			if req.Key == "" {
				res.Result = "GET 命令需要 key 参数"
				res.Success = false
				res.Error = "GET 命令需要 key 参数"
				return res, nil
			}
			result, errMsg = client.Get(ctx, req.Key).Result()
		case "HSET":
			if req.Key == "" || req.Field == "" || req.Value == "" {
				res.Result = "HSET 命令需要 key、field 和 value 参数"
				res.Success = false
				res.Error = "HSET 命令需要 key、field 和 value 参数"
				return res, nil
			}
			result, errMsg = client.HSet(ctx, req.Key, req.Field, req.Value).Result()
		case "HGET":
			if req.Key == "" || req.Field == "" {
				res.Result = "HGET 命令需要 key 和 field 参数"
				res.Success = false
				res.Error = "HGET 命令需要 key 和 field 参数"
				return res, nil
			}
			result, errMsg = client.HGet(ctx, req.Key, req.Field).Result()
		case "LPUSH":
			if req.Key == "" || req.Value == "" {
				res.Result = "LPUSH 命令需要 key 和 value 参数"
				res.Success = false
				res.Error = "LPUSH 命令需要 key 和 value 参数"
				return res, nil
			}
			result, errMsg = client.LPush(ctx, req.Key, req.Value).Result()
		case "LPOP":
			if req.Key == "" {
				res.Result = "LPOP 命令需要 key 参数"
				res.Success = false
				res.Error = "LPOP 命令需要 key 参数"
				return res, nil
			}
			if req.Count > 0 {
				result, errMsg = client.LPopCount(ctx, req.Key, req.Count).Result()
			} else {
				result, errMsg = client.LPop(ctx, req.Key).Result()
			}
		case "SADD":
			if req.Key == "" || req.Value == "" {
				res.Result = "SADD 命令需要 key 和 value 参数"
				res.Success = false
				res.Error = "SADD 命令需要 key 和 value 参数"
				return res, nil
			}
			result, errMsg = client.SAdd(ctx, req.Key, req.Value).Result()
		case "SMEMBERS":
			if req.Key == "" {
				res.Result = "SMEMBERS 命令需要 key 参数"
				res.Success = false
				res.Error = "SMEMBERS 命令需要 key 参数"
				return res, nil
			}
			result, errMsg = client.SMembers(ctx, req.Key).Result()
		case "ZADD":
			if req.Key == "" || req.Value == "" {
				res.Result = "ZADD 命令需要 key、value 和 score 参数"
				res.Success = false
				res.Error = "ZADD 命令需要 key、value 和 score 参数"
				return res, nil
			}
			result, errMsg = client.ZAdd(ctx, req.Key, redisCli.Z{Score: req.Score, Member: req.Value}).Result()
		case "ZRANGE":
			if req.Key == "" {
				res.Result = "ZRANGE 命令需要 key 参数"
				res.Success = false
				res.Error = "ZRANGE 命令需要 key 参数"
				return res, nil
			}
			result, errMsg = client.ZRange(ctx, req.Key, int64(req.Start), int64(req.Stop)).Result()
		case "DEL":
			if req.Key == "" {
				res.Result = "DEL 命令需要 key 参数"
				res.Success = false
				res.Error = "DEL 命令需要 key 参数"
				return res, nil
			}
			result, errMsg = client.Del(ctx, req.Key).Result()
		case "EXISTS":
			if req.Key == "" {
				res.Result = "EXISTS 命令需要 key 参数"
				res.Success = false
				res.Error = "EXISTS 命令需要 key 参数"
				return res, nil
			}
			result, errMsg = client.Exists(ctx, req.Key).Result()
		case "EXPIRE":
			if req.Key == "" || req.Expire <= 0 {
				res.Result = "EXPIRE 命令需要 key 和 expire 参数"
				res.Success = false
				res.Error = "EXPIRE 命令需要 key 和 expire 参数"
				return res, nil
			}
			result, errMsg = client.Expire(ctx, req.Key, time.Duration(req.Expire)*time.Second).Result()
		case "TTL":
			if req.Key == "" {
				res.Result = "TTL 命令需要 key 参数"
				res.Success = false
				res.Error = "TTL 命令需要 key 参数"
				return res, nil
			}
			result, errMsg = client.TTL(ctx, req.Key).Result()
		default:
			res.Result = fmt.Sprintf("不支持的命令: %s", req.Command)
			res.Success = false
			res.Error = fmt.Sprintf("不支持的命令: %s", req.Command)
			return res, nil
		}

		// 处理执行结果
		if errMsg != nil {
			slog.ErrorContext(ctx, "[redis] 执行命令失败", "error", errMsg)
			res.Result = "执行命令失败"
			res.Success = false
			res.Error = errMsg.Error()
			return res, nil
		}

		// 处理执行结果
		slog.InfoContext(ctx, "[redis] 命令执行成功", "result", result)
		res.Result = fmt.Sprintf("%v", result)
		res.Success = true

		return res, nil
	}

	// 使用请求参数中的连接信息
	tempConfig := *r.config
	if req.Host != "" {
		tempConfig.Host = req.Host
	}
	if req.Port != 0 {
		tempConfig.Port = req.Port
	}
	if req.Password != "" {
		tempConfig.Password = req.Password
	}
	if req.DB != 0 {
		tempConfig.DB = req.DB
	}

	// 创建客户端
	client := redisCli.NewClient(&redisCli.Options{
		Addr:         fmt.Sprintf("%s:%d", tempConfig.Host, tempConfig.Port),
		Password:     tempConfig.Password,
		DB:           tempConfig.DB,
		TLSConfig:    nil,
		MaxRetries:   tempConfig.MaxRetries,
		DialTimeout:  time.Duration(tempConfig.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(tempConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(tempConfig.WriteTimeout) * time.Second,
	})

	// 测试连接
	_, err = client.Ping(ctx).Result()
	if err != nil {
		slog.ErrorContext(ctx, "[redis] 连接 Redis 失败", "error", err)
		res.Result = "连接 Redis 失败"
		res.Success = false
		res.Error = err.Error()
		return res, nil
	}

	// 执行 Redis 命令
	var result interface{}
	var errMsg error

	switch req.Command {
	case "SET":
		if req.Key == "" || req.Value == "" {
			res.Result = "SET 命令需要 key 和 value 参数"
			res.Success = false
			res.Error = "SET 命令需要 key 和 value 参数"
			return res, nil
		}
		if req.Expire > 0 {
			result, errMsg = client.Set(ctx, req.Key, req.Value, time.Duration(req.Expire)*time.Second).Result()
		} else {
			result, errMsg = client.Set(ctx, req.Key, req.Value, 0).Result()
		}
	case "GET":
		if req.Key == "" {
			res.Result = "GET 命令需要 key 参数"
			res.Success = false
			res.Error = "GET 命令需要 key 参数"
			return res, nil
		}
		result, errMsg = client.Get(ctx, req.Key).Result()
	case "HSET":
		if req.Key == "" || req.Field == "" || req.Value == "" {
			res.Result = "HSET 命令需要 key、field 和 value 参数"
			res.Success = false
			res.Error = "HSET 命令需要 key、field 和 value 参数"
			return res, nil
		}
		result, errMsg = client.HSet(ctx, req.Key, req.Field, req.Value).Result()
	case "HGET":
		if req.Key == "" || req.Field == "" {
			res.Result = "HGET 命令需要 key 和 field 参数"
			res.Success = false
			res.Error = "HGET 命令需要 key 和 field 参数"
			return res, nil
		}
		result, errMsg = client.HGet(ctx, req.Key, req.Field).Result()
	case "LPUSH":
		if req.Key == "" || req.Value == "" {
			res.Result = "LPUSH 命令需要 key 和 value 参数"
			res.Success = false
			res.Error = "LPUSH 命令需要 key 和 value 参数"
			return res, nil
		}
		result, errMsg = client.LPush(ctx, req.Key, req.Value).Result()
	case "LPOP":
		if req.Key == "" {
			res.Result = "LPOP 命令需要 key 参数"
			res.Success = false
			res.Error = "LPOP 命令需要 key 参数"
			return res, nil
		}
		if req.Count > 0 {
			result, errMsg = client.LPopCount(ctx, req.Key, req.Count).Result()
		} else {
			result, errMsg = client.LPop(ctx, req.Key).Result()
		}
	case "SADD":
		if req.Key == "" || req.Value == "" {
			res.Result = "SADD 命令需要 key 和 value 参数"
			res.Success = false
			res.Error = "SADD 命令需要 key 和 value 参数"
			return res, nil
		}
		result, errMsg = client.SAdd(ctx, req.Key, req.Value).Result()
	case "SMEMBERS":
		if req.Key == "" {
			res.Result = "SMEMBERS 命令需要 key 参数"
			res.Success = false
			res.Error = "SMEMBERS 命令需要 key 参数"
			return res, nil
		}
		result, errMsg = client.SMembers(ctx, req.Key).Result()
	case "ZADD":
		if req.Key == "" || req.Value == "" {
			res.Result = "ZADD 命令需要 key、value 和 score 参数"
			res.Success = false
			res.Error = "ZADD 命令需要 key、value 和 score 参数"
			return res, nil
		}
		result, errMsg = client.ZAdd(ctx, req.Key, redisCli.Z{Score: req.Score, Member: req.Value}).Result()
	case "ZRANGE":
		if req.Key == "" {
			res.Result = "ZRANGE 命令需要 key 参数"
			res.Success = false
			res.Error = "ZRANGE 命令需要 key 参数"
			return res, nil
		}
		result, errMsg = client.ZRange(ctx, req.Key, int64(req.Start), int64(req.Stop)).Result()
	case "DEL":
		if req.Key == "" {
			res.Result = "DEL 命令需要 key 参数"
			res.Success = false
			res.Error = "DEL 命令需要 key 参数"
			return res, nil
		}
		result, errMsg = client.Del(ctx, req.Key).Result()
	case "EXISTS":
		if req.Key == "" {
			res.Result = "EXISTS 命令需要 key 参数"
			res.Success = false
			res.Error = "EXISTS 命令需要 key 参数"
			return res, nil
		}
		result, errMsg = client.Exists(ctx, req.Key).Result()
	case "EXPIRE":
		if req.Key == "" || req.Expire <= 0 {
			res.Result = "EXPIRE 命令需要 key 和 expire 参数"
			res.Success = false
			res.Error = "EXPIRE 命令需要 key 和 expire 参数"
			return res, nil
		}
		result, errMsg = client.Expire(ctx, req.Key, time.Duration(req.Expire)*time.Second).Result()
	case "TTL":
		if req.Key == "" {
			res.Result = "TTL 命令需要 key 参数"
			res.Success = false
			res.Error = "TTL 命令需要 key 参数"
			return res, nil
		}
		result, errMsg = client.TTL(ctx, req.Key).Result()
	default:
		res.Result = fmt.Sprintf("不支持的命令: %s", req.Command)
		res.Success = false
		res.Error = fmt.Sprintf("不支持的命令: %s", req.Command)
		return res, nil
	}

	// 处理执行结果
	if errMsg != nil {
		slog.ErrorContext(ctx, "[redis] 执行命令失败", "error", errMsg)
		res.Result = "执行命令失败"
		res.Success = false
		res.Error = errMsg.Error()
		return res, nil
	}

	// 处理执行结果
	slog.InfoContext(ctx, "[redis] 命令执行成功", "result", result)
	res.Result = fmt.Sprintf("%v", result)
	res.Success = true

	return res, nil
}
