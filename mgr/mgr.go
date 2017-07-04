package mgr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/amenzhinsky/vsftpdmgr/crypt"
	_ "github.com/lib/pq"
)

// User represents a vsftpd virtual user.
type User struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

// Mgr is vsftpd users management entity.
type Mgr struct {
	mu      sync.Mutex
	db      *sql.DB
	pwdfile string
	root    string
}

// New creates new Mgr.
func New(root, pwdfile, databaseURL string) (*Mgr, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		username VARCHAR(32) NOT NULL PRIMARY KEY,
		password VARCHAR(34) NOT NULL
	)`); err != nil {
		return nil, err
	}

	// create local root if it doesn't exist
	if _, err := os.Lstat(root); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err = os.MkdirAll(root, 0755); err != nil {
			return nil, err
		}
	}

	// try to open pwdfile and create it if it doesn't exist
	f, err := os.OpenFile(pwdfile, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return &Mgr{pwdfile: pwdfile, root: root, db: db}, nil
}

// List returns list of all users.
func (m *Mgr) List(ctx context.Context) ([]*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	users, err := m.list(ctx)
	if err != nil {
		return nil, err
	}

	// hide passwords
	for _, u := range users {
		u.Password = ""
	}
	return users, nil
}

// ErrInvalidUser is returned when user cannot be saved.
var ErrInvalidUser = errors.New("user is not valid")

// Save saves user to the database or update it's password if
// it already exists.
func (m *Mgr) Save(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(user.Username) < 4 || len(user.Password) < 4 {
		return ErrInvalidUser
	}

	// encrypt password
	password, err := crypt.MD5(user.Password)
	if err != nil {
		return err
	}

	// create user's local root
	err = os.Mkdir(filepath.Join(m.root, user.Username), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// upsert record on username conflict
	_, err = m.db.ExecContext(ctx, `INSERT INTO users (username, password) VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET password = $2`, user.Username, password)
	if err != nil {
		return err
	}
	return m.sync(ctx)
}

// Delete deletes a virtual user.
func (m *Mgr) Delete(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.RemoveAll(filepath.Join(m.root, user.Username)); err != nil {
		return err
	}

	_, err := m.db.ExecContext(ctx, `DELETE FROM users WHERE username = $1`, user.Username)
	if err != nil {
		return err
	}
	return m.sync(ctx)
}

// Close shuts down manager.
func (m *Mgr) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.db.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "mgr error: %v\n", err)
	}

	return nil
}

// list retrieves list of users from the database.
func (m *Mgr) list(ctx context.Context) (users []*User, err error) {
	rows, err := m.db.QueryContext(ctx, `SELECT username, password FROM users`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		if err = rows.Scan(&u.Username, &u.Password); err != nil {
			return
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

// header is written to pwdfile every sync.
var header = []byte("# This file is managed by vsftpdmgr, all changes will be overwritten\n\n")

// sync saves users list from database to the pwdfile.
func (m *Mgr) sync(ctx context.Context) (err error) {
	users, err := m.list(ctx)
	if err != nil {
		return
	}

	t, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer func() {
		t.Close()
		if err != nil {
			os.Remove(t.Name())
		}
	}()

	// sort users alphabetically
	sort.Slice(users, func(i, j int) bool {
		return users[i].Username < users[j].Username
	})

	// write header and content
	if _, err = t.Write(header); err != nil {
		return
	}
	for _, u := range users {
		if _, err = t.Write([]byte(u.Username + ":" + u.Password + "\n")); err != nil {
			return
		}
	}

	// safely replace existing pwdfile file
	if err = os.Remove(m.pwdfile); err != nil {
		return
	}
	if err = os.Rename(t.Name(), m.pwdfile); err != nil {
		return
	}
	return
}

// Clean delete all records from the users table.
func (m *Mgr) Clean() error {
	_, err := m.db.Exec(`DELETE FROM users`)
	return err
}
