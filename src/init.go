package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func initialize(args []string) {
	// cwd, err := os.Getwd()
	// a := app{}

	// if err != nil {

	// }
	// fmt.Println()
}

func isGappProject(path string) bool {
	gappPath := filepath.Join(path, GAPP_FILE)

	if _, err := os.Stat(gappPath); err != nil {
		return false
	}

	return true
}

func initGappProject(projRootDir string, a app) error {
	js, err := json.Marshal(a)

	if err != nil {
		return err
	}

	gappFilePath := path.Join(projRootDir, GAPP_FILE)

	err2 := ioutil.WriteFile(gappFilePath, js, 0664)

	if err2 != nil {
		return err2
	}

	return nil
}
