package errors

import (
	"fmt"

	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
)

// Op is an operation that caused an error.
type Op string

// Code is a code that describes an error.
// Should be modified to match the environment (e.g. GRPC).
type Code codes.Code

var (
	KindNotFound     = codes.NotFound
	KindInvalid      = codes.InvalidArgument
	KindConflict     = codes.AlreadyExists
	KindUnauthorized = codes.Unauthenticated
	KindInternal     = codes.Internal
	KindUnknown      = codes.Unknown
)

type Error struct {
	Op       Op    // Operation that caused the error
	Kind     Code  // Kind of error
	Err      error // Wrapped error
	Severity zapcore.Level
	// application specific data
}

// Error returns the error string.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Op, e.Err.Error())
}

// New returns a new error with the given arguments.
// See Error struct for more details.
func New(args ...interface{}) error {
	e := &Error{}
	for _, arg := range args {
		switch arg.(type) {
		case Op:
			e.Op = arg.(Op)
		case Code:
			e.Kind = arg.(Code)
		case error:
			e.Err = arg.(error)
		case zapcore.Level:
			e.Severity = arg.(zapcore.Level)
		default:
			panic("bad call to E")
		}
	}
	return e
}

// Ops returns the operations that caused the error.
func Ops(e *Error) []Op {
	res := []Op{e.Op}

	subErr, ok := e.Err.(*Error)
	if !ok {
		return res
	}
	res = append(res, Ops(subErr)...)
	return res
}

// Kind returns the first found error kind.
func Kind(err error) codes.Code {
	e, ok := err.(*Error)
	if !ok {
		return KindUnknown
	}

	if e.Kind != 0 {
		return codes.Code(e.Kind)
	}
	return Kind(e.Err)
}
