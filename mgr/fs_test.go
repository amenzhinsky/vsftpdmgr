package mgr

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMkfs(t *testing.T) {
	t.Parallel()

	root, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	fs := FS{
		Mode: 0755,
		Children: []FS{
			{
				Name: "a/b",
				Mode: 0755,
				Children: []FS{
					{
						Name: "c",
						Mode: 0555,
					},
					{
						Name: "../c",
					},
				},
			},
		},
	}

	if err := mkfs(root, fs, true); err != nil {
		t.Fatal(err)
	}

	testDir(t, root, "", 0755)
	testDir(t, root, "a/b", 0755)
	testDir(t, root, "a/b/c", 0555)
	testDir(t, root, "a/c", 0755)
}

func testDir(t *testing.T, root, name string, mode os.FileMode) {
	path := filepath.Join(root, name)
	stat, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}

	// remove type bytes
	m := stat.Mode() &^ os.ModeType
	if m != mode {
		t.Errorf("%s mode = %s, want %s", path, m, mode)
	}
}
