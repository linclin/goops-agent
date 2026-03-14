// Package utils 提供通用工具函数
//
// 该包提供应用程序的通用工具函数。
// 主要功能包括：
//   - 调试信息打印
//   - 安全的 goroutine 启动
//   - JSON 序列化工具
package utils

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"runtime"
)

// 调试输出相关常量
var (
	// dunno 未知函数名占位符
	dunno = []byte("???")

	// centerDot 中点字符
	centerDot = []byte("·")

	// dot 点字符
	dot = []byte(".")
)

// pointerInfo 指针信息结构体
//
// 用于跟踪指针引用关系，防止循环引用导致的无限递归。
type pointerInfo struct {
	// prev 前一个指针信息
	prev *pointerInfo

	// n 指针编号
	n int

	// addr 指针地址
	addr uintptr

	// pos 在输出缓冲区中的位置
	pos int

	// used 被引用的位置列表
	used []int
}

// Display 在控制台打印调试数据
//
// 该函数用于调试目的，在控制台打印变量的详细信息。
// 包括调用位置、变量名和变量值。
//
// 参数:
//   - data: 变量名和变量值的交替列表
//
// 使用示例:
//
//	name := "张三"
//	age := 25
//	utils.Display("name", name, "age", age)
//
// 输出示例:
//
//	[Debug] at main() [main.go:10]
//
//	[Variables]
//	name = "张三"
//	age = 25
func Display(data ...interface{}) {
	display(true, data...)
}

// GetDisplayString 返回调试数据的字符串表示
//
// 该函数返回变量的详细字符串表示，不打印到控制台。
//
// 参数:
//   - data: 变量名和变量值的交替列表
//
// 返回:
//   - string: 调试信息字符串
//
// 使用示例:
//
//	str := utils.GetDisplayString("name", name)
//	fmt.Println(str)
func GetDisplayString(data ...interface{}) string {
	return display(false, data...)
}

// display 内部实现函数
//
// 该函数实现调试信息的格式化和输出。
//
// 参数:
//   - displayed: 是否打印到控制台
//   - data: 变量名和变量值的交替列表
//
// 返回:
//   - string: 调试信息字符串
func display(displayed bool, data ...interface{}) string {
	// 获取调用者信息
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return ""
	}

	// 创建输出缓冲区
	buf := new(bytes.Buffer)

	// 打印调用位置
	fmt.Fprintf(buf, "[Debug] at %s() [%s:%d]\n", function(pc), file, line)
	fmt.Fprintf(buf, "\n[Variables]\n")

	// 打印变量
	for i := 0; i < len(data); i += 2 {
		output := fomateinfo(len(data[i].(string))+3, data[i+1])
		fmt.Fprintf(buf, "%s = %s", data[i], output)
	}

	// 如果需要打印到控制台
	if displayed {
		log.Print(buf)
	}
	return buf.String()
}

// fomateinfo 格式化变量信息
//
// 该函数将变量格式化为可读的字符串表示。
//
// 参数:
//   - headlen: 头部长度
//   - data: 要格式化的变量
//
// 返回:
//   - []byte: 格式化后的字节切片
func fomateinfo(headlen int, data ...interface{}) []byte {
	buf := new(bytes.Buffer)

	// 如果有多个变量，使用数组格式
	if len(data) > 1 {
		fmt.Fprint(buf, "    ")
		fmt.Fprint(buf, "[")
		fmt.Fprintln(buf)
	}

	// 遍历并格式化每个变量
	for k, v := range data {
		buf2 := new(bytes.Buffer)
		var pointers *pointerInfo
		interfaces := make([]reflect.Value, 0, 10)

		printKeyValue(buf2, reflect.ValueOf(v), &pointers, &interfaces, nil, true, "    ", 1)

		if k < len(data)-1 {
			fmt.Fprint(buf2, ", ")
		}
		fmt.Fprintln(buf2)

		buf.Write(buf2.Bytes())
	}

	if len(data) > 1 {
		fmt.Fprintln(buf)
		fmt.Fprint(buf, "    ")
		fmt.Fprint(buf, "]")
	}

	return buf.Bytes()
}

// isSimpleType 检查是否为简单类型
//
// 该函数检查值是否为 Go 的基本类型，用于决定格式化方式。
//
// 参数:
//   - val: 反射值
//   - kind: 类型种类
//   - pointers: 指针信息
//   - interfaces: 接口值列表
//
// 返回:
//   - bool: 是否为简单类型
func isSimpleType(val reflect.Value, kind reflect.Kind, pointers **pointerInfo, interfaces *[]reflect.Value) bool {
	switch kind {
	case reflect.Bool:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint8, reflect.Uint16, reflect.Uint, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Complex64, reflect.Complex128:
		return true
	case reflect.String:
		return true
	case reflect.Chan:
		return true
	case reflect.Invalid:
		return true
	case reflect.Interface:
		for _, in := range *interfaces {
			if reflect.DeepEqual(in, val) {
				return true
			}
		}
		return false
	case reflect.UnsafePointer:
		if val.IsNil() {
			return true
		}
		elem := val.Elem()
		if isSimpleType(elem, elem.Kind(), pointers, interfaces) {
			return true
		}
		addr := val.Elem().UnsafeAddr()
		for p := *pointers; p != nil; p = p.prev {
			if addr == p.addr {
				return true
			}
		}
		return false
	}
	return false
}

