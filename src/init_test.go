package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whitecypher/gapp/testutils"
)

var (
	dir testutils.TempDir
)

func TestMain(m *testing.M) {
	// Set up
	dir = testutils.NewTempDir()

	// Run tests
	result := m.Run()

	// Tear down
	dir.Destroy()

	// Exit
	os.Exit(result)
}

func TestIsGappProject(t *testing.T) {
	assert := assert.New(t)

	t.Log(dir.GetPath())

	assert.False(IsGappProject(dir.GetPath()), "Expected a false value")
}

func TestInitGappProject(t *testing.T) {

}
