package backoff

type ErrorAlreadyRunning struct{}

func (e ErrorAlreadyRunning) Error() string {
	return "Backoff Already Running"
}
