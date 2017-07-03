package mgr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/amenzhinsky/vsftpdmgr/crypt"
	_ "github.com/lib/pq"
)

// User
type User struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

// Mgr
type Mgr struct {
	mu sync.Mutex

	db   *sql.DB
	conf *os.File
	root string
}

// New
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

	return &Mgr{conf: f, root: root, db: db}, nil
}

// Close shuts down manager.
func (m *Mgr) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.db.Close()

	// unlock file
	if err := syscall.Flock(int(m.conf.Fd()), syscall.LOCK_UN); err != nil {
		fmt.Fprintf(os.Stderr, "mgr error: %v\n", err)
	}
	return m.conf.Close()
}

func (m *Mgr) List(ctx context.Context) (users []*User, err error) {
	rows, err := m.db.QueryContext(ctx, `SELECT username FROM users`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		if err = rows.Scan(&u.Username); err != nil {
			return
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

// Save saves user to the database or update it's password if
// it already exists.
func (m *Mgr) Save(ctx context.Context, user *User) error {
	if user.Username == "" || user.Password == "" {
		return errors.New("username or password is blank")
	}

	// encrypt password
	pass, err := crypt.MD5(user.Password)
	if err != nil {
		return err
	}

	// upsert record on username conflict
	_, err = m.db.ExecContext(ctx, `INSERT INTO users (username, password) VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET password = $2`, user.Username, pass)

	return err
}

func (m *Mgr) Delete(ctx context.Context, user *User) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM users WHERE username = $1`, user.Username)
	return err
}

func (m *Mgr) Sync() error {
	return nil
}

func (m *Mgr) Clean() error {
	_, err := m.db.Exec(`DELETE FROM users`)
	return err
}
