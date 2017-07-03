package main

import (
	"fmt"
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

	rs, err := request(http.MethodPost, ts.URL+"/users", strings.NewReader(`{
		"username": "test",
		"password": "test"
	}`))

	if err != nil {
		t.Fatal(err)
	}

	if rs.StatusCode != http.StatusOK {
		t.Fatalf("POST /users code = %d, want %d", rs.StatusCode, http.StatusOK)
	}

	rs, err = request(http.MethodGet, ts.URL+"/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))
}

func request(method, url string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(r)
}
