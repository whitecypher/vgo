package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func IsGappProject(path string) bool {
	gappPath := filepath.Join(path, "gapp.json")

	if _, err := os.Stat(gappPath); err != nil {
		return false
	}

	return true
}

func InitGappProject(projRootDir string, app Application) error {
	js, err := json.Marshal(app)

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