// printKeyValue 打印键值
//
// 该函数递归打印反射值的详细信息。
//
// 参数:
//   - buf: 输出缓冲区
//   - val: 反射值
//   - pointers: 指针信息
//   - interfaces: 接口值列表
//   - structFilter: 结构体字段过滤器
//   - formatOutput: 是否格式化输出
//   - indent: 缩进字符串
//   - level: 嵌套层级
func printKeyValue(buf *bytes.Buffer, val reflect.Value, pointers **pointerInfo, interfaces *[]reflect.Value, structFilter func(string, string) bool, formatOutput bool, indent string, level int) {
	t := val.Kind()

	switch t {
	case reflect.Bool:
		fmt.Fprint(buf, val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprint(buf, val.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint, reflect.Uint32, reflect.Uint64:
		fmt.Fprint(buf, val.Uint())
	case reflect.Float32, reflect.Float64:
		fmt.Fprint(buf, val.Float())
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprint(buf, val.Complex())
	case reflect.UnsafePointer:
		fmt.Fprintf(buf, "unsafe.Pointer(0x%X)", val.Pointer())
	case reflect.Ptr:
		if val.IsNil() {
			fmt.Fprint(buf, "nil")
			return
		}
		addr := val.Elem().UnsafeAddr()
		for p := *pointers; p != nil; p = p.prev {
			if addr == p.addr {
				p.used = append(p.used, buf.Len())
				fmt.Fprintf(buf, "0x%X", addr)
				return
			}
		}
		*pointers = &pointerInfo{
			prev: *pointers,
			addr: addr,
			pos:  buf.Len(),
			used: make([]int, 0),
		}
		fmt.Fprint(buf, "&")
		printKeyValue(buf, val.Elem(), pointers, interfaces, structFilter, formatOutput, indent, level)
	case reflect.String:
		fmt.Fprint(buf, "\"", val.String(), "\"")
	case reflect.Interface:
		value := val.Elem()
		if !value.IsValid() {
			fmt.Fprint(buf, "nil")
		} else {
			for _, in := range *interfaces {
				if reflect.DeepEqual(in, val) {
					fmt.Fprint(buf, "repeat")
					return
				}
			}
			*interfaces = append(*interfaces, val)
			printKeyValue(buf, value, pointers, interfaces, structFilter, formatOutput, indent, level+1)
		}
	case reflect.Struct:
		t := val.Type()
		fmt.Fprint(buf, t)
		fmt.Fprint(buf, "{")
		for i := 0; i < val.NumField(); i++ {
			if formatOutput {
				fmt.Fprintln(buf)
			} else {
				fmt.Fprint(buf, " ")
			}
			name := t.Field(i).Name
			if formatOutput {
				for ind := 0; ind < level; ind++ {
					fmt.Fprint(buf, indent)
				}
			}
			fmt.Fprint(buf, name)
			fmt.Fprint(buf, ": ")
			if structFilter != nil && structFilter(t.String(), name) {
				fmt.Fprint(buf, "ignore")
			} else {
				printKeyValue(buf, val.Field(i), pointers, interfaces, structFilter, formatOutput, indent, level+1)
			}
			fmt.Fprint(buf, ",")
		}
		if formatOutput {
			fmt.Fprintln(buf)
			for ind := 0; ind < level-1; ind++ {
				fmt.Fprint(buf, indent)
			}
		} else {
			fmt.Fprint(buf, " ")
		}
		fmt.Fprint(buf, "}")
	case reflect.Array, reflect.Slice:
		fmt.Fprint(buf, val.Type())
		fmt.Fprint(buf, "{")
		allSimple := true
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			isSimple := isSimpleType(elem, elem.Kind(), pointers, interfaces)
			if !isSimple {
				allSimple = false
			}
			if formatOutput && !isSimple {
				fmt.Fprintln(buf)
			} else {
				fmt.Fprint(buf, " ")
			}
			if formatOutput && !isSimple {
				for ind := 0; ind < level; ind++ {
					fmt.Fprint(buf, indent)
				}
			}
			printKeyValue(buf, elem, pointers, interfaces, structFilter, formatOutput, indent, level+1)
			if i != val.Len()-1 || !allSimple {
				fmt.Fprint(buf, ",")
			}
		}
		if formatOutput && !allSimple {
			fmt.Fprintln(buf)
			for ind := 0; ind < level-1; ind++ {
				fmt.Fprint(buf, indent)
			}
		} else {
			fmt.Fprint(buf, " ")
		}
		fmt.Fprint(buf, "}")
	case reflect.Map:
		t := val.Type()
		keys := val.MapKeys()
		fmt.Fprint(buf, t)
		fmt.Fprint(buf, "{")
		allSimple := true
		for i := 0; i < len(keys); i++ {
			elem := val.MapIndex(keys[i])
			isSimple := isSimpleType(elem, elem.Kind(), pointers, interfaces)
			if !isSimple {
				allSimple = false
			}
			if formatOutput && !isSimple {
				fmt.Fprintln(buf)
			} else {
				fmt.Fprint(buf, " ")
			}
			if formatOutput && !isSimple {
				for ind := 0; ind <= level; ind++ {
					fmt.Fprint(buf, indent)
				}
			}
			printKeyValue(buf, keys[i], pointers, interfaces, structFilter, formatOutput, indent, level+1)
			fmt.Fprint(buf, ": ")
			printKeyValue(buf, elem, pointers, interfaces, structFilter, formatOutput, indent, level+1)
			if i != val.Len()-1 || !allSimple {
				fmt.Fprint(buf, ",")
			}
		}
		if formatOutput && !allSimple {
			fmt.Fprintln(buf)
			for ind := 0; ind < level-1; ind++ {
				fmt.Fprint(buf, indent)
			}
		} else {
			fmt.Fprint(buf, " ")
		}
		fmt.Fprint(buf, "}")
	case reflect.Chan:
		fmt.Fprint(buf, val.Type())
	case reflect.Invalid:
		fmt.Fprint(buf, "invalid")
	default:
		fmt.Fprint(buf, "unknow")
	}
}

// PrintPointerInfo 打印指针信息
//
// 该函数打印指针引用关系的可视化图表。
//
// 参数:
//   - buf: 输出缓冲区
//   - headlen: 头部长度
//   - pointers: 指针信息
func PrintPointerInfo(buf *bytes.Buffer, headlen int, pointers *pointerInfo) {
	anyused := false
	pointerNum := 0

	for p := pointers; p != nil; p = p.prev {
		if len(p.used) > 0 {
			anyused = true
		}
		pointerNum++
		p.n = pointerNum
	}

	if anyused {
		pointerBufs := make([][]rune, pointerNum+1)
		for i := 0; i < len(pointerBufs); i++ {
			pointerBuf := make([]rune, buf.Len()+headlen)
			for j := 0; j < len(pointerBuf); j++ {
				pointerBuf[j] = ' '
			}
			pointerBufs[i] = pointerBuf
		}

		for pn := 0; pn <= pointerNum; pn++ {
			for p := pointers; p != nil; p = p.prev {
				if len(p.used) > 0 && p.n >= pn {
					if pn == p.n {
						pointerBufs[pn][p.pos+headlen] = '└'
						maxpos := 0
						for i, pos := range p.used {
							if i < len(p.used)-1 {
								pointerBufs[pn][pos+headlen] = '┴'
							} else {
								pointerBufs[pn][pos+headlen] = '┘'
							}
							maxpos = pos
						}
						for i := 0; i < maxpos-p.pos-1; i++ {
							if pointerBufs[pn][i+p.pos+headlen+1] == ' ' {
								pointerBufs[pn][i+p.pos+headlen+1] = '─'
							}
						}
					} else {
						pointerBufs[pn][p.pos+headlen] = '│'
						for _, pos := range p.used {
							if pointerBufs[pn][pos+headlen] == ' ' {
								pointerBufs[pn][pos+headlen] = '│'
							} else {
								pointerBufs[pn][pos+headlen] = '┼'
							}
						}
					}
				}
			}
			buf.WriteString(string(pointerBufs[pn]) + "\n")
		}
	}
}

// Stack 获取调用堆栈
//
// 该函数返回调用堆栈的字节切片。
//
// 参数:
//   - skip: 跳过的调用层数
//   - indent: 缩进字符串
//
// 返回:
//   - []byte: 堆栈信息字节切片
//
// 使用示例:
//
//	stack := utils.Stack(0, "  ")
//	fmt.Println(string(stack))
func Stack(skip int, indent string) []byte {
	buf := new(bytes.Buffer)

	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buf.WriteString(indent)
		fmt.Fprintf(buf, "at %s() [%s:%d]\n", function(pc), file, line)
	}

	return buf.Bytes()
}

// function 获取函数名
//
// 该函数返回包含 PC 的函数名。
//
// 参数:
//   - pc: 程序计数器
//
// 返回:
//   - []byte: 函数名字节切片
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// 去除包路径，只保留函数名
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	// 替换中点为点
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
