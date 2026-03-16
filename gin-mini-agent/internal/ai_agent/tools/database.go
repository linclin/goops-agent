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
//   - database: 数据库操作工具，支持 MySQL 和 PostgreSQL
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

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseToolImpl 数据库工具实现
//
// 该工具用于操作数据库，支持 MySQL 和 PostgreSQL。
// 支持执行 SQL 查询、插入、更新、删除等操作。
//
// 使用场景:
//   - 用户要求查询数据库中的数据
//   - 用户要求执行数据库操作
//   - 用户要求管理数据库表结构
//
// 示例:
//   - 查询数据: SELECT * FROM users WHERE id = 1
//   - 插入数据: INSERT INTO users (name, email) VALUES ('John', 'john@example.com')
//   - 更新数据: UPDATE users SET name = 'Jane' WHERE id = 1
//   - 删除数据: DELETE FROM users WHERE id = 1
type DatabaseToolImpl struct {
	// config 工具配置
	config *DatabaseToolConfig
	// db 数据库连接
	db *gorm.DB
}

// DatabaseToolConfig 数据库工具配置
//
// 定义了数据库工具的配置选项，包括数据库类型、连接信息等。
type DatabaseToolConfig struct {
	// Type 数据库类型
	// 支持: mysql, postgres
	Type string `json:"type" jsonschema_description:"数据库类型，支持 mysql, postgres"`

	// Host 数据库主机地址
	Host string `json:"host" jsonschema_description:"数据库主机地址"`

	// Port 数据库端口
	Port int `json:"port" jsonschema_description:"数据库端口"`

	// User 数据库用户名
	User string `json:"user" jsonschema_description:"数据库用户名"`

	// Password 数据库密码
	Password string `json:"password" jsonschema_description:"数据库密码"`

	// Database 数据库名称
	Database string `json:"database" jsonschema_description:"数据库名称"`

	// Charset 字符集
	Charset string `json:"charset" jsonschema_description:"字符集"`

	// ParseTime 是否解析时间
	ParseTime bool `json:"parse_time" jsonschema_description:"是否解析时间"`

	// Loc 时区
	Loc string `json:"loc" jsonschema_description:"时区"`
}

// defaultDatabaseToolConfig 创建默认配置
//
// 返回默认的工具配置实例。
//
// 返回:
//   - *DatabaseToolConfig: 配置实例
func defaultDatabaseToolConfig() *DatabaseToolConfig {
	return &DatabaseToolConfig{
		Type:      "mysql",
		Host:      "localhost",
		Port:      3306,
		User:      "root",
		Password:  "",
		Database:  "test",
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}
}

