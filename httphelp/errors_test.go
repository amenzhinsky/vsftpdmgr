package httphelp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerFunc(t *testing.T) {
	t.Parallel()

	h := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return &HTTPError{Code: http.StatusBadRequest, Err: errors.New("foo")}
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(w, r)

	testResponseCode(t, w, http.StatusBadRequest)
	testBodyContains(t, w, "foo")
}
