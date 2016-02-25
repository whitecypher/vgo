package main

import (
	"fmt"
	"os"

	"github.com/Masterminds/vcs"
)

// MustGetwd handles the error returned by Getwd or returns the returns the resulting current working directory path.
func MustGetwd(cwd string, err error) string {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return cwd
}

// PackageRepoMapper maps packages to repositories
func PackageRepoMapper(p *Pkg, d *Pkg) {
	pr := NewRepo(p.RepoName(), p.Dir)
	pr.UsedPkgs = append(pr.UsedPkgs, d)
	if p.RepoName() == d.RepoName() {
		return
	}
	dr := NewRepo(d.RepoName(), d.Dir)
	pr.AddDep(dr)
	dr.parent = pr
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
