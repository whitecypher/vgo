package testutils

import (
	"os"
	"stretchr/testify/assert"
	"testing"
)

var (
	dir TempDir
)

func TestMain(m *testing.M) {
	// Set up
	dir = NewTempDir()

	// Run tests
	result := m.Run()

	// Tear down
	dir.Destroy()

	// Exit
	os.Exit(result)
}

func TestNewTempDir(t *testing.T) {
	assert := assert.New(t)

	t.Log(dir.GetPath())

	assert.IsType(TempDir{}, dir, "Expected value to be of type TempDir")

	if _, err := os.Stat(dir.GetPath()); os.IsNotExist(err) {
		assert.Fail("Expected temp directory to exist", dir.GetPath())
	}
}

func TestGetPath(t *testing.T) {
	assert.Equal(t, dir.GetPath(), dir.GetPath(), "Expected multiple calls to TempDir.GetPath to be equal")
}

func TestDestroy(t *testing.T) {
	assert := assert.New(t)

	dir.Destroy()

	if _, err := os.Stat(dir.GetPath()); !os.IsNotExist(err) {
		assert.Fail("Expected temp directory to be removed", dir.GetPath())
	}
}
