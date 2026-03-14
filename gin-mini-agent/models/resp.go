// Package models 提供数据模型和响应工具
//
// 该包定义了应用程序的数据模型和 HTTP 响应工具函数。
// 主要功能包括：
//   - 统一的 HTTP 响应格式
//   - 标准错误码定义
//   - 响应辅助函数
//
// 响应格式:
//
//	{
//	    "request_id": "abc123",
//	    "success": true,
//	    "data": {...},
//	    "msg": "操作成功",
//	    "total": 0
//	}
package models

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// Resp HTTP 请求响应体
//
// 该结构体定义了统一的 HTTP 响应格式。
// 所有 API 接口都应该使用此格式返回数据。
//
// 字段说明:
//   - RequestId: 请求唯一标识，用于追踪和调试
//   - Success: 请求是否成功
//   - Data: 响应数据，可以是任意类型
//   - Msg: 响应消息，成功或失败的描述
//   - Total: 数据总数，用于分页场景
//
// 使用示例:
//
//	// 成功响应
//	{
//	    "request_id": "abc123",
//	    "success": true,
//	    "data": {"name": "张三"},
//	    "msg": "操作成功",
//	    "total": 0
//	}
//
//	// 失败响应
//	{
//	    "request_id": "abc123",
//	    "success": false,
//	    "data": null,
//	    "msg": "参数错误",
//	    "total": 0
//	}
type Resp struct {
	// RequestId 请求 ID
	// 从中间件注入的请求唯一标识
	RequestId string `json:"request_id"`

	// Success 请求是否成功
	// true: 请求成功
	// false: 请求失败
	Success bool `json:"success"`

	// Data 响应数据
	// 可以是对象、数组或基本类型
	Data interface{} `json:"data"`

	// Msg 响应消息
	// 成功或失败的描述信息
	Msg string `json:"msg"`

	// Total 数据总数
	// 用于分页场景，表示总记录数
	Total int64 `json:"total"`
}

// HTTP 状态码常量
const (
	// Ok 成功状态码
	Ok = 200

	// NotOk 失败状态码
	NotOk = 405

	// Unauthorized 未授权状态码
	Unauthorized = 401

	// Forbidden 禁止访问状态码
	Forbidden = 403

	// InternalServerError 服务器内部错误状态码
	InternalServerError = 500
)

// HTTP 状态消息常量
const (
	// OkMsg 成功消息
	OkMsg = "操作成功"

	// NotOkMsg 失败消息
	NotOkMsg = "操作失败"

	// UnauthorizedMsg 未授权消息
	UnauthorizedMsg = "登录过期, 需要重新登录"

	// LoginCheckErrorMsg 登录错误消息
	LoginCheckErrorMsg = "用户名或密码错误"

	// ForbiddenMsg 禁止访问消息
	ForbiddenMsg = "无权访问该资源, 请联系网站管理员授权"

	// InternalServerErrorMsg 服务器错误消息
	InternalServerErrorMsg = "服务器内部错误"
)

// CustomError 自定义错误码与错误信息映射
//
// 该映射定义了错误码与错误消息的对应关系。
// 用于快速获取错误消息。
//
// 使用示例:
//
//	msg := CustomError[InternalServerError] // "服务器内部错误"
var CustomError = map[int]string{
	Ok:                  OkMsg,
	NotOk:               NotOkMsg,
	Unauthorized:        UnauthorizedMsg,
	Forbidden:           ForbiddenMsg,
	InternalServerError: InternalServerErrorMsg,
}

// 响应状态常量
const (
	// ERROR 失败状态
	ERROR = false

	// SUCCESS 成功状态
	SUCCESS = true
)

// EmptyArray 空数组
//
// 用于返回空数组时使用，避免返回 null。
// 前端处理时更方便。
var EmptyArray = []interface{}{}

