package httphelp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadJSON(t *testing.T) {
	t.Parallel()

	s := map[string]string{}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"foo": "bar"}`))
	if err := ReadJSON(r, &s); err != nil {
		t.Fatal(err)
	}

	if s["foo"] != "bar" {
		t.Errorf("ReadJSON(r, s): s.Foo = %q, want %q", s["foo"], "bar")
	}
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	s := map[string]string{"foo": "bar"}
	w := httptest.NewRecorder()
	if err := WriteJSON(w, &s); err != nil {
		t.Fatal(err)
	}

	if w.Body.String() != `{"foo":"bar"}` {
		t.Errorf("WriteJSON(w, s); body = %q, want %q", w.Body.String(), `{"foo":"bar"}`)
	}
}
