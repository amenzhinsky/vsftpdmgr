package httphelp

import (
	"log"
	"net/http"
)

// HandlerFunc is an http.Handle compatible object,
// but it returns an error opposed to the http.HandlerFunc.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP implements the http.Handler interface.
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil && ErrorHandler != nil {
		if ErrorHandler == nil {
			panic("ErrorHandler is nil")
		}
		ErrorHandler(w, r, err)
	}
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
	return http.StatusText(e.Code) + i
}

// Common HTTP errors.
var (
	ErrBadRequest                = &HTTPError{Code: http.StatusBadRequest}
	ErrMethodNotAllowed          = &HTTPError{Code: http.StatusMethodNotAllowed}
	ErrStatusUnauthorized        = &HTTPError{Code: http.StatusUnauthorized}
	ErrStatusUnprocessableEntity = &HTTPError{Code: http.StatusUnprocessableEntity}
	ErrInternalServerError       = &HTTPError{Code: http.StatusInternalServerError}
)

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
