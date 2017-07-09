package httputil

import (
	"log"
	"net/http"
	"time"
)

// logResponseWriter is a middleware that logs incoming HTTP requests.
type logResponseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader implements http.ResponseWriter interface.
func (w *logResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Log is a logging middleware.
func Log(h HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		n := time.Now()
		l := &logResponseWriter{ResponseWriter: w, status: http.StatusOK}
		if err := h(l, r); err != nil {
			return err
		}

		// TODO: test what happens when WriteHeader or Write haven't called here
		log.Printf("%s %s code=%d took=%s", r.Method, r.URL, l.status, time.Since(n).String())
		return nil
	}
}
