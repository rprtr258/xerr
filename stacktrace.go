package xerr

import "runtime"

const _maxStackDepth = 50

type stackFrame struct {
	Function string
	File     string
	Line     int
}

func stacktrace(skip int) []stackFrame {
	callers := make([]uintptr, _maxStackDepth)
	length := runtime.Callers(2, callers[:])
	if length >= len(callers) {
		return nil
	}
	callers = callers[:length]

	frames := runtime.CallersFrames(callers)
	for i := 0; i < skip; i++ {
		_, more := frames.Next()
		if !more {
			return nil
		}
	}

	stack := make([]stackFrame, 0, len(callers)-skip)
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

func caller() stackFrame {
	callers := make([]uintptr, 1)
	length := runtime.Callers(3, callers[:])
	callers = callers[:length]

	frames := runtime.CallersFrames(callers)
	frame, _ := frames.Next()
	return stackFrame{
		Function: frame.Function,
		File:     frame.File,
		Line:     frame.Line,
	}
}
