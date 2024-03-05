package cloud_client

type cError struct {
	err error
}

type InternalError struct {
	cError
}

type NotFoundError struct {
	cError
}

type UnauthorizedError struct {
	cError
}

func (e *cError) Error() string {
	return e.err.Error()
}

func (e *cError) Unwrap() error {
	return e.err
}

func NewInternalError(err error) error {
	return &InternalError{cError{err: err}}
}

func NewNotFoundError(err error) error {
	return &NotFoundError{cError{err: err}}
}

func NewUnauthorizedError(err error) error {
	return &UnauthorizedError{cError{err: err}}
}
