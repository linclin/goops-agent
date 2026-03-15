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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// FileEditorToolImpl 文件编辑器工具实现
//
// 该工具用于创建、查看和编辑文件。
// 支持字符串替换方式修改文件内容。
//
// 使用场景:
//   - 创建新文件
//   - 查看文件内容
//   - 使用字符串替换方式修改文件内容
//   - 代码编辑和修改
//
// 支持的命令:
//   - view: 查看文件内容
//   - create: 创建新文件
//   - str_replace: 字符串替换
//   - insert: 插入内容
type FileEditorToolImpl struct {
	// config 工具配置
	config *FileEditorToolConfig
}

// FileEditorToolConfig 文件编辑器工具配置
//
// 当前配置为空，保留用于未来扩展。
type FileEditorToolConfig struct {
}

// defaultFileEditorToolConfig 创建默认配置
//
// 返回默认的工具配置实例。
//
// 返回:
//   - *FileEditorToolConfig: 配置实例
func defaultFileEditorToolConfig() *FileEditorToolConfig {
	return &FileEditorToolConfig{}
}

// NewFileEditorTool 创建文件编辑器工具实例
//
// 该函数创建一个用于编辑文件的工具。
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
//	tool, err := NewFileEditorTool(ctx, nil)
//	result, err := tool.Invoke(ctx, FileEditorReq{
//	    Command: "view",
//	    Path: "/path/to/file.txt",
//	})
func NewFileEditorTool(ctx context.Context, config *FileEditorToolConfig) (tool.BaseTool, error) {
	if config == nil {
		config = defaultFileEditorToolConfig()
	}

	t := &FileEditorToolImpl{config: config}

	return t.ToEinoTool()
}

// ToEinoTool 转换为 Eino 工具接口
//
// 该方法将工具实现转换为 Eino 框架的工具接口。
// 使用 utils.InferTool 自动推断工具的参数和返回值类型。
//
// 返回:
//   - tool.InvokableTool: Eino 工具实例
//   - error: 转换错误
func (f *FileEditorToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	return utils.InferTool("str_replace_editor", "文件编辑器，支持创建、查看、编辑文件，使用字符串替换方式修改文件内容。命令: view(查看), create(创建), str_replace(替换), insert(插入)", f.Invoke)
}

// Invoke 执行文件编辑操作
//
// 该方法是工具的核心实现，负责执行文件编辑操作。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - req: 编辑请求，包含命令类型和文件路径等
//
// 返回:
//   - FileEditorRes: 操作结果
//   - error: 执行错误
func (f *FileEditorToolImpl) Invoke(ctx context.Context, req FileEditorReq) (FileEditorRes, error) {
	slog.InfoContext(ctx, "[str_replace_editor] 工具调用开始", "command", req.Command, "path", req.Path)

	if req.Path == "" {
		slog.WarnContext(ctx, "[str_replace_editor] 文件路径为空")
		return FileEditorRes{
			Error: "文件路径不能为空",
		}, nil
	}

	if req.Command == "" {
		slog.WarnContext(ctx, "[str_replace_editor] 命令为空")
		return FileEditorRes{
			Error: "命令不能为空",
		}, nil
	}

	var result FileEditorRes
	switch req.Command {
	case "view":
		result, _ = f.viewFile(ctx, req.Path, req.ViewRange)
	case "create":
		result, _ = f.createFile(ctx, req.Path, req.FileText)
	case "str_replace":
		result, _ = f.strReplace(ctx, req.Path, req.OldStr, req.NewStr)
	case "insert":
		result, _ = f.insertContent(ctx, req.Path, req.InsertLine, req.NewStr)
	default:
		slog.WarnContext(ctx, "[str_replace_editor] 未知命令", "command", req.Command)
		return FileEditorRes{
			Error: fmt.Sprintf("未知命令: %s。支持的命令: view, create, str_replace, insert", req.Command),
		}, nil
	}

	if result.Error != "" {
		slog.ErrorContext(ctx, "[str_replace_editor] 操作失败", "command", req.Command, "path", req.Path, "error", result.Error)
	} else {
		slog.InfoContext(ctx, "[str_replace_editor] 操作成功", "command", req.Command, "path", req.Path)
	}

	return result, nil
}

