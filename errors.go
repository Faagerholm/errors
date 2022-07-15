package main

import (
	"bytes"
	"fmt"
)

// Op is an operation that caused an error.
type Op string

// Separator is the string used to separate nested errors. By
// default, to make errors easier on the eye, nested errors are
// indented on a new line. A server may instead choose to keep each
// error on a single line by modifying the separator string, perhaps
// to ":: ".
var Separator = ":\n\t"

// Kind describes the kind of error.
// Should be modified to match the environment (e.g. GRPC).
type Kind uint8

const (
	Other Kind = iota // Other is used for errors that are not our own.
	NotFound
	Invalid
	Conflict
	Unauthorized
	Internal
)

// Level represents the severity of the error.
type Level uint8

const (
	InfoLevel Level = iota
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

func (k Kind) String() string {
	switch k {
	case NotFound:
		return "not found"
	case Invalid:
		return "invalid"
	case Conflict:
		return "conflict"
	case Unauthorized:
		return "unauthorized"
	case Internal:
		return "internal"
	case Other:
		return "other error" // can be a normal error and not our own type
	default:
		return "unknown"
	}
}

type Error struct {
	Op       Op    // Operation that caused the error
	Kind     Kind  // Kind of error
	Err      error // Wrapped error
	Severity Level
	// application specific data
}

// isZero returns true if the error is a zero value.
func (e *Error) isZero() bool {
	return e.Op == "" && e.Kind == 0 && e.Err == nil
}

// New returns a new error with the given arguments.
// See Error struct for more details.
func New(args ...interface{}) error {
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.Op = arg
		case Kind:
			e.Kind = arg
		case error:
			e.Err = arg
		case *Error:
			// Make a copy
			copy := *arg
			e.Err = &copy
		case Level:
			e.Severity = arg
		default:
			panic(fmt.Sprintf("unknown error argument: %v", arg))
		}
	}
	prev, ok := e.Err.(*Error)
	if !ok {
		return e
	}
	// The previous error was also one of ours. Suppress duplications
	// so the message won't contain the same kind of error twice, also
	// the severity is the highest of the two.
	if prev.Kind == e.Kind {
		e.Kind = Other
	}
	if e.Kind == Other {
		e.Kind = prev.Kind
		prev.Kind = Other
	}
	if prev.Severity > e.Severity {
		e.Severity = prev.Severity
	}
	return e
}

// Error returns the error string.
func (e *Error) Error() string {
	b := new(bytes.Buffer)
	if e.Op != "" {
		pad(b, ": ")
		b.WriteString(string(e.Op))
	}
	if e.Kind != 0 {
		pad(b, ": ")
		b.WriteString(e.Kind.String())
	}
	if e.Err != nil {
		// Indent on new line if we are cascading non-empty Upspin errors.
		if prevErr, ok := e.Err.(*Error); ok {
			if !prevErr.isZero() {
				pad(b, Separator)
				b.WriteString(e.Err.Error())
			}
		} else {
			pad(b, ": ")
			b.WriteString(e.Err.Error())
		}
	}
	if b.Len() == 0 {
		return "no error"
	}
	return b.String()
}

// pad appends str to the buffer if the buffer already has some data.
func pad(b *bytes.Buffer, str string) {
	if b.Len() == 0 {
		return
	}
	b.WriteString(str)
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

// Is reports whether err is an *Error of the given Kind.
// If err is nil then Is returns false.
func Is(kind Kind, err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	if e.Kind != Other {
		return e.Kind == kind
	}
	if e.Err != nil {
		return Is(kind, e.Err)
	}
	return false
}
