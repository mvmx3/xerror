package xerror

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// 定义一些基础错误
var (
	debugMode = true // 控制是否显示函数调用信息
)

// 自定义错误类型
type validationError struct {
	Field string
	Msg   string
}

func (e *validationError) Error() string {
	return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Msg)
}

// 用于存储发送方的错误信息
type wrappedError struct {
	err        error
	senderInfo string
	senderLoc  string
	rawErr     string // 添加原始错误信息
}

func (e *wrappedError) Error() string {
	return e.err.Error()
}

// 包装发送方错误
func WrapSendError(err error) error {
	_, funcName, location := currentFunction()
	creator := getGoroutineCreator()
	senderInfo := fmt.Sprintf("%s [goroutine: %s]", funcName, creator)

	// 检查是否已经被 WrapError 包装过
	errStr := err.Error()
	var rawErr string
	if !strings.Contains(errStr, "[goroutine:") {
		rawErr = fmt.Sprintf("error: %v", err)
	} else {
		rawErr = errStr
	}

	return &wrappedError{
		err:        err,
		senderInfo: senderInfo,
		senderLoc:  location,
		rawErr:     rawErr,
	}
}

// 包装接收方错误
func WrapReceiveError(err error) error {
	_, funcName, location := currentFunction()
	creator := getGoroutineCreator()

	if werr, ok := err.(*wrappedError); ok {
		return fmt.Errorf("%s -> %s [goroutine: %s]\n"+
			"     %s -> %s\n"+
			"     \n"+
			"%s",
			werr.senderInfo, funcName, creator,
			werr.senderLoc, location,
			werr.rawErr)
	}

	return fmt.Errorf("%s [goroutine: %s]\n     %s",
		funcName, creator, location)
}

// 获取当前函数名和位置信息
func currentFunction() (string, string, string) {
	if !debugMode {
		return "", "", ""
	}
	pc, file, line, _ := runtime.Caller(2)         // 获取实际调用者
	filename := filepath.Base(file)                // 只获取文件名
	filename = strings.TrimSuffix(filename, ".go") // 移除 .go 后缀
	funcName := runtime.FuncForPC(pc).Name()
	location := fmt.Sprintf("%s:%d", file, line) // 完整路径和行号
	return filename, funcName, location
}

// 获取启动当前 goroutine 的函数名
func getGoroutineCreator() string {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	stack := string(buf[:n])
	lines := strings.Split(stack, "\n")

	// 查找 "created by" 行
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "created by ") {
			// 获取前一个函数调用行
			if i >= 2 { // 确保有前一行
				caller := strings.TrimSpace(lines[i-2]) // 跳过文件行,获取函数调用行
				// 提取函数名
				if idx := strings.Index(caller, "("); idx != -1 {
					caller = caller[:idx]
				}
				// 提取最后一个点后的函数名
				if idx := strings.LastIndex(caller, "."); idx != -1 {
					return caller[idx+1:]
				}
				return caller
			}
			break
		}
	}
	return "main"
}

func WrapError(err error) error {
	_, funcName, location := currentFunction()
	creator := getGoroutineCreator()

	// 检查 err 是否已经被 WrapError 包装过
	errStr := err.Error()
	if !strings.Contains(errStr, "[goroutine:") {
		// 第一次调用,添加 "error: " 前缀
		return fmt.Errorf("%s [goroutine: %s]\n"+
			"     %s\n"+
			"     \n"+
			"error: %v", funcName, creator, location, err)
	}

	// 非第一次调用,保持原格式
	return fmt.Errorf("%s [goroutine: %s]\n"+
		"     %s\n"+
		"     \n"+
		"%v", funcName, creator, location, err)
}
