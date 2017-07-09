package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLog(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h := func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}

	h = Log(h)
	if err := h(w, r); err != nil {
		t.Fatal(err)
	}

	// TODO: test log output
}
