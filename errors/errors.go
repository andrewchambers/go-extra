package errors

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
)

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

const MAX_FRAME_SIZE = 200

type Frame struct {
	OurError    bool
	Values      []KV
	DroppedInfo bool
	SourceHost  string
	SourceFile  string
	SourceLine  int
	Depth       uint64
}

type Trace struct {
	Frames []Frame
}

type KV struct {
	K string
	V string
}

type Error struct {
	// Error context, intended to be immutable,
	// Never mutate values stored in the context.
	// Never modify this in place.
	Values []KV

	// Program counter of where this error originates.
	SourcePC uintptr

	WasDerserialized bool
	SourceHost       string
	SourceFile       string
	SourceLine       int

	Depth       uint64
	DroppedInfo bool
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

func getDepth(err error) uint64 {
	e, ok := err.(*Error)
	if !ok {
		return 0
	}
	return e.Depth
}

func New(msg string) error {
	var pc [1]uintptr
	runtime.Callers(2, pc[:])

	return &Error{
		SourcePC:  pc[0],
		Values:    []KV{KV{K: "msg", V: msg}},
		Cause:     nil,
		RootCause: nil,
		Depth:     0,
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
		Depth:     0,
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
		values = append(values, KV{K: "msg", V: fmt.Sprintf("%v", args[0])})
	}
	args = args[1:]

	for i := 0; i < len(args); i += 2 {
		k, ok := args[i].(string)
		if !ok {
			continue
		}
		v := args[i+1]
		values = append(values, KV{K: k, V: fmt.Sprintf("%v", v)})
	}

	depth := getDepth(err) + 1
	droppedInfo := false
	cause := err
	rootCause := RootCause(err)
	if depth > MAX_FRAME_SIZE {
		cause = Cause(cause)
		droppedInfo = true
	}

	return &Error{
		SourcePC:    pc[0],
		Values:      values,
		Cause:       cause,
		RootCause:   rootCause,
		Depth:       depth,
		DroppedInfo: droppedInfo,
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	var pc [1]uintptr
	runtime.Callers(2, pc[:])

	depth := getDepth(err) + 1
	droppedInfo := false
	cause := err
	rootCause := RootCause(err)
	if depth > MAX_FRAME_SIZE {
		cause = Cause(cause)
		droppedInfo = true
	}

	return &Error{
		SourcePC:    pc[0],
		Values:      []KV{KV{K: "msg", V: fmt.Sprintf(format, args...)}},
		Cause:       cause,
		RootCause:   rootCause,
		Depth:       depth,
		DroppedInfo: droppedInfo,
	}
}

// Return an error trace up to the maximum of 10000 frames.
func GetTrace(err error) *Trace {
	t := &Trace{}

	for i := 0; i < MAX_FRAME_SIZE; i++ {
		if err == nil {
			return t
		}
		e, ok := err.(*Error)
		if !ok {
			t.Frames = append(t.Frames, Frame{
				Values:   []KV{KV{K: "msg", V: err.Error()}},
				OurError: false,
			})
			return t
		}

		if e.WasDerserialized {
			t.Frames = append(t.Frames, Frame{
				OurError:    true,
				DroppedInfo: e.DroppedInfo,
				SourceHost:  e.SourceHost,
				SourceFile:  e.SourceFile,
				SourceLine:  e.SourceLine,
				Depth:       e.Depth,
			})
		} else {
			fn := runtime.FuncForPC(e.SourcePC)
			file, line := fn.FileLine(e.SourcePC)
			t.Frames = append(t.Frames, Frame{
				OurError:    true,
				DroppedInfo: e.DroppedInfo,
				SourceHost:  hostname,
				SourceFile:  file,
				SourceLine:  line,
				Depth:       e.Depth,
			})
		}

		err = e.Cause
	}

	return t
}

// Return the next error cause of an error if possible.
func Cause(err error) error {
	if err == nil {
		return nil
	}

	e, ok := err.(*Error)
	if ok {
		if e.Cause == nil {
			return err
		}

		return e.Cause
	}

	return err
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

	for _, f := range t.Frames {
		if f.OurError {
			_, _ = fmt.Fprintf(&buf, "%s:%s:%d\n", f.SourceHost, f.SourceFile, f.SourceLine)
		} else {
			_, _ = fmt.Fprintf(&buf, "???:???:???\n")
		}
		if len(f.Values) != 0 {
			_, _ = fmt.Fprintf(&buf, "Where:\n")
			for _, kv := range f.Values {
				_, _ = fmt.Fprintf(&buf, "  %#v = %#v\n", kv.K, kv.V)
			}
		}

		if f.DroppedInfo {
			_, _ = fmt.Fprintf(&buf, "... Dropped info ...\n")
		}
	}

	return buf.String()
}
