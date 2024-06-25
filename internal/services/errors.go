package services

type AlreadyExistsError struct {
	msg string
}

func (err *AlreadyExistsError) Error() string {
	return err.msg
}

type InvalidArgumentError struct {
	msg string
}

func (err *InvalidArgumentError) Error() string {
	return err.msg
}

type NotFoundError struct {
	msg string
}

func (err *NotFoundError) Error() string {
	return err.msg
}
