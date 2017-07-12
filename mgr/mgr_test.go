package mgr

import (
	"context"
	"io/ioutil"
	"os"
	"os/user"
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

	usr, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}

	// create
	u := &User{
		Username: "test",
		Password: "insecurePassword",
		FS: FS{
			Mode:  0755,
			Owner: usr.Username,
			Group: usr.Username,
			Children: []FS{
				{
					Name:  "read",
					Mode:  0555,
					Owner: usr.Username,
					Group: usr.Username,
				},
			},
		},
	}
	if err := m.Save(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	testFileContains(t, pwdfile.Name(), u.Username)
	testListContains(t, m, u)
	testLocalRootExists(t, root, u.Username)

	// update
	u.Password = "securePassword"
	if err := m.Save(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	testFileContains(t, pwdfile.Name(), u.Username)
	testListContains(t, m, u)
	testLocalRootExists(t, root, u.Username)

	// delete
	if err := m.Delete(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	testFileDoesntContain(t, pwdfile.Name(), u.Username)
	testListDoesntContain(t, m, u)
	testLocalRootDoesntExists(t, root, u.Username)
}

func testFileContains(t *testing.T, f, s string) {
	if !fileContains(t, f, s) {
		t.Errorf("file expected to contain %q but it doesn't", s)
	}
}

func testFileDoesntContain(t *testing.T, f, s string) {
	if fileContains(t, f, s) {
		t.Errorf("file expected not to contain %q but it does", s)
	}
}

func fileContains(t *testing.T, f string, s string) bool {
	b, err := ioutil.ReadFile(f)
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
