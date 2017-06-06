package makross

import (
	"errors"
	"fmt"
	"net/http"
)

// Errors
var (
	ErrUnsupportedMediaType        = NewHTTPError(StatusUnsupportedMediaType)
	ErrNotFound                    = NewHTTPError(StatusNotFound)
	ErrStatusBadRequest            = NewHTTPError(StatusBadRequest)
	ErrUnauthorized                = NewHTTPError(StatusUnauthorized)
	ErrForbidden                   = NewHTTPError(http.StatusForbidden)
	ErrMethodNotAllowed            = NewHTTPError(StatusMethodNotAllowed)
	ErrStatusRequestEntityTooLarge = NewHTTPError(StatusRequestEntityTooLarge)
	ErrRendererNotRegistered       = errors.New("renderer not registered")
	ErrInvalidRedirectCode         = errors.New("invalid redirect status code")
	ErrCookieNotFound              = errors.New("cookie not found")
)

// HTTPError represents an HTTP error with HTTP status code and error message
type HTTPError interface {
	error
	// StatusCode returns the HTTP status code of the error
	StatusCode() int
}

// Error contains the error information reported by calling Context.Error().
type httpError struct {
	Status  int    `json:"Status" xml:"Status"`
	Message string `json:"Message" xml:"Message"`
}

// NewHTTPError creates a new HttpError instance.
// If the error message is not given, http.StatusText() will be called
// to generate the message based on the status code.
func NewHTTPError(status int, message ...interface{}) HTTPError {
	he := httpError{Status: status, Message: StatusText(status)}
	if len(message) > 0 {
		he.Message = fmt.Sprint(message...)
	}
	return &he
}

// Error returns the error message.
func (e *httpError) Error() string {
	return e.Message
}

// StatusCode returns the HTTP status code.
func (e *httpError) StatusCode() int {
	return e.Status
}
