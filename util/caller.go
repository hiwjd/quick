package util

import (
	"fmt"
	"runtime"
)

// Caller 返回调用方的代码路径，行号，方法名
func Caller(skip int) (file string, line int, function string) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.File, frame.Line, frame.Function
}

func WrapCaller(file string, line int, fn string) string {
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	return fmt.Sprintf("%s:%d", file, line)
}
