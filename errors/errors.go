package errors

import (
	"fmt"
	"io"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
)

type prettyerror struct {
	msg string
	*stack
}

// Error returns error message
func (p *prettyerror) Error() string {
	return p.msg
}

// Format formats error
func (p *prettyerror) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, p.msg)
			p.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		if s.Flag('+') {
			io.WriteString(s, p.msg)
			p.stack.Format(s, verb)
			return
		}

		io.WriteString(s, p.msg)
	case 'q':
		fmt.Fprintf(s, "%q", p.msg)
	}
}

type wrappederror struct {
	msg   string
	cause error
	*stack
}

// Error returns error message
func (w *wrappederror) Error() string {
	return w.msg + ": " + w.cause.Error()
}

func (w *wrappederror) Cause() error { return w.cause }

// Format formats error
func (w *wrappederror) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, w.msg+": "+w.cause.Error())
			w.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		if s.Flag('+') {
			io.WriteString(s, w.msg+": "+w.cause.Error())
			w.stack.Format(s, verb)
			return
		}

		io.WriteString(s, w.msg+": "+w.cause.Error())
	case 'q':
		fmt.Fprintf(s, "%q", w.msg+": "+w.cause.Error())
	}
}

type stack []uintptr

func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := errors.Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}

	case 's':
		if st.Flag('+') {
			frames := runtime.CallersFrames(*s)
			more := true
			var frame runtime.Frame
			for more {
				frame, more = frames.Next()
				io.WriteString(st, " func "+frame.Function+" in "+frame.File+" on line "+strconv.Itoa(frame.Line))
				return
			}
		}
	}
}

func (s *stack) StackTrace() errors.StackTrace {
	f := make([]errors.Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = errors.Frame((*s)[i])
	}
	return f
}

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(message string) error {
	return &prettyerror{
		msg:   message,
		stack: callers(),
	}
}

// Errorf is
func Errorf(format string, args ...interface{}) error {
	return &prettyerror{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(),
	}
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	return &wrappederror{
		cause: err,
		msg:   message,
		stack: callers(),
	}
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is called, and the format specifier.
// If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return &wrappederror{
		cause: err,
		msg:   fmt.Sprintf(format, args...),
		stack: callers(),
	}
}

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}
