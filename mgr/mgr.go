package mgr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"syscall"

	"github.com/amenzhinsky/vsftpdmgr/crypt"
	_ "github.com/lib/pq"
)

// User represents a vsftpd virtual user.
type User struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

// Mgr is vstfpd users management entity.
type Mgr struct {
	mu sync.Mutex

	db   *sql.DB
	f    *os.File
	root string
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

	if _, err := os.Lstat(root); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(pwdfile, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	// lock pwdfile exclusively to prevent running
	// multiple Mgr instances on the same file
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, err
	}

	return &Mgr{f: f, root: root, db: db}, nil
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

// Save saves user to the database or update it's password if
// it already exists.
func (m *Mgr) Save(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user.Username == "" || user.Password == "" {
		return errors.New("username or password is blank")
	}

	// encrypt password
	password, err := crypt.MD5(user.Password)
	if err != nil {
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

	_, err := m.db.ExecContext(ctx, `DELETE FROM users WHERE username = $1`, user.Username)
	if err != nil {
		return err
	}
	return m.sync(ctx)
}

// Sync synchronizes users list from database with pwdfile and local root.
func (m *Mgr) Sync(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.sync(ctx)
}

// Close shuts down manager.
func (m *Mgr) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.db.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "mgr error: %v\n", err)
	}

	if err := syscall.Flock(int(m.f.Fd()), syscall.LOCK_UN); err != nil {
		fmt.Fprintf(os.Stderr, "mgr error: %v\n", err)
	}

	if err := m.f.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "mgr error: %v\n", err)
	}

	return nil
}

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

// flush saves users list from database to pwdfile.
func (m *Mgr) sync(ctx context.Context) error {
	users, err := m.list(ctx)
	if err != nil {
		return err
	}
	if err := os.Truncate(m.f.Name(), 0); err != nil {
		return err
	}
	if _, err := m.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// sort users alphabetically
	sort.Slice(users, func(i, j int) bool {
		return users[i].Username < users[j].Username
	})

	for _, u := range users {
		if _, err := m.f.Write([]byte(u.Username + ":" + u.Password + "\n")); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mgr) Clean() error {
	_, err := m.db.Exec(`DELETE FROM users`)
	return err
}
