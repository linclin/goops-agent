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
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// CommandToolImpl 命令执行工具实现
//
// 该工具用于在系统终端中执行命令，支持 Windows、Linux 和 macOS。
// 提供安全检查机制，防止执行危险命令。
//
// 使用场景:
//   - 执行系统命令获取信息
//   - 运行脚本或程序
//   - 文件操作（复制、移动、删除等）
//
// 安全特性:
//   - 危险命令检测与拦截
//   - 超时控制
//   - 工作目录限制
type CommandToolImpl struct {
	config *CommandToolConfig
}

// CommandToolConfig 命令执行工具配置
//
// 可配置超时时间、工作目录和安全限制。
type CommandToolConfig struct {
	// Timeout 命令执行超时时间（秒）
	// 默认: 60 秒
	// 超时后命令将被强制终止
	Timeout int

	// WorkingDir 命令执行的工作目录
	// 如果为空，使用当前目录
	WorkingDir string

	// RestrictToWorkspace 是否限制在工作目录内执行
	// 启用后，命令中的路径必须在工作目录内
	RestrictToWorkspace bool

	// MaxOutputLength 最大输出长度
	// 超过此长度的输出将被截断
	// 默认: 50000 字符
	MaxOutputLength int
}

// defaultCommandToolConfig 创建默认配置
//
// 返回默认的工具配置实例。
func defaultCommandToolConfig() *CommandToolConfig {
	return &CommandToolConfig{
		Timeout:             60,
		WorkingDir:          "",
		RestrictToWorkspace: false,
		MaxOutputLength:     50000,
	}
}

// NewCommandTool 创建命令执行工具实例
//
// 该函数创建一个用于执行系统命令的工具。
// 如果未提供配置，将使用默认配置。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - config: 工具配置，可选
//
// 返回:
//   - tool.InvokableTool: 工具实例
//   - error: 创建过程中的错误
func NewCommandTool(ctx context.Context, config *CommandToolConfig) (tool.InvokableTool, error) {
	if config == nil {
		config = defaultCommandToolConfig()
	}

	if config.Timeout <= 0 {
		config.Timeout = 60
	}

	if config.MaxOutputLength <= 0 {
		config.MaxOutputLength = 50000
	}

	t := &CommandToolImpl{config: config}

	return utils.InferTool(
		"execute",
		`在终端中执行命令并返回输出。
支持 Windows (PowerShell)、Linux 和 macOS (sh)。
注意：危险命令（如 rm -rf、format、shutdown 等）将被拦截。
输入应为要执行的命令字符串。`,
		t.Invoke,
	)
}

