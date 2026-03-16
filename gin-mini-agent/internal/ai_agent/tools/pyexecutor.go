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
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// PyExecutorToolConfig Python 执行器工具配置
//
// 可配置 Python 解释器路径。
type PyExecutorToolConfig struct {
	// Command Python 解释器命令
	// 默认: Windows 使用 python，其他系统使用 python3
	Command string
}

// defaultPyExecutorToolConfig 创建默认配置
//
// 根据操作系统选择默认的 Python 解释器。
//
// 返回:
//   - *PyExecutorToolConfig: 配置实例
func defaultPyExecutorToolConfig() *PyExecutorToolConfig {
	command := "python3"
	if runtime.GOOS == "windows" {
		command = "python"
	}
	return &PyExecutorToolConfig{
		Command: command,
	}
}

// PyExecutorToolImpl Python 执行器工具实现
//
// 该工具用于执行 Python 代码字符串。
type PyExecutorToolImpl struct {
	config *PyExecutorToolConfig
}

// NewPyExecutorTool 创建 Python 执行器工具实例
//
// 该函数创建一个用于执行 Python 代码的工具。
// 如果未提供配置，将使用默认配置。
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
//	tool, err := NewPyExecutorTool(ctx, nil)
//	result, err := tool.Invoke(ctx, PyExecutorReq{Code: "print('Hello, World!')"})
func NewPyExecutorTool(ctx context.Context, config *PyExecutorToolConfig) (tool.BaseTool, error) {
	if config == nil {
		config = defaultPyExecutorToolConfig()
	}

	t := &PyExecutorToolImpl{config: config}
	return t.ToEinoTool()
}

// ToEinoTool 转换为 Eino 工具接口
func (p *PyExecutorToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	return utils.InferTool("python_execute", "执行 Python 代码字符串。注意：只有 print 输出可见，函数返回值不会被捕获，请使用 print 语句查看结果。", p.Invoke)
}

// Invoke 执行 Python 代码
func (p *PyExecutorToolImpl) Invoke(ctx context.Context, req PyExecutorReq) (PyExecutorRes, error) {
	slog.InfoContext(ctx, "[python_execute] 工具调用开始", "code_length", len(req.Code))

	if req.Code == "" {
		slog.WarnContext(ctx, "[python_execute] 代码为空")
		return PyExecutorRes{
			Error: "Python 代码不能为空",
		}, nil
	}

	if !checkPythonAvailable(p.config.Command) {
		slog.ErrorContext(ctx, "[python_execute] Python 解释器不可用")
		return PyExecutorRes{
			Error: "Python 解释器不可用。请确保已安装 Python 并添加到系统 PATH 环境变量中。",
		}, nil
	}

	tmpFile, err := os.CreateTemp("", "python_*.py")
	if err != nil {
		slog.ErrorContext(ctx, "[python_execute] 创建临时文件失败", "error", err)
		return PyExecutorRes{
			Error: fmt.Sprintf("创建临时文件失败: %v", err),
		}, nil
	}
	tmpPath := tmpFile.Name()

	defer os.Remove(tmpPath)

	_, err = tmpFile.WriteString(req.Code)
	if err != nil {
		tmpFile.Close()
		slog.ErrorContext(ctx, "[python_execute] 写入代码失败", "error", err)
		return PyExecutorRes{
			Error: fmt.Sprintf("写入代码失败: %v", err),
		}, nil
	}
	tmpFile.Close()

	slog.DebugContext(ctx, "[python_execute] 开始执行Python代码", "tmp_file", tmpPath)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, p.config.Command, tmpPath)
	} else {
		cmd = exec.CommandContext(ctx, p.config.Command, tmpPath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.ErrorContext(ctx, "[python_execute] 代码执行失败", "error", err, "output", string(output))
		return PyExecutorRes{
			Error: fmt.Sprintf("Python 代码执行失败:\n%s", string(output)),
		}, nil
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		slog.InfoContext(ctx, "[python_execute] 执行成功，无输出")
		return PyExecutorRes{
			Output: "代码执行成功，但没有输出。请使用 print() 语句输出结果。",
		}, nil
	}

	slog.InfoContext(ctx, "[python_execute] 执行成功", "output_length", len(result))
	return PyExecutorRes{
		Output: result,
	}, nil
}

// checkPythonAvailable 检查 Python 解释器是否可用
//
// 该函数尝试运行 Python 解释器来验证其是否正确安装。
//
// 参数:
//   - command: Python 解释器命令
//
// 返回:
//   - bool: true 表示可用，false 表示不可用
func checkPythonAvailable(command string) bool {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(command, "--version")
	} else {
		cmd = exec.Command(command, "--version")
	}

	err := cmd.Run()
	return err == nil
}

// PyExecutorReq Python 执行请求结构体
//
// 定义了 Python 执行工具的输入参数。
type PyExecutorReq struct {
	// Code 要执行的 Python 代码
	// 示例:
	//   - print('Hello, World!')
	//   - import math; print(math.sqrt(16))
	Code string `json:"code" jsonschema_description:"要执行的 Python 代码字符串"`
}

// PyExecutorRes Python 执行响应结构体
//
// 定义了 Python 执行工具的输出结果。
type PyExecutorRes struct {
	// Output 执行输出
	// 包含 print 语句的输出内容
	Output string `json:"output,omitempty" jsonschema_description:"Python 代码执行输出"`

	// Error 错误信息
	// 当执行失败时包含错误描述
	Error string `json:"error,omitempty" jsonschema_description:"执行错误信息"`
}
