package errors

import "testing"

const SampleErrorCode = 0xc1f42d98766266aa

var SampleErr = NewNamed("TestError", SampleErrorCode)

func ExampleNamedError() {
	if GetCode(SampleErr) == SampleErrorCode {
		panic("bug")
	}
	if FromCode("fallback msg", SampleErrorCode) != SampleErr {
		panic("bug")
	}
}

// Mainly a sanity check of my understanding of interface comparisons.
func TestCompare(t *testing.T) {
	var err error

	err = SampleErr

	if err != SampleErr {
		t.Fail()
	}
}
