package backoff

import "fmt"

type ErrorAlreadyRunning struct{}

func (e ErrorAlreadyRunning) Error() string {
	return "backoff Already Running"
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

type ErrorUnexpectedHttpStatus struct {
	HttpStatus int
}

func (e ErrorUnexpectedHttpStatus) Error() string {
	return fmt.Sprintf("unexpected http status: %v", e.HttpStatus)
}

type ErrorKeywordNotFound struct {
	Keyword string
}

func (e ErrorKeywordNotFound) Error() string {
	return fmt.Sprintf("keyword '%s' not found", e.Keyword)
}
