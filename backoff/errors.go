package backoff

import "fmt"

type ErrorAlreadyRunning struct{}

func (e ErrorAlreadyRunning) Error() string {
	return "Backoff Already Running"
}

type ErrorPanic struct {
	Reason any
	Stack  string
}

func (e ErrorPanic) Error() string {
	return fmt.Sprintf("panic: %v\n%s", e.Reason, e.Stack)
}

type ErrorMaxRetryExceeded struct {
	LastError error
}

func (e ErrorMaxRetryExceeded) Error() string {
	return "max retry exceeded"
}
