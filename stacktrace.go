package xerr

import (
	"fmt"
	"runtime"
	"sync"
)

type funcID struct {
	Function string
	File     string
}

var (
	_helperPCs = map[funcID]struct{}{}
	helpersMu  = &sync.Mutex{}
)

func Helper() {
	helpersMu.Lock()
	defer helpersMu.Unlock()

	var pc [1]uintptr
	if n := runtime.Callers(2, pc[:]); n == 0 { // skip runtime.Callers + Helper
		panic("testing: zero callers found")
	}
	frames := runtime.CallersFrames(pc[:])
	frame, _ := frames.Next()
	_helperPCs[funcID{
		Function: frame.Function,
		File:     frame.File,
	}] = struct{}{}
}

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

func getCaller() *stackFrame {
	Helper()

	callers := make([]uintptr, 1+len(_helperPCs))
	length := runtime.Callers(1, callers[:])
	callers = callers[:length]

	frames := runtime.CallersFrames(callers)
	for {
		rawFrame, more := frames.Next()
		if !more {
			break
		}

		funcID := funcID{
			Function: rawFrame.Function,
			File:     rawFrame.File,
		}

		if _, ok := _helperPCs[funcID]; ok {
			continue
		}

		return &stackFrame{
			Function: rawFrame.Function,
			File:     rawFrame.File,
			Line:     rawFrame.Line,
		}
	}

	panic("all functions in stack are helpers")
}
