package httphelp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// ReadJSON reads out JSON-encoded HTTP request, parses it
// and stores the result is into the value where v points.
func ReadJSON(r *http.Request, v interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, v); err != nil {
		return &HTTPError{Code: http.StatusBadRequest, Err: err}
	}
	return nil
}

// WriteJSON writes the JSON encoding of v into w.
func WriteJSON(w http.ResponseWriter, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Fprintf(os.Stderr, "http write error: %v", err)
	}
	return nil
}
