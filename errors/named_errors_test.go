package errors

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