// Invoke 执行命令
//
// 该方法是工具的核心实现，负责执行指定的命令。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - req: 命令请求，包含要执行的命令
//
// 返回:
//   - CommandRes: 执行结果，包含输出和错误信息
//   - error: 执行错误
//
// 支持的平台:
//   - Windows: 使用 PowerShell 执行
//   - Linux/macOS: 使用 sh 执行
func (c *CommandToolImpl) Invoke(ctx context.Context, req CommandReq) (CommandRes, error) {
	slog.InfoContext(ctx, "[execute] 工具调用开始", "command", req.Command)

	if req.Command == "" {
		slog.WarnContext(ctx, "[execute] 命令为空")
		return CommandRes{
			Success: false,
			Error:   "命令不能为空",
		}, nil
	}

	if err := c.guardCommand(req.Command); err != nil {
		slog.ErrorContext(ctx, "[execute] 安全检查失败", "command", req.Command, "error", err)
		return CommandRes{
			Success: false,
			Error:   fmt.Sprintf("安全检查失败: %s", err.Error()),
		}, nil
	}

	timeout := c.config.Timeout
	if timeout <= 0 {
		timeout = 60
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := c.buildCommand(ctx, req.Command)

	if c.config.WorkingDir != "" {
		cmd.Dir = c.config.WorkingDir
	}

	slog.DebugContext(ctx, "[execute] 开始执行命令", "timeout", timeout, "working_dir", cmd.Dir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var result strings.Builder
	if stdout.Len() > 0 {
		result.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString("[STDERR] " + stderr.String())
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			slog.ErrorContext(ctx, "[execute] 命令执行超时", "command", req.Command, "timeout", timeout)
			return CommandRes{
				Success: false,
				Error:   fmt.Sprintf("命令执行超时（%d秒）", timeout),
			}, nil
		}
		slog.ErrorContext(ctx, "[execute] 命令执行失败", "command", req.Command, "error", err)
		if result.Len() > 0 {
			return CommandRes{
				Success: false,
				Output:  c.truncateOutput(result.String()),
				Error:   fmt.Sprintf("命令执行失败: %s", err.Error()),
			}, nil
		}
		return CommandRes{
			Success: false,
			Error:   fmt.Sprintf("命令执行失败: %s", err.Error()),
		}, nil
	}

	if result.Len() == 0 {
		slog.InfoContext(ctx, "[execute] 执行成功，无输出", "command", req.Command)
		return CommandRes{
			Success: true,
			Output:  "(无输出)",
		}, nil
	}

	slog.InfoContext(ctx, "[execute] 执行成功", "command", req.Command, "output_length", result.Len())
	return CommandRes{
		Success: true,
		Output:  c.truncateOutput(result.String()),
	}, nil
}

// CommandReq 命令请求结构体
//
// 定义了命令执行工具的输入参数。
type CommandReq struct {
	// Command 要执行的命令
	// 示例:
	//   - Windows: Get-Process, dir, ipconfig
	//   - Linux/Mac: ls -la, ps aux, ifconfig
	Command string `json:"command" jsonschema_description:"要执行的终端命令"`
}

// CommandRes 命令响应结构体
//
// 定义了命令执行工具的输出结果。
type CommandRes struct {
	// Success 命令是否执行成功
	Success bool `json:"success" jsonschema_description:"命令是否执行成功"`

	// Output 命令的标准输出和错误输出
	Output string `json:"output" jsonschema_description:"命令执行的输出内容"`

	// Error 错误信息
	Error string `json:"error,omitempty" jsonschema_description:"错误信息（如果有）"`
}

// buildCommand 根据操作系统构建命令
//
// 该函数根据操作系统选择合适的 Shell 执行命令。
//
// 参数:
//   - ctx: 上下文
//   - command: 要执行的命令字符串
//
// 返回:
//   - *exec.Cmd: 构建好的命令对象
func (c *CommandToolImpl) buildCommand(ctx context.Context, command string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		return exec.CommandContext(ctx, "powershell", "-Command", command)
	default:
		return exec.CommandContext(ctx, "sh", "-c", command)
	}
}

// truncateOutput 截断过长的输出
//
// 该函数限制输出长度，防止返回过多数据。
//
// 参数:
//   - output: 原始输出
//
// 返回:
//   - string: 截断后的输出
func (c *CommandToolImpl) truncateOutput(output string) string {
	maxLen := c.config.MaxOutputLength
	if maxLen <= 0 {
		maxLen = 50000
	}
	if len(output) > maxLen {
		return output[:maxLen] + "\n... (输出已截断)"
	}
	return output
}

