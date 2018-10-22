package errors

type ErrorCode uint64

const AnonymousErrorCode ErrorCode = 0

// All registered named errors. Errors are
// automatically registered with with NewNamed.
var NamedErrors map[ErrorCode]*NamedError = make(map[ErrorCode]*NamedError)

// A named error is a globally identified error.
// The code can  be used to traverse network boundaries and
// the meaning should never change. Generate the id randomly.
// If an error code is registered twice, the program will panic
// on init. The chance of collision in a program is low.
type NamedError struct {
	msg  string
	code ErrorCode
}

func (ne *NamedError) Error() string {
	return ne.msg
}

func (ne *NamedError) Code() ErrorCode {
	return ne.code
}

func GetCode(err error) ErrorCode {
	ne, ok := err.(*NamedError)
	if !ok {
		return AnonymousErrorCode
	}

	return ne.code
}

func NewNamed(msg string, code ErrorCode) *NamedError {
	if code == AnonymousErrorCode {
		panic("error code 0 is reserved to mean anonymous error")
	}
	err := &NamedError{
		msg:  msg,
		code: code,
	}
	NamedErrors[code] = err
	return err
}
