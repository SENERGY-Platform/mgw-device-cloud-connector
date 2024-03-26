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

type BadRequestError struct {
	cError
}

type ForbiddenError struct {
	cError
}

type NotAllowedError struct {
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

func newUnauthorizedError(err error) error {
	return &UnauthorizedError{cError{err: err}}
}

func newBadRequestError(err error) error {
	return &BadRequestError{cError{err: err}}
}

func newForbiddenError(err error) error {
	return &ForbiddenError{cError{err: err}}
}

func newNotAllowedError(err error) error {
	return &NotAllowedError{cError{err: err}}
}