// viewFile 查看文件内容
//
// 该方法读取并返回文件内容。
//
// 参数:
//   - path: 文件路径
//   - viewRange: 可选的行范围 [start, end]
//
// 返回:
//   - FileEditorRes: 操作结果
func (f *FileEditorToolImpl) viewFile(ctx context.Context, path string, viewRange []int) (FileEditorRes, error) {
	slog.DebugContext(ctx, "[str_replace_editor] 查看文件", "path", path, "viewRange", viewRange)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return FileEditorRes{
			Error: fmt.Sprintf("文件不存在: %s", path),
		}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return FileEditorRes{
			Error: fmt.Sprintf("读取文件失败: %v", err),
		}, nil
	}

	if len(viewRange) == 2 {
		lines := strings.Split(string(content), "\n")
		start, end := viewRange[0], viewRange[1]
		if start < 1 {
			start = 1
		}
		if end > len(lines) || end == -1 {
			end = len(lines)
		}
		if start <= end && start <= len(lines) {
			content = []byte(strings.Join(lines[start-1:end], "\n"))
		}
	}

	return FileEditorRes{
		Output: string(content),
	}, nil
}

// createFile 创建新文件
//
// 该方法创建新文件并写入内容。
//
// 参数:
//   - path: 文件路径
//   - content: 文件内容
//
// 返回:
//   - FileEditorRes: 操作结果
func (f *FileEditorToolImpl) createFile(ctx context.Context, path string, content *string) (FileEditorRes, error) {
	slog.DebugContext(ctx, "[str_replace_editor] 创建文件", "path", path)

	if content == nil || *content == "" {
		return FileEditorRes{
			Error: "文件内容不能为空",
		}, nil
	}

	if _, err := os.Stat(path); err == nil {
		return FileEditorRes{
			Error: fmt.Sprintf("文件已存在: %s", path),
		}, nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return FileEditorRes{
			Error: fmt.Sprintf("创建目录失败: %v", err),
		}, nil
	}

	if err := os.WriteFile(path, []byte(*content), 0644); err != nil {
		return FileEditorRes{
			Error: fmt.Sprintf("写入文件失败: %v", err),
		}, nil
	}

	return FileEditorRes{
		Output: fmt.Sprintf("文件创建成功: %s", path),
	}, nil
}

// strReplace 字符串替换
//
// 该方法在文件中查找并替换字符串。
//
// 参数:
//   - path: 文件路径
//   - oldStr: 要替换的字符串
//   - newStr: 新字符串
//
// 返回:
//   - FileEditorRes: 操作结果
func (f *FileEditorToolImpl) strReplace(ctx context.Context, path string, oldStr, newStr *string) (FileEditorRes, error) {
	slog.DebugContext(ctx, "[str_replace_editor] 字符串替换", "path", path)

	if oldStr == nil || *oldStr == "" {
		return FileEditorRes{
			Error: "要替换的字符串不能为空",
		}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return FileEditorRes{
			Error: fmt.Sprintf("读取文件失败: %v", err),
		}, nil
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, *oldStr) {
		return FileEditorRes{
			Error: fmt.Sprintf("未找到要替换的字符串: %s", *oldStr),
		}, nil
	}

	count := strings.Count(contentStr, *oldStr)
	if count > 1 {
		return FileEditorRes{
			Error: fmt.Sprintf("找到多个匹配项 (%d 个)，请提供更具体的字符串", count),
		}, nil
	}

	newContent := strings.Replace(contentStr, *oldStr, *newStr, 1)

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return FileEditorRes{
			Error: fmt.Sprintf("写入文件失败: %v", err),
		}, nil
	}

	return FileEditorRes{
		Output: "字符串替换成功",
	}, nil
}

