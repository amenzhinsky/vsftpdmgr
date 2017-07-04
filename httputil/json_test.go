package httputil

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"strings"
)

func TestReadJSON(t *testing.T) {
	t.Parallel()

	var s struct {
		Foo string `json:"foo"`
	}

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"foo": "bar"}`))
	if err := ReadJSON(r, &s); err != nil {
		t.Fatal(err)
	}

	if s.Foo != "bar" {
		t.Errorf("ReadJSON(r, s): s.Foo = %q, want %q", s.Foo, "bar")
	}
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	s := struct {
		Foo string `json:"foo"`
	}{Foo: "bar"}

	w := httptest.NewRecorder()
	if err := WriteJSON(w, &s); err != nil {
		t.Fatal(err)
	}

	if w.Body.String() != `{"foo":"bar"}` {
		t.Errorf("WriteJSON(w, s); body = %q, want %q", w.Body.String(), `{"foo":"bar"}`)
	}
}