// denyPatterns 危险命令正则模式列表
//
// 这些模式用于检测可能造成系统损坏的危险命令。
var denyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brm\s+-[rf]{1,2}\b`),
	regexp.MustCompile(`(?i)\bdel\s+/[fq]\b`),
	regexp.MustCompile(`(?i)\brmdir\s+/s\b`),
	regexp.MustCompile(`(?i)(?:^|[;&|]\s*)format\b`),
	regexp.MustCompile(`(?i)\b(mkfs|diskpart)\b`),
	regexp.MustCompile(`(?i)\bdd\s+if=`),
	regexp.MustCompile(`(?i)>\s*/dev/sd`),
	regexp.MustCompile(`(?i)\b(shutdown|reboot|poweroff|halt)\b`),
	regexp.MustCompile(`:\(\)\s*\{.*\};\s*:`),
	regexp.MustCompile(`(?i)\b(Format-Volume|Clear-Disk)\b`),
	regexp.MustCompile(`(?i)\bRemove-Item\s+.*-Recurse\s+.*-Force\b`),
	regexp.MustCompile(`(?i)\bStop-Computer\b`),
	regexp.MustCompile(`(?i)\bRestart-Computer\b`),
}

// guardCommand 命令安全检查
//
// 该函数检查命令是否包含危险操作。
//
// 参数:
//   - command: 要检查的命令
//
// 返回:
//   - error: 如果命令危险，返回错误
func (c *CommandToolImpl) guardCommand(command string) error {
	for _, pattern := range denyPatterns {
		if pattern.MatchString(command) {
			return fmt.Errorf("检测到潜在危险操作，命令已被拦截")
		}
	}

	if c.config.RestrictToWorkspace && c.config.WorkingDir != "" {
		if strings.Contains(command, "../") {
			return fmt.Errorf("工作目录限制模式下不允许使用路径遍历 (../)")
		}

		absPaths := c.extractAbsolutePaths(command)
		cwdAbs, err := filepath.Abs(c.config.WorkingDir)
		if err == nil {
			for _, p := range absPaths {
				pAbs, err := filepath.Abs(p)
				if err != nil {
					continue
				}
				if !strings.HasPrefix(pAbs, cwdAbs) {
					return fmt.Errorf("路径 %q 不在工作目录 %q 内", p, c.config.WorkingDir)
				}
			}
		}
	}

	return nil
}

// absolutePathRe 匹配命令中的绝对路径
var absolutePathRe = regexp.MustCompile(`(?:^|\s)(/[^\s;|&>]+)`)

// extractAbsolutePaths 从命令字符串中提取绝对路径
//
// 该函数用于路径安全检查。
//
// 参数:
//   - command: 命令字符串
//
// 返回:
//   - []string: 提取的绝对路径列表
func (c *CommandToolImpl) extractAbsolutePaths(command string) []string {
	matches := absolutePathRe.FindAllStringSubmatch(command, -1)
	var paths []string
	for _, m := range matches {
		if len(m) > 1 {
			p := m[1]
			if p == "/dev/null" || p == "/dev/stdin" || p == "/dev/stdout" || p == "/dev/stderr" {
				continue
			}
			paths = append(paths, p)
		}
	}
	return paths
}

// GetShellInfo 获取当前系统 Shell 信息
//
// 该函数返回当前系统的 Shell 类型。
//
// 返回:
//   - string: Shell 类型（powershell/sh）
//   - string: Shell 路径
func GetShellInfo() (shellType, shellPath string) {
	switch runtime.GOOS {
	case "windows":
		shellType = "powershell"
		if path, err := exec.LookPath("powershell"); err == nil {
			shellPath = path
		} else {
			shellPath = "powershell"
		}
	default:
		shellType = "sh"
		if path, err := exec.LookPath("sh"); err == nil {
			shellPath = path
		} else {
			shellPath = "/bin/sh"
		}
	}
	return
}

// IsCommandAvailable 检查命令是否可用
//
// 该函数检查指定的命令是否在系统 PATH 中可用。
//
// 参数:
//   - cmd: 命令名称
//
// 返回:
//   - bool: 命令是否可用
func IsCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetEnv 获取环境变量
//
// 该函数获取指定环境变量的值。
//
// 参数:
//   - key: 环境变量名
//
// 返回:
//   - string: 环境变量值
func GetEnv(key string) string {
	return os.Getenv(key)
}

// GetHomeDir 获取用户主目录
//
// 该函数返回当前用户的主目录路径。
//
// 返回:
//   - string: 主目录路径
//   - error: 错误信息
func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

// GetCurrentDir 获取当前工作目录
//
// 该函数返回当前工作目录路径。
//
// 返回:
//   - string: 当前目录路径
//   - error: 错误信息
func GetCurrentDir() (string, error) {
	return os.Getwd()
}
