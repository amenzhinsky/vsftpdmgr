package httputil

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrap(t *testing.T) {
	t.Parallel()

	h := WrapFunc(func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("test error")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Error(w.Code, http.StatusInternalServerError)
	}
}