// insertContent 插入内容
//
// 该方法在指定行后插入内容。
//
// 参数:
//   - path: 文件路径
//   - insertLine: 插入位置（在该行之后插入）
//   - newStr: 要插入的内容
//
// 返回:
//   - FileEditorRes: 操作结果
func (f *FileEditorToolImpl) insertContent(ctx context.Context, path string, insertLine *int, newStr *string) (FileEditorRes, error) {
	slog.DebugContext(ctx, "[str_replace_editor] 插入内容", "path", path, "insertLine", insertLine)

	if insertLine == nil {
		return FileEditorRes{
			Error: "插入行号不能为空",
		}, nil
	}

	if newStr == nil || *newStr == "" {
		return FileEditorRes{
			Error: "插入内容不能为空",
		}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return FileEditorRes{
			Error: fmt.Sprintf("读取文件失败: %v", err),
		}, nil
	}

	lines := strings.Split(string(content), "\n")

	if *insertLine < 0 || *insertLine > len(lines) {
		return FileEditorRes{
			Error: fmt.Sprintf("行号超出范围: %d (文件共 %d 行)", *insertLine, len(lines)),
		}, nil
	}

	insertIdx := *insertLine
	if insertIdx < len(lines) {
		lines = append(lines[:insertIdx], append([]string{*newStr}, lines[insertIdx:]...)...)
	} else {
		lines = append(lines, *newStr)
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return FileEditorRes{
			Error: fmt.Sprintf("写入文件失败: %v", err),
		}, nil
	}

	return FileEditorRes{
		Output: fmt.Sprintf("内容已插入到第 %d 行之后", *insertLine),
	}, nil
}

// FileEditorReq 文件编辑请求结构体
//
// 定义了文件编辑工具的输入参数。
type FileEditorReq struct {
	// Command 命令类型
	// 可选值: view, create, str_replace, insert
	Command string `json:"command" jsonschema_description:"命令类型: view(查看), create(创建), str_replace(替换), insert(插入)"`

	// Path 文件路径（绝对路径）
	Path string `json:"path" jsonschema_description:"文件的绝对路径"`

	// FileText 文件内容
	// 用于 create 命令
	FileText *string `json:"file_text,omitempty" jsonschema_description:"文件内容，用于 create 命令"`

	// ViewRange 查看范围
	// 用于 view 命令，[start_line, end_line]
	ViewRange []int `json:"view_range,omitempty" jsonschema_description:"查看行范围，用于 view 命令，如 [1, 10]"`

	// OldStr 要替换的字符串
	// 用于 str_replace 命令
	OldStr *string `json:"old_str,omitempty" jsonschema_description:"要替换的字符串，用于 str_replace 命令"`

	// NewStr 新字符串
	// 用于 str_replace 和 insert 命令
	NewStr *string `json:"new_str,omitempty" jsonschema_description:"新字符串，用于 str_replace 和 insert 命令"`

	// InsertLine 插入行号
	// 用于 insert 命令，在该行之后插入
	InsertLine *int `json:"insert_line,omitempty" jsonschema_description:"插入行号，用于 insert 命令，在该行之后插入内容"`
}

// FileEditorRes 文件编辑响应结构体
//
// 定义了文件编辑工具的输出结果。
type FileEditorRes struct {
	// Output 操作输出
	Output string `json:"output,omitempty" jsonschema_description:"操作输出"`

	// Error 错误信息
	// 当操作失败时包含友好的错误描述
	Error string `json:"error,omitempty" jsonschema_description:"错误信息"`
}

// LocalOperator 本地文件系统操作器
//
// 该结构体实现了 commandline.Operator 接口，
// 提供本地文件系统的读写和命令执行能力。
type LocalOperator struct{}

// NewLocalOperator 创建本地文件系统操作器
func NewLocalOperator() *LocalOperator {
	return &LocalOperator{}
}

// ReadFile 读取文件内容
func (o *LocalOperator) ReadFile(ctx context.Context, path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile 写入文件内容
func (o *LocalOperator) WriteFile(ctx context.Context, path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// IsDirectory 检查路径是否为目录
func (o *LocalOperator) IsDirectory(ctx context.Context, path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// Exists 检查路径是否存在
func (o *LocalOperator) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// RunCommand 执行系统命令
func (o *LocalOperator) RunCommand(ctx context.Context, command []string) (*CommandOutput, error) {
	if len(command) == 0 {
		return nil, nil
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-Command", strings.Join(command, " "))
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", strings.Join(command, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &CommandOutput{
			Stdout:   "",
			Stderr:   string(output),
			ExitCode: 1,
		}, err
	}

	return &CommandOutput{
		Stdout:   string(output),
		Stderr:   "",
		ExitCode: 0,
	}, nil
}

// CommandOutput 命令执行输出
type CommandOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
}
