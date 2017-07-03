package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/amenzhinsky/vsftpdmgr/mgr"
)

func TestAll(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is empty")
	}

	root, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	pwdfile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		pwdfile.Close()
		os.Remove(pwdfile.Name())
	}()

	m, err := mgr.New(root, pwdfile.Name(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		m.Clean()
		m.Close()
	}()

	ts := httptest.NewServer(handler(m))
	defer ts.Close()

	rs := request(t, http.MethodPost, ts.URL+"/users", strings.NewReader(`{
		"username": "test",
		"password": "test"
	}`))

	if rs.StatusCode != http.StatusOK {
		t.Fatalf("POST /users code = %d, want %d", rs.StatusCode, http.StatusOK)
	}

	rs = request(t, http.MethodGet, ts.URL+"/users", nil)
	testResponseContains(t, rs, "test")
}

func request(t *testing.T, method, url string, body io.Reader) *http.Response {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	rs, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	return rs
}

func testResponseContains(t *testing.T, r *http.Response, s string) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(b), s) {
		t.Errorf("response %q doesn't contain %q", string(b), s)
	}
}
