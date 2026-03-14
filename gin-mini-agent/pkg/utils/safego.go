package utils

import (
	"fmt"
	"runtime/debug"

	"gin-mini-agent/pkg/global"
)

// SafeGo 安全地启动 goroutine
//
// 该函数封装了 goroutine 的启动，自动捕获 panic 并记录日志。
// 防止 goroutine 中的 panic 导致整个程序崩溃。
//
// 参数:
//   - f: 要在 goroutine 中执行的函数
//
// 功能特点:
//   - 自动捕获 panic
//   - 记录详细的错误信息和堆栈
//   - 不影响主程序运行
//
// 使用示例:
//
//	utils.SafeGo(func() {
//	    // 这里执行可能 panic 的代码
//	    result := someRiskyOperation()
//	    processResult(result)
//	})
//
// 日志输出:
//
//	{
//	    "level": "ERROR",
//	    "msg": "运行panic异常: runtime error: ...",
//	    "stack": "goroutine 1 [running]:\n..."
//	}
func SafeGo(f func()) {
	go func() {
		// 延迟捕获 panic
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 错误和堆栈信息
				global.Log.Error(fmt.Sprintf("运行panic异常: %v\n堆栈信息: %v", err, string(debug.Stack())))
			}
		}()

		// 执行传入的函数
		f()
	}()
}
