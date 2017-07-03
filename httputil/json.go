package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// JSON writes marshal v and sends it as a JSON response.
// If returned err is a nil response's sent successfully.
// Write errors are logged to STDERR.
func JSON(w http.ResponseWriter, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "httputil write error: %v", err)
	}
	return nil
}
