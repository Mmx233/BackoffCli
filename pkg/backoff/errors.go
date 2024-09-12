package backoff

import "fmt"

type ErrorAlreadyRunning struct{}

func (e ErrorAlreadyRunning) Error() string {
	return "Backoff Already Running"
}

type ErrorPanic struct {
	Err   any
	Stack string
}

func (e ErrorPanic) Error() string {
	return fmt.Sprintf("panic: %v\n%s", e.Err, e.Stack)
}
