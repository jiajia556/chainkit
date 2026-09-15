package service

import (
	"fmt"
	"runtime"
	"strings"
)

// Error adds the service operation and source location where an error crossed
// the public package boundary. The original error remains available through
// Unwrap, so callers can continue to use errors.Is and errors.As.
type Error struct {
	Operation string
	File      string
	Line      int
	Err       error
}

func (e *Error) Error() string {
	return fmt.Sprintf("service.%s (%s:%d): %v", e.Operation, e.File, e.Line, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// wrapServiceErrors is intended to be deferred at the start of an exported
// service function. runtime.Caller reports that defer statement, making the
// returned location stable and immediately useful to package callers.
func wrapServiceErrors(operation string, targets ...*error) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	} else {
		file = serviceRelativePath(file)
	}

	for _, target := range targets {
		if target == nil || *target == nil {
			continue
		}
		*target = &Error{
			Operation: operation,
			File:      file,
			Line:      line,
			Err:       *target,
		}
	}
}

func serviceRelativePath(file string) string {
	file = strings.ReplaceAll(file, "\\", "/")
	if index := strings.LastIndex(file, "/service/"); index >= 0 {
		return file[index+1:]
	}
	return file
}
