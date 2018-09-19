package exec

import (
	"io"
	"io/ioutil"
	"os/exec"

	extraio "github.com/andrewchambers/go-extra/io"
)

// Set cmd.Stderr to io.Discard
// set cmd.Stdout and cmd.Stdin to pipes connected
// to the returned read write closer.
func CmdReadWriteCloser(cmd *exec.Cmd) *extraio.MergedReadWriteCloser {
	a, b := io.Pipe()
	x, y := io.Pipe()

	cmd.Stderr = ioutil.Discard
	cmd.Stdout = b
	cmd.Stdin = x

	rwc := &extraio.MergedReadWriteCloser{
		RC: a,
		WC: y,
	}

	return rwc
}
