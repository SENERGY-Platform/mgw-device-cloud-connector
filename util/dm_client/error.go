package dm_client

type cError struct {
	err error
}

type InternalError struct {
	cError
}

type NotFoundError struct {
	cError
}

func (e *cError) Error() string {
	return e.err.Error()
}

func (e *cError) Unwrap() error {
	return e.err
}

func newInternalError(err error) error {
	return &InternalError{cError{err: err}}
}

func newNotFoundError(err error) error {
	return &NotFoundError{cError{err: err}}
}
