package httputil

import (
	"log"
	"net/http"
)

type logResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *logResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Log is a logging middleware.
func Log(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ww := &logResponseWriter{ResponseWriter: w}
		h(ww, r)
		log.Printf("%s %s status=%d", r.Method, r.URL, ww.status)
	}
}
