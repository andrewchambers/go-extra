package main

import (
	"fmt"
	"io"
	"os"

	"github.com/andrewchambers/go-extra/errors"
)

func A(a int) error {
	err := B(a + 1)
	if err != nil {
		return errors.Wrap(err, "A failed", "a", a)
	}

	return nil
}

func B(b int) error {
	err := C()
	if err != nil {
		return errors.Wrap(err, "B failed", "b", b)
	}
	return nil
}

func C() error {
	return errors.Wrap(io.EOF)
}

func main() {
	err := A(4)
	fmt.Printf("%s\n\n", err.Error())
	fmt.Printf("error trace:\n\n")
	fmt.Printf("%s", errors.GetTrace(err))
	os.Exit(1)
}