// Result 通用响应方法
//
// 该方法是所有响应函数的基础，用于构造和发送 HTTP 响应。
//
// 参数:
//   - success: 请求是否成功
//   - data: 响应数据
//   - msg: 响应消息
//   - total: 数据总数
//   - c: Gin 上下文
//
// 使用示例:
//
//	Result(SUCCESS, user, "获取成功", 0, c)
//	Result(ERROR, nil, "参数错误", 0, c)
func Result(success bool, data interface{}, msg string, total int64, c *gin.Context) {
	// 从上下文获取请求 ID
	requestId, _ := c.Get("RequestId")

	// 发送 JSON 响应
	c.JSON(http.StatusOK, Resp{
		cast.ToString(requestId),
		success,
		data,
		msg,
		total,
	})
}

// OkResult 成功响应（无数据）
//
// 该方法返回一个成功的空响应。
//
// 参数:
//   - c: Gin 上下文
//
// 使用示例:
//
//	OkResult(c)
//	// 响应: {"request_id":"xxx","success":true,"data":{},"msg":"操作成功","total":0}
func OkResult(c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, "操作成功", 0, c)
}

// OkWithMessage 成功响应（带消息）
//
// 该方法返回一个带自定义消息的成功响应。
//
// 参数:
//   - message: 响应消息
//   - c: Gin 上下文
//
// 使用示例:
//
//	OkWithMessage("创建成功", c)
//	// 响应: {"request_id":"xxx","success":true,"data":{},"msg":"创建成功","total":0}
func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, message, 0, c)
}

// OkWithData 成功响应（带数据）
//
// 该方法返回一个带数据的成功响应。
//
// 参数:
//   - data: 响应数据
//   - c: Gin 上下文
//
// 使用示例:
//
//	OkWithData(user, c)
//	// 响应: {"request_id":"xxx","success":true,"data":{"name":"张三"},"msg":"操作成功","total":0}
func OkWithData(data interface{}, c *gin.Context) {
	Result(SUCCESS, data, "操作成功", 0, c)
}

// OkWithDataList 成功响应（带数据列表和总数）
//
// 该方法返回一个带数据列表和总数的成功响应，用于分页场景。
//
// 参数:
//   - data: 数据列表
//   - total: 数据总数
//   - c: Gin 上下文
//
// 使用示例:
//
//	OkWithDataList(users, 100, c)
//	// 响应: {"request_id":"xxx","success":true,"data":[...],"msg":"操作成功","total":100}
func OkWithDataList(data interface{}, total int64, c *gin.Context) {
	Result(SUCCESS, data, "操作成功", total, c)
}

// OkWithDetailed 成功响应（带数据和自定义消息）
//
// 该方法返回一个带数据和自定义消息的成功响应。
//
// 参数:
//   - data: 响应数据
//   - message: 响应消息
//   - c: Gin 上下文
//
// 使用示例:
//
//	OkWithDetailed(user, "用户信息获取成功", c)
//	// 响应: {"request_id":"xxx","success":true,"data":{"name":"张三"},"msg":"用户信息获取成功","total":0}
func OkWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(SUCCESS, data, message, 0, c)
}

// FailResult 失败响应（无数据）
//
// 该方法返回一个失败的空响应。
//
// 参数:
//   - c: Gin 上下文
//
// 使用示例:
//
//	FailResult(c)
//	// 响应: {"request_id":"xxx","success":false,"data":{},"msg":"操作失败","total":0}
func FailResult(c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, "操作失败", 0, c)
}

// FailWithMessage 失败响应（带消息）
//
// 该方法返回一个带自定义消息的失败响应。
//
// 参数:
//   - message: 错误消息
//   - c: Gin 上下文
//
// 使用示例:
//
//	FailWithMessage("参数错误", c)
//	// 响应: {"request_id":"xxx","success":false,"data":{},"msg":"参数错误","total":0}
func FailWithMessage(message string, c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, message, 0, c)
}

// FailWithDetailed 失败响应（带数据和消息）
//
// 该方法返回一个带数据和消息的失败响应。
//
// 参数:
//   - data: 错误详情数据
//   - message: 错误消息
//   - c: Gin 上下文
//
// 使用示例:
//
//	FailWithDetailed(errors, "表单验证失败", c)
//	// 响应: {"request_id":"xxx","success":false,"data":{"field":"name"},"msg":"表单验证失败","total":0}
func FailWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(ERROR, data, message, 0, c)
}
