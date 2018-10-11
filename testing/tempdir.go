package test

import (
	"io/ioutil"
	"os"
	"testing"
)

// ScratchDir ...
// Create a temporary directory.
// Call the returned closure to cleanup.
func ScratchDir(t *testing.T) (string, func()) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	return d, func() { os.RemoveAll(d) }
}
