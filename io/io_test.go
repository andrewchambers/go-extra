package io

import (
	"io/ioutil"
	"testing"
)

func TestLimitedWriter(t *testing.T) {

	w := &LimitedWriter{
		W: ioutil.Discard,
		N: 10,
	}

	n, err := w.Write(make([]byte, 11))
	if err != ErrOutOfSpace {
		t.Fatal(err)
	}
	if n != 0 {
		t.FailNow()
	}

	n, err = w.Write(make([]byte, 6))
	if err != nil {
		t.Fatal(err)
	}
	if n != 6 || w.N != 4 {
		t.FailNow()
	}

	n, err = w.Write(make([]byte, 4))
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 || w.N != 0 {
		t.FailNow()
	}
}
