package fragment

import (
	"errors"
	"fmt"
	"strings"
)

// Error is an error when parsing a fragment
type Error struct {
	Path []string
	Err  error
}

func (e Error) Error() string {
	if e.Path == nil {
		return e.Err.Error()
	}
	return e.Err.Error() + " (" + strings.Join(e.Path, ".") + ")"
}

func (e Error) Register(path ...string) Error {
	if e.Path == nil {
		e.Path = path
	} else {
		e.Path = append(path, e.Path...)
	}
	return e
}

// NewError returns a new fragment error. If v is a fragment error, a copy is returned.
// If v is a string or error, that is set to the error
func NewError(v interface{}) Error {
	if vf, ok := v.(Error); ok {
		return vf
	} else if err, ok := v.(error); ok {
		return Error{Err: err}
	} else if text, ok := v.(string); ok {
		return Error{Err: errors.New(text)}
	} else {
		panic("unrecognized argument " + fmt.Sprint(v))
	}
}
