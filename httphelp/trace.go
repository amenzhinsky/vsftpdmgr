package httphelp

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
)

// Trace outputs incoming requests info to STDERR.
func Trace(h HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		b, err := httputil.DumpRequest(r, true)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "%s\n", b)
		return h(w, r)
	}
}
