package errors

import (
	"bytes"
	"fmt"
	"runtime"
)

type Frame struct {
	PCKnown bool
	Values  []KV
	PC      uintptr
}

type Trace struct {
	Frames []Frame
}

type KV struct {
	K string
	V interface{}
}

type Error struct {
	// A message for th eend user.
	Message string
	// Error context, intended to be immutable,
	// Never mutate values stored in the context.
	// Never modify this in place.
	Values []KV
	// Program counter of where this error originates.
	SourcePC uintptr
	// The original cause of the error, if nil,
	// the error itself is the cause,
	Cause error
	// Root cause of the error chain.
	RootCause error
}

func (err *Error) String() string {
	return err.Error()
}

func (err *Error) Error() string {
	rootCause := RootCause(err)

	msg := "error"

	msgVal, ok := err.LookupValue("msg")
	if ok {
		msgStr, ok := msgVal.(string)
		if ok {
			msg = msgStr
		}
	}

	if rootCause == nil || err.Cause == nil {
		return msg
	}

	return fmt.Sprintf("%s: %s", msg, rootCause.Error())
}

func (err *Error) LookupValue(k string) (interface{}, bool) {
	for _, kv := range err.Values {
		if kv.K == k {
			return kv.V, true
		}
	}

	return nil, false
}

func New(msg string) error {
	var pc [1]uintptr
	runtime.Callers(2, pc[:])

	return &Error{
		SourcePC:  pc[0],
		Values:    []KV{KV{K: "msg", V: msg}},
		Cause:     nil,
		RootCause: nil,
	}
}

func Errorf(format string, args ...interface{}) error {

	var pc [1]uintptr
	runtime.Callers(2, pc[:])

	return &Error{
		SourcePC:  pc[0],
		Values:    []KV{KV{K: "msg", V: fmt.Sprintf(format, args...)}},
		Cause:     nil,
		RootCause: nil,
	}
}

func Wrap(err error, args ...interface{}) error {
	if err == nil {
		return nil
	}

	var pc [1]uintptr
	runtime.Callers(2, pc[:])
	values := []KV{}

	if len(args) == 0 {
		args = []interface{}{"error"}
	}

	if len(args)%2 != 0 {
		values = append(values, KV{K:"msg", V:args[0]})
	}
	args = args[1:]

	for i := 0; i < len(args); i += 2 {
		k, ok := args[i].(string)
		if !ok {
			continue
		}
		v := args[i+1]
		values = append(values, KV{K: k, V: v})
	}

	return &Error{
		SourcePC:  pc[0],
		Values:    values,
		Cause:     err,
		RootCause: RootCause(err),
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	var pc [1]uintptr
	runtime.Callers(2, pc[:])

	return &Error{
		SourcePC:  pc[0],
		Values:    []KV{KV{K: "msg", V: fmt.Sprintf(format, args...)}},
		Cause:     err,
		RootCause: RootCause(err),
	}
}

// Return an error trace up to the maximum of 10000 frames.
func GetTrace(err error) *Trace {
	t := &Trace{}

	for i := 0; i < 10000; i++ {
		if err == nil {
			return t
		}
		e, ok := err.(*Error)
		if !ok {
			t.Frames = append(t.Frames, Frame{
				Values:  []KV{KV{K: "msg", V: err.Error()}},
				PCKnown: false,
			})
			return t
		}

		t.Frames = append(t.Frames, Frame{
			PCKnown: true,
			PC:      e.SourcePC,
			Values:  e.Values,
		})
		err = e.Cause
	}

	return t
}

// Return the original error cause of an error if possible.
func RootCause(err error) error {
	if err == nil {
		return nil
	}

	e, ok := err.(*Error)
	if ok {
		if e.RootCause != nil {
			return e.RootCause
		}
	}

	return err
}

func (t *Trace) String() string {
	var buf bytes.Buffer

	hasPrev := false
	for _, f := range t.Frames {
		if hasPrev {
			_, _ = fmt.Fprintf(&buf, "Cause:\n")
		}
		hasPrev = true
		if f.PCKnown {
			fn := runtime.FuncForPC(f.PC)
			file, line := fn.FileLine(f.PC)
			_, _ = fmt.Fprintf(&buf, "%s:%s:%d\n", file, fn.Name(), line)
		} else {
			_, _ = fmt.Fprintf(&buf, "???:???:???\n")
		}
		_, _ = fmt.Fprintf(&buf, "Where:\n")
		for _, kv := range f.Values {
			_, _ = fmt.Fprintf(&buf, "  %#v = %#v\n", kv.K, kv.V)
		}
	}

	return buf.String()
}
