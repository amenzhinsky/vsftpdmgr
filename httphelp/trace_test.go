package httphelp

import (
	"testing"
	"net/http/httptest"
	"net/http"
)

func TestTrace(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h := Trace(func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})

	if err := h(w, r); err != nil {
		t.Fatal(err)
	}

	// TODO: test output
}
