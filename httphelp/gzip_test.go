package httphelp

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzip(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Encoding", "gzip, deflate")
	h := Gzip(func(w http.ResponseWriter, r *http.Request) error {
		return Text(w, http.StatusOK, "foo")
	})

	if err := h(w, r); err != nil {
		t.Fatal(err)
	}

	gz, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "foo" {
		t.Errorf("gzip body = %q, want %q", string(b), "foo")
	}
}
