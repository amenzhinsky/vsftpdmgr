package mgr

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// FS is file system tree.
type FS struct {
	Name     string `json:"name"`
	Mode     uint   `json:"mode"`
	Owner    string `json:"owner"`
	Group    string `json:"group"`
	Children []FS   `json:"children"`
}

// mkfs creates a real file system representation of fs inside of root,
// hence fs.Name is replaced with the root value.
func mkfs(root string, fs FS, first bool) error {
	if first {
		if fs.Name != "" {
			return errors.New("name must be blank for root node")
		}

		// TODO: avoid modifying fs in case we use pointer here
		fs.Name = root
	}

	// check that directory is to create within the root.
	fs.Name = filepath.Clean(fs.Name)
	if !strings.HasPrefix(fs.Name, root) {
		return fmt.Errorf("%s is outside of %q root", fs.Name, root)
	}

	mode := os.FileMode(0755)
	if fs.Mode != 0 {
		mode = os.FileMode(fs.Mode)
	}

	if err := os.MkdirAll(fs.Name, mode); err != nil && !os.IsExist(err) {
		return err
	}

	// we can skip chmod here when we know that MkdirAll succeeds.
	if fs.Mode != 0 {
		if err := os.Chmod(fs.Name, os.FileMode(fs.Mode)); err != nil {
			return err
		}
	}

	c, err := user.Current()
	if err != nil {
		return err
	}
	uid, gid := c.Uid, c.Gid

	if fs.Owner != "" {
		u, err := user.Lookup(fs.Owner)
		if err != nil {
			return err
		}
		uid = u.Uid
	}

	if fs.Group != "" {
		g, err := user.LookupGroup(fs.Group)
		if err != nil {
			return err
		}
		gid = g.Gid
	}

	iuid, err := strconv.Atoi(uid)
	if err != nil {
		return err
	}

	igid, err := strconv.Atoi(gid)
	if err != nil {
		return err
	}

	// TODO: posix uid and gid are ints, won't work on windows
	// chown only when uid or gid are different from the current user.
	if uid != c.Uid || gid != c.Gid {
		if err = os.Lchown(fs.Name, iuid, igid); err != nil {
			return err
		}
	}

	// recursively create children
	for _, ch := range fs.Children {
		if ch.Name == "" {
			return errors.New("node name is blank")
		}

		// TODO: avoid modifying ch in case we use pointer here
		ch.Name = filepath.Join(fs.Name, ch.Name)
		if err = mkfs(root, ch, false); err != nil {
			return err
		}
	}
	return nil
}