// NewDatabaseTool 创建数据库工具实例
//
// 该函数创建一个用于操作数据库的工具。
// 如果未提供配置，将使用默认配置，但默认配置仅作为 fallback，
// 实际使用时应通过请求参数提供完整的数据库连接信息。
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
//	tool, err := NewDatabaseTool(ctx, nil)
//	// 调用时通过请求参数提供连接信息
func NewDatabaseTool(ctx context.Context, config *DatabaseToolConfig) (tool.BaseTool, error) {
	slog.InfoContext(ctx, "[database] 创建数据库工具")

	// 如果配置为空，使用默认配置
	if config == nil {
		config = defaultDatabaseToolConfig()
	}

	// 注意：不再在创建时连接数据库，而是在调用时根据请求参数连接
	// 这样可以让大模型在调用时提供完整的连接信息

	slog.InfoContext(ctx, "[database] 数据库工具创建成功，将在调用时根据请求参数连接数据库")

	// 创建工具实例
	t := &DatabaseToolImpl{
		config: config,
		db:     nil, // 初始化为 nil，在调用时根据请求参数连接
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
func (d *DatabaseToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	// 使用 InferTool 自动生成工具信息
	// 参数:
	//   - name: 工具名称，用于 Agent 识别
	//   - description: 工具描述，帮助 Agent 理解工具用途
	//   - invoke: 工具调用函数
	return utils.InferTool("database", "数据库操作工具，支持 MySQL 和 PostgreSQL，可执行 SQL 查询、插入、更新、删除等操作", d.Invoke)
}

// DatabaseReq 数据库请求结构体
//
// 定义了数据库工具的输入参数。
type DatabaseReq struct {
	// SQL SQL 语句
	// 要执行的 SQL 语句
	SQL string `json:"sql" jsonschema_description:"要执行的 SQL 语句"`

	// Type 数据库类型
	// 可选，覆盖配置中的数据库类型
	Type string `json:"type" jsonschema_description:"数据库类型，支持 mysql, postgres"`

	// Host 数据库主机地址
	// 可选，覆盖配置中的主机地址
	Host string `json:"host" jsonschema_description:"数据库主机地址"`

	// Port 数据库端口
	// 可选，覆盖配置中的端口
	Port int `json:"port" jsonschema_description:"数据库端口"`

	// User 数据库用户名
	// 可选，覆盖配置中的用户名
	User string `json:"user" jsonschema_description:"数据库用户名"`

	// Password 数据库密码
	// 可选，覆盖配置中的密码
	Password string `json:"password" jsonschema_description:"数据库密码"`

	// Database 数据库名称
	// 可选，覆盖配置中的数据库名称
	Database string `json:"database" jsonschema_description:"数据库名称"`
}

// DatabaseRes 数据库响应结构体
//
// 定义了数据库工具的输出结果。
type DatabaseRes struct {
	// Result 执行结果
	// 包含查询结果或执行状态
	Result string `json:"result" jsonschema_description:"执行结果，包含查询结果或执行状态"`

	// Success 是否执行成功
	// true 表示执行成功，false 表示执行失败
	Success bool `json:"success" jsonschema_description:"是否执行成功，true 表示执行成功，false 表示执行失败"`

	// RowsAffected 影响的行数
	// 对于 INSERT、UPDATE、DELETE 操作
	RowsAffected int64 `json:"rows_affected" jsonschema_description:"影响的行数，对于 INSERT、UPDATE、DELETE 操作"`

	// Error 错误信息
	// 如果执行失败，包含错误信息
	Error string `json:"error" jsonschema_description:"错误信息，如果执行失败，包含错误信息"`
}

// Invoke 执行数据库操作
//
// 该方法是工具的核心实现，负责执行指定的 SQL 语句。
// 支持通过请求参数提供完整的数据库连接信息，
// 这样大模型可以在调用时动态提供连接参数。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - req: 数据库请求，包含 SQL 语句和连接信息
//
// 返回:
//   - DatabaseRes: 执行结果，包含执行状态、结果内容等
//   - error: 执行错误
func (d *DatabaseToolImpl) Invoke(ctx context.Context, req DatabaseReq) (res DatabaseRes, err error) {
	slog.InfoContext(ctx, "[database] 调用数据库工具", "sql", req.SQL)

	// 验证 SQL 参数
	if req.SQL == "" {
		slog.WarnContext(ctx, "[database] 缺少 SQL 参数")
		res.Result = "缺少 SQL 参数"
		res.Success = false
		res.Error = "缺少 SQL 参数"
		return res, nil
	}

	// 检查是否提供了连接信息
	if req.Type == "" && req.Host == "" && req.User == "" && req.Password == "" && req.Database == "" {
		slog.WarnContext(ctx, "[database] 未提供数据库连接信息，将使用默认配置")
		// 使用默认配置
		db, err := connectDatabase(d.config)
		if err != nil {
			slog.ErrorContext(ctx, "[database] 使用默认配置连接数据库失败", "error", err)
			res.Result = "使用默认配置连接数据库失败，请在请求中提供完整的数据库连接信息"
			res.Success = false
			res.Error = err.Error()
			return res, nil
		}

		// 执行 SQL 语句
		var result interface{}
		err = db.WithContext(ctx).Raw(req.SQL).Scan(&result).Error
		if err != nil {
			slog.ErrorContext(ctx, "[database] 执行 SQL 失败", "error", err)
			res.Result = "执行 SQL 失败"
			res.Success = false
			res.Error = err.Error()
			return res, nil
		}

		// 获取影响的行数
		rowsAffected := db.RowsAffected

		// 处理执行结果
		slog.InfoContext(ctx, "[database] SQL 执行成功", "rows_affected", rowsAffected)
		res.Result = fmt.Sprintf("SQL 执行成功，影响 %d 行", rowsAffected)
		res.Success = true
		res.RowsAffected = rowsAffected

		return res, nil
	}

	// 使用请求参数中的连接信息
	tempConfig := *d.config
	if req.Type != "" {
		tempConfig.Type = req.Type
	}
	if req.Host != "" {
		tempConfig.Host = req.Host
	}
	if req.Port != 0 {
		tempConfig.Port = req.Port
	}
	if req.User != "" {
		tempConfig.User = req.User
	}
	if req.Password != "" {
		tempConfig.Password = req.Password
	}
	if req.Database != "" {
		tempConfig.Database = req.Database
	}

	// 创建连接
	db, err := connectDatabase(&tempConfig)
	if err != nil {
		slog.ErrorContext(ctx, "[database] 连接数据库失败", "error", err)
		res.Result = "连接数据库失败"
		res.Success = false
		res.Error = err.Error()
		return res, nil
	}

	// 执行 SQL 语句
	var result interface{}
	err = db.WithContext(ctx).Raw(req.SQL).Scan(&result).Error
	if err != nil {
		slog.ErrorContext(ctx, "[database] 执行 SQL 失败", "error", err)
		res.Result = "执行 SQL 失败"
		res.Success = false
		res.Error = err.Error()
		return res, nil
	}

	// 获取影响的行数
	rowsAffected := db.RowsAffected

	// 处理执行结果
	slog.InfoContext(ctx, "[database] SQL 执行成功", "rows_affected", rowsAffected)
	res.Result = fmt.Sprintf("SQL 执行成功，影响 %d 行", rowsAffected)
	res.Success = true
	res.RowsAffected = rowsAffected

	return res, nil
}

// connectDatabase 连接数据库
//
// 根据配置连接到指定的数据库。
//
// 参数:
//   - config: 数据库配置
//
// 返回:
//   - *gorm.DB: 数据库连接
//   - error: 连接错误
func connectDatabase(config *DatabaseToolConfig) (*gorm.DB, error) {
	var dsn string
	var dialector gorm.Dialector

	switch config.Type {
	case "mysql":
		// 构建 MySQL DSN
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
			config.User, config.Password, config.Host, config.Port, config.Database, config.Charset, config.ParseTime, config.Loc)
		dialector = mysql.Open(dsn)
	case "postgres":
		// 构建 PostgreSQL DSN
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.User, config.Password, config.Database)
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// 连接数据库
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}

	// 获取底层 SQL DB 并配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}
