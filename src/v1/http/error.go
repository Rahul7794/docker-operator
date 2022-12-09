package http

import (
	"errors"
	"fmt"
)

type HttpError struct {
	err  error
	code int
}

// HTTPError creates an HttpError error object.
// err can be a string or of type error.
// wrapMsg can be used to provide extra context to the error.
func HTTPError(err interface{}, statusCode int, wrapMsg ...string) HttpError {
	var e error
	switch err.(type) {
	case nil:
		e = nil
	case error:
		e = err.(error)
	case string:
		e = errors.New(err.(string))
	default:
		panic("err must be an error or string")
	}

	if len(wrapMsg) == 0 {
		return HttpError{e, statusCode}
	}
	return HttpError{fmt.Errorf("%s: %s", wrapMsg[0], e), statusCode}
}

func (h HttpError) StatusCode() int { return h.code }
func (h HttpError) Unwrap() error   { return h.err }
func (h HttpError) Error() string {
	if h.err == nil {
		return "<nil>"
	}
	return h.err.Error()
}
