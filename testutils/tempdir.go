package testutils

import (
	"crypto/rand"
	"fmt"
	"os"
	"path"
)

// TempDir contains information about the temp dir
type TempDir struct {
	// path of the test dir
	path string
}

// makeTempDirPath returns a path to a random test directory in $TMPDIR
func makeTempDirPath() string {
	name := make([]byte, 25)
	_, err := rand.Read(name)

	if err != nil {
		println(err.Error())
	}

	tmpdir := os.TempDir()

	return path.Join(tmpdir, fmt.Sprintf("%x", name))
}

// Create a new temporary unique directory in os temp dir
func NewTempDir() TempDir {
	tmp := TempDir{
		path: makeTempDirPath(),
	}

	os.Mkdir(tmp.GetPath(), 0775)

	return tmp
}

// GetPath return the full path to the created temp directory
func (t *TempDir) GetPath() string {
	return t.path
}

// Destroy removes the temp directory and its contents
func (t *TempDir) Destroy() {
	if _, err := os.Stat(t.path); os.IsNotExist(err) {
		return
	}

	err := os.Remove(t.path)

	if err != nil {
		fmt.Println(err.Error())

		return
	}
}
