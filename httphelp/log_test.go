package httphelp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLog(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h := Log(func(w http.ResponseWriter, r *http.Request) error {
		return Text(w, http.StatusOK, "foo")
	})

	if err := h(w, r); err != nil {
		t.Fatal(err)
	}

	// TODO: test log output
}
