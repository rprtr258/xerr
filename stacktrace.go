package xerr

import (
	"fmt"
	"runtime"
)

const _maxStackDepth = 50

type stackFrame struct {
	Function string
	File     string
	Line     int
}

func (sf *stackFrame) String() string {
	if sf == nil {
		return ""
	}

	return fmt.Sprintf("%s#%s:%d", sf.File, sf.Function, sf.Line)
}

func stacktrace(skip int) []stackFrame {
	callers := make([]uintptr, _maxStackDepth)
	length := runtime.Callers(2+skip, callers[:])
	callers = callers[:length]

	frames := runtime.CallersFrames(callers)
	stack := make([]stackFrame, 0, len(callers))
	for {
		frame, more := frames.Next()
		stack = append(
			stack,
			stackFrame{
				Function: frame.Function,
				File:     frame.File,
				Line:     frame.Line,
			},
		)
		if !more {
			break
		}
	}

	return stack
}

func caller(skip int) *stackFrame {
	callers := make([]uintptr, 1)
	length := runtime.Callers(3+skip, callers[:])
	callers = callers[:length]

	frames := runtime.CallersFrames(callers)
	frame, _ := frames.Next()
	return &stackFrame{
		Function: frame.Function,
		File:     frame.File,
		Line:     frame.Line,
	}
}
