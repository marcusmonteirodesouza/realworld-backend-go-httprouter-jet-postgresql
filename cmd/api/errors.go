package main

type forbiddenError struct {
	msg string
}

func (err *forbiddenError) Error() string {
	return err.msg
}

type malformedRequest struct {
	msg string
}

func (err *malformedRequest) Error() string {
	return err.msg
}

type unauthorizedError struct {
	msg string
}

func (err *unauthorizedError) Error() string {
	return err.msg
}
