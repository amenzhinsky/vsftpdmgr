package httphelp

import (
	"fmt"
	"net/http"
	"os"
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

// LogSink is where logs are written to.
var LogSink = os.Stdout

// Log is a logging middleware.
func Log(h HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if LogSink == nil {
			panic("LogSink is nil, use io.Discard to silence logging")
		}

		n := time.Now()
		l := &logResponseWriter{ResponseWriter: w, status: http.StatusOK}
		if err := h(l, r); err != nil {
			return err
		}

		_, err := fmt.Fprintf(LogSink, "%s %s %s code=%d took=%s\n",
			n.Format(time.RFC3339),
			r.Method,
			r.URL,
			l.status,
			time.Since(n).String(),
		)

		if err != nil {
			fmt.Fprintf(os.Stderr, "httphelp log write error: %v\n", err)
		}
		return nil
	}
}
