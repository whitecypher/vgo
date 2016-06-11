package native

import (
	"fmt"
	"os"
	"runtime"
	"path"
	"path/filepath"
	"io/ioutil"
)

var (
	pkgMap = map[string]bool{}
	pkgs = []string{}
)

func init() {
	root := runtime.GOROOT()
	if len(root) == 0 {
		fmt.Print("Go not installed or missing GOROOT environment value")
		os.Exit(1)
	}
	src := path.Join(root, "src")
	dirs := listDirsRecursive(src)
	for _, d := range dirs {
		rel, err := filepath.Rel(src, d)
		if err != nil {
			fmt.Print("Unable to resolve native packages with error:", err.Error())
			os.Exit(1)
		}
		pkgs = append(pkgs, rel)
	}
	for _, name := range pkgs {
		pkgMap[name] = true
	}
}

func listDirsRecursive(dir string) (r []string) {
	dirs, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("Unable to read dir %s with error: %s", dir, err.Error())
		return
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		dir := path.Join(dir, d.Name())
		r = append(r, dir)
		r = append(r, listDirsRecursive(dir)...)
	}
	return
}

// Packages returns an array of native packages
func Packages() []string {
	return pkgs
}

// IsNative returns whether given package name is a native package or not
func IsNative(name string) bool {
	_, ok := pkgMap[name]
	return ok
}
