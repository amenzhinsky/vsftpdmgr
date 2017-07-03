package mgr

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCRUD(t *testing.T) {
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

	m, err := New(root, pwdfile.Name(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		m.Clean()

		cerr := m.Close()
		if err == nil && cerr != nil {
			t.Fatal(err)
		}
	}()

	// create
	user := &User{Username: "test", Password: "notverysecurepassword"}
	if err := m.Save(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	testFileContains(t, pwdfile, user.Username)
	testListContains(t, m, user)
	testLocalRootExists(t, root, user.Username)

	// update
	user.Password = "verysecurepassword"
	if err := m.Save(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	testFileContains(t, pwdfile, user.Username)
	testListContains(t, m, user)
	testLocalRootExists(t, root, user.Username)

	// delete
	if err := m.Delete(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	testFileDoesntContain(t, pwdfile, user.Username)
	testListDoesntContain(t, m, user)
	testLocalRootDoesntExists(t, root, user.Username)
}

func testFileContains(t *testing.T, f *os.File, s string) {
	if !fileContains(t, f, s) {
		t.Errorf("file expected to contain %q but it doesn't", s)
	}
}

func testFileDoesntContain(t *testing.T, f *os.File, s string) {
	if fileContains(t, f, s) {
		t.Errorf("file expected not to contain %q but it does", s)
	}
}

func fileContains(t *testing.T, f *os.File, s string) bool {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Contains(string(b), s)
}

func testListContains(t *testing.T, m *Mgr, user *User) {
	if !listContains(t, m, user) {
		t.Errorf("List: expected to contain %v but it doesn't", user)
	}
}

func testListDoesntContain(t *testing.T, m *Mgr, user *User) {
	if listContains(t, m, user) {
		t.Errorf("List: expected not to contain %v but it does", user)
	}
}

func listContains(t *testing.T, m *Mgr, user *User) bool {
	users, err := m.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range users {
		if user.Username == u.Username {
			return true
		}
	}
	return false
}

func testLocalRootExists(t *testing.T, root, username string) {
	if !localRootExists(t, root, username) {
		t.Errorf("local root for user %s doesn't exist", username)
	}
}

func testLocalRootDoesntExists(t *testing.T, root, username string) {
	if localRootExists(t, root, username) {
		t.Errorf("local root for user %s exists, but it's not expected to", username)
	}
}

func localRootExists(t *testing.T, root, username string) bool {
	f, err := os.Lstat(filepath.Join(root, username))
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		t.Fatal(err)
	}
	return f.IsDir()
}
