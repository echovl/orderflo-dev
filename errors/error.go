package errors

import (
	"errors"
	"fmt"
)

type Kind int

const (
	KindUnexpected Kind = iota + 1
	KindValidation
	KindNotFound
	KindAuthentication
)

type Error struct {
	Kind Kind
	Err  error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func E(args ...any) error {
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Kind:
			e.Kind = arg
		case error:
			e.Err = arg
		case string:
			e.Err = errors.New(arg)
		default:
			panic("bad call to errors.E")
		}
	}
	return e
}

func NotFound(args ...any) error {
	args = append(args, KindNotFound)
	return E(args...)
}

func Authentication(args ...any) error {
	args = append(args, KindAuthentication)
	return E(args...)
}

func Unexpected(args ...any) error {
	args = append(args, KindUnexpected)
	return E(args...)
}

func Validation(args ...any) error {
	args = append(args, KindValidation)
	return E(args...)
}

func Is(err error, target any) bool {
	switch t := target.(type) {
	case error:
		return errors.Is(err, t)
	case Kind:
		e, ok := err.(*Error)
		if !ok {
			return target == KindUnexpected
		}
		if e.Kind != 0 {
			return e.Kind == target
		}
		return Is(e.Err, target)
	default:
		panic("errors: bad Is usage, unsupported target type")
	}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// Errorf is equivalent to fmt.Errorf, but allows clients to import only this
// package for all error handling.
func Errorf(format string, args ...any) error {
	return &errorString{fmt.Sprintf(format, args...)}
}
