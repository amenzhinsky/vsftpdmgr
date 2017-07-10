package httphelp

import (
	"fmt"
	"log"
	"net/http"
)

// Handler is equivalent of http.Handler but ServeHTTP returns an error.
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) error
}

// HandlerFunc is equivalent of http.HandlerFunc but it returns an error.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

// HTTPError is an error based on a HTTP response code.
type HTTPError struct {
	Code int
	Err  error
}

// HTTPError implements error interface.
func (e *HTTPError) Error() string {
	var i string
	if e.Err != nil {
		i = ": " + e.Err.Error()
	}
	return fmt.Sprintf("%s%s", http.StatusText(e.Code), i)
}

// Common HTTP errors.
var (
	ErrBadRequest                = &HTTPError{Code: http.StatusBadRequest}
	ErrMethodNotAllowed          = &HTTPError{Code: http.StatusMethodNotAllowed}
	ErrStatusUnauthorized        = &HTTPError{Code: http.StatusUnauthorized}
	ErrStatusUnprocessableEntity = &HTTPError{Code: http.StatusUnprocessableEntity}
	ErrInternalServerError       = &HTTPError{Code: http.StatusInternalServerError}
)

// Wrap transforms a Handler into a http.Handler,
// errors are passed to the ErrorHandler.
func Wrap(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h.ServeHTTP(w, r); err != nil {
			if ErrorHandler == nil {
				panic("ErrorHandler is nil")
			}
			ErrorHandler(w, r, err)
		}
	})
}

// WrapFunc transforms a HandlerFunc into a http.HandlerFunc,
// errors are passed to the ErrorHandler.
func WrapFunc(f HandlerFunc) http.HandlerFunc {
	return Wrap(f).(http.HandlerFunc)
}

// ErrorHandler handlers errors returned from Handler's and HandlerFunc's.
var ErrorHandler = DefaultErrorHandler

// ErrorHandler handles errors returned from a HandlerFunc.
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

// DefaultErrorHandler logs errors using the log package and replies
// to the request with the Internal Server Error code and message.
// HTTPError instances are not logged.
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	e, ok := err.(*HTTPError)
	if !ok {
		e = ErrInternalServerError
		log.Printf("[ERR] %s %s %v", r.Method, r.URL, err)
	}
	http.Error(w, err.Error(), e.Code)
}
