package httputil

import (
	"encoding/json"
	"io/ioutil"

	"net/http"
)

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

func WriteJSON(w http.ResponseWriter, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b)

	return nil
}
