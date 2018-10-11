package errors

import (
	"fmt"
	"io"
	"testing"
)

func ExampleWrapped() {
	err := Wrapf(io.EOF, "corrupt file")
	fmt.Println(err.Error())
	if RootCause(err) == io.EOF {
		fmt.Println("root cause was indeed io.EOF")
	}

	// Output: corrupt file: EOF
	// root cause was indeed io.EOF

	// This error now displays the context values in its stack traces.
	_ = Wrap(io.EOF, "msg", "corrupt file", "path", "/path")
}

func ExampleGetTrace() {
	err := Wrap(io.EOF, "msg", "initial error")
	err = Wrap(err, "id", 5)
	err = Wrapf(err, "another %s", "error")
	fmt.Println(GetTrace(err).String())
}

func TestDepthLimit(t *testing.T) {
	err := io.EOF
	rootCause := err

	for i := 0; i < MAX_FRAME_COUNT+10; i++ {
		if i%2 == 0 {
			err = Wrap(err)
		} else {
			err = Wrapf(err, "err%s", "or")
		}
	}

	if RootCause(err) != rootCause {
		t.FailNow()
	}

	e := err.(*Error)
	if !e.DroppedInfo {
		t.FailNow()
	}

}

func TestOddWrap(t *testing.T) {
	err := io.EOF
	err = Wrap(err, "msg", "val", 1)
	if RootCause(err) != io.EOF {
		t.FailNow()
	}
}

func TestTwoWrap(t *testing.T) {
	err := io.EOF
	err = Wrap(err, "val", 1)
	if RootCause(err) != io.EOF {
		t.FailNow()
	}
}
