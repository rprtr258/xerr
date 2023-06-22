package xerr

import (
	"fmt"
	"runtime"
	"sync"
)

var (
	// _helperPCs - helper functions names
	_helperPCs = map[string]struct{}{}
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
	_helperPCs[frame.Function] = struct{}{}
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

func stacktrace() []stackFrame {
	callers := make([]uintptr, _maxStackDepth)
	length := runtime.Callers(2, callers[:])
	callers = callers[:length]

	frames := runtime.CallersFrames(callers)
	stack := make([]stackFrame, 0, len(callers))
	for {
		frame, more := frames.Next()
		if _, ok := _helperPCs[frame.Function]; !ok {
			stack = append(
				stack,
				stackFrame{
					Function: frame.Function,
					File:     frame.File,
					Line:     frame.Line,
				},
			)
		}
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
		frame, more := frames.Next()
		if !more {
			break
		}

		if _, ok := _helperPCs[frame.Function]; ok {
			continue
		}

		return &stackFrame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		}
	}

	panic("all functions in stack are helpers")
}
