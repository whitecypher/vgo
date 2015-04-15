package testutils

import (
	"crypto/rand"
	"fmt"
	"os"
	"path"
)

type TempDir struct {
	path string
}

// getTestDirPath returns a path to a random test directory in $TMPDIR
func makeTempDirPath() string {
	name := make([]byte, 25)
	_, err := rand.Read(name)

	if err != nil {
		println(err.Error())
	}

	tmpdir := os.TempDir()

	return path.Join(tmpdir, fmt.Sprintf("%x", name))
}

func NewTempDir() TempDir {
	tmp := TempDir{
		path: makeTempDirPath(),
	}

	os.Mkdir(tmp.GetPath(), 0664)

	return tmp
}

func (t *TempDir) GetPath() string {
	return t.path
}

func (t *TempDir) Destroy() {
	err := os.Remove(t.path)

	if err != nil {
		println(err.Error())

		return
	}
}
