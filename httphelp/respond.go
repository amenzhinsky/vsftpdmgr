package httphelp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Bind reads out JSON-encoded HTTP request, parses it
// and stores the result is into the value where v points.
func Bind(r *http.Request, v interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, v); err != nil {
		return &HTTPError{Code: http.StatusBadRequest, Err: err}
	}
	return nil
}

// JSON writes the JSON encoding of v into w.
func JSON(w http.ResponseWriter, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Fprintf(os.Stderr, "httphelp.JSON write error: %v", err)
	}
	return nil
}

// Text responds with the provided HTTP status code and body string.
func Text(w http.ResponseWriter, code int, body string) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	if _, err := w.Write([]byte(body)); err != nil {
		fmt.Fprintf(os.Stderr, "httphelp.Text write error: %v", err)
	}
	return nil
}

// Empty responds only with the provided status code and no content.
func Empty(w http.ResponseWriter, code int) error {
	w.WriteHeader(code)
	return nil
}

// Redirect redirects browser to the url.
func Redirect(w http.ResponseWriter, code int, url string) error {
	if code < 300 || code >= 400 {
		panic("invalid redirect status code")
	}
	w.Header().Set("Location", url)
	w.WriteHeader(code)
	return nil
}
