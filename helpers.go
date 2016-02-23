package main

import (
	"fmt"
	"go/build"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/vcs"
	"github.com/whitecypher/vgo/lib/native"
)

// RepoName extracts the repository name from a package name
func RepoName(packageName string) string {
	// Remove any vendor path prefixes
	vendorParts := strings.Split(packageName, "/vendor/")
	name := vendorParts[len(vendorParts)-1]
	// Limit root package name to 3 levels
	parts := strings.Split(name, "/")
	if len(parts) > 3 {
		parts = parts[0:3]
	}
	return strings.Join(parts, "/")
}

// MustGetwd handles the error returned by Getwd or returns the returns the resulting current working directory path.
func MustGetwd(cwd string, err error) string {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return cwd
}

// repoFromPath attempts to resolve the vcs.Repo from any of the given paths in sequence.
func repoFromPath(paths ...string) vcs.Repo {
	for _, path := range paths {
		repoType, err := vcs.DetectVcsFromFS(path)
		if err != nil {
			continue
		}
		var repo vcs.Repo
		switch repoType {
		case vcs.Git:
			repo, err = vcs.NewGitRepo("", path)
		case vcs.Bzr:
			repo, err = vcs.NewBzrRepo("", path)
		case vcs.Hg:
			repo, err = vcs.NewHgRepo("", path)
		case vcs.Svn:
			repo, err = vcs.NewSvnRepo("", path)
		}
		if err != nil {
			continue
		}
		return repo
	}
	return nil
}

func resolveImportsRecursive(path string, imports []string) []string {
	r := []string{}
	for _, i := range imports {
		// Skip native packages
		if native.IsNative(i) {
			continue
		}
		// Skip vendor packages
		if !vendoring || strings.Contains(i, "vendor") {
			continue
		}
		// check subpackages for dependencies
		m, err := build.Import(i, cwd, build.ImportMode(0))
		if err != nil {
			// Skip this error. It's is likely the package is not installed yet.
		} else {
			r = append(r, resolveImportsRecursive(i, m.Imports)...)
		}
		// add base package to deps list
		name := RepoName(i)
		if name == path {
			continue
		}
		r = append(r, name)
	}
	// return only unique imports
	u := []string{}
	m := map[string]bool{}
	for _, i := range r {
		if _, ok := m[i]; ok {
			continue
		}
		m[i] = true
		u = append(u, i)
	}
	sort.Strings(u)
	return u
}
