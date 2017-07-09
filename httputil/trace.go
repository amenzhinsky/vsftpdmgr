package httputil

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
)

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
