package errors

import (
	"fmt"
	"io"
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
