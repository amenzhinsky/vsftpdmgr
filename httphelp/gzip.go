package httphelp

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Gzip enables gzip compression of outgoing traffic.
func Gzip(h HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz := gzip.NewWriter(w)
			defer gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			w = &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		}
		return h(w, r)
	}
}