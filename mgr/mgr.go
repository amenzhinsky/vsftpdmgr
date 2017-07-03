package mgr

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"syscall"
)

// Mgr
type Mgr struct {
	mu sync.Mutex

	conf *os.File
	root string
}

// New
func New(root, pwdfile string, db *sql.DB) (*Mgr, error) {
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

	return &Mgr{conf: f, root: root}, nil
}

// Close closes pwdfile.
func (m *Mgr) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// unlock file
	if err := syscall.Flock(int(m.conf.Fd()), syscall.LOCK_UN); err != nil {
		fmt.Fprintf(os.Stderr, "mgr error: %v\n", err)
	}
	return m.conf.Close()
}

func (m *Mgr) List() ([]string, error) {
	return nil, nil
}

func (m *Mgr) Update(name, password string) error {
	return nil
}

func (m *Mgr) Delete(name string) error {
	return nil
}
