package httphelp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBind(t *testing.T) {
	t.Parallel()

	s := map[string]string{}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"foo": "bar"}`))
	if err := Bind(r, &s); err != nil {
		t.Fatal(err)
	}

	if s["foo"] != "bar" {
		t.Errorf("Bind(r, s): s.Foo = %q, want %q", s["foo"], "bar")
	}
}

func TestJSON(t *testing.T) {
	t.Parallel()

	s := map[string]string{"foo": "bar"}
	w := httptest.NewRecorder()
	if err := JSON(w, &s); err != nil {
		t.Fatal(err)
	}

	if w.Body.String() != `{"foo":"bar"}` {
		t.Errorf("JSON(w, s); body = %q, want %q", w.Body.String(), `{"foo":"bar"}`)
	}
}

func TestText(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	if err := Text(w, http.StatusOK, "ok"); err != nil {
		t.Fatal(err)
	}

	testResponseCode(t, w, http.StatusOK)
	testBodyContains(t, w, "ok")
}

func TestEmpty(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	if err := Empty(w, http.StatusNotModified); err != nil {
		t.Fatal(err)
	}
	testResponseCode(t, w, http.StatusNotModified)
}

func TestRedirect(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	if err := Redirect(w, http.StatusFound, "/foo"); err != nil {
		t.Fatal(err)
	}
	testResponseCode(t, w, http.StatusFound)
}

func testResponseCode(t *testing.T, w *httptest.ResponseRecorder, want int) {
	if w.Code != want {
		t.Errorf("code = %d, want %d", w.Code, want)
	}
}

func testBodyContains(t *testing.T, w *httptest.ResponseRecorder, chunk string) {
	if !strings.Contains(w.Body.String(), chunk) {
		t.Errorf("body %q doesn't contain %q", w.Body.String(), chunk)
	}
}
