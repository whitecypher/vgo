package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Masterminds/vcs"
	"gopkg.in/yaml.v2"
)

// Version compatibility string e.g. "~1.0.0" or "1.*"
type Version string

// List structure from go list -json .
type List struct {
	Dir        string   `json:"Dir"`
	Importpath string   `json:"ImportPath"`
	Name       string   `json:"Name"`
	Target     string   `json:"Target"`
	Root       string   `json:"Root"`
	Gofiles    []string `json:"GoFiles"`
	Imports    []string `json:"Imports"`
	Deps       []string `json:"Deps"`
}

// Pkg ...
type Pkg struct {
	sync.Mutex `yaml:"-"`

	repo   vcs.Repo `yaml:"-"`
	parent *Pkg     `yaml:"-"`

	Name         string  `yaml:"pkg,omitempty"`
	Version      Version `yaml:"ver,omitempty"`
	Reference    string  `yaml:"ref,omitempty"`
	Dependencies []*Pkg  `yaml:"deps,omitempty"`
	URL          string  `yaml:"url,omitempty"`
}

// Load ...
func (p *Pkg) Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, p)
	if err != nil {
		return err
	}
	p.updateDepsParents()
	return nil
}

// updateDepsParents resolves the parent (caller) pkg for all dependencies recursively
func (p *Pkg) updateDepsParents() {
	for _, d := range p.Dependencies {
		d.parent = p
		d.updateDepsParents()
	}
}

// Find looks for a package in it's dependencies or parents dependencies recursively
func (p *Pkg) Find(name string) *Pkg {
	for _, d := range p.Dependencies {
		if d.Name == name {
			return d
		}
	}
	if p.parent != nil {
		return p.parent.Find(name)
	}
	return nil
}

// Save ...
func (p *Pkg) Save(path string) error {
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, data, os.FileMode(0644))
	if err != nil {
		return err
	}
	return nil
}

// ResolveImports ...
func (p *Pkg) ResolveImports(wg *sync.WaitGroup, install bool) error {
	defer wg.Done()

	name := p.Name
	if len(name) == 0 || strings.HasSuffix(cwd, p.Name) {
		name = "."
	}

	if install {
		err := p.Install()
		if err != nil {
			fmt.Println(err)
		} else {
			err := p.Checkout()
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	var deps []string
	deps = append(deps, resolveDeps(name, getDepsFromPackage)...)
	for _, name := range deps {
		// Skip packages already in manifest
		name := repoName(name)
		packageLock.RLock()
		// replace packages list in favor of p.Parent with search method
		sp, ok := packages[name]
		packageLock.RUnlock()
		if ok {
			wg.Add(1)
			go sp.ResolveImports(wg, install)
			continue
		}

		dep := &Pkg{Name: name}
		addToPackagesMap(dep)
		wg.Add(1)
		go dep.ResolveImports(wg, install)
		p.Lock()
		p.Dependencies = append(p.Dependencies, dep)
		p.Unlock()
	}
	return nil
}

// Install the package
func (p *Pkg) Install() error {
	// don't touch the current working directory
	if p.Name == "." || strings.HasSuffix(cwd, p.Name) {
		return nil
	}
	p.Lock()
	defer p.Unlock()
	repo, err := p.VCS()
	if err != nil {
		return err
	}
	if !repo.CheckLocal() {
		err = repo.Get()
		if err != nil {
			return err
		}
	}
	p.Reference, err = repo.Version()
	return err
}

// RepoPath path to the package
func (p Pkg) RepoPath() string {
	return path.Join(installPath, p.Name)
}

// Checkout switches the package version to the commit nearest maching the Compat string
func (p *Pkg) Checkout() error {
	// don't touch the current working directory
	if p.Name == "." || strings.HasSuffix(cwd, p.Name) {
		return nil
	}
	p.Lock()
	defer p.Unlock()
	repo, err := p.VCS()
	if err != nil {
		return err
	}
	version := p.Version
	if p.Reference != "" {
		version = Version(p.Reference)
	}
	err = repo.UpdateVersion(string(version))
	if err != nil {
		return err
	}
	p.Reference, err = repo.Version()
	return err
}

// VCS resolves the vcs.Repo for the Pkg
func (p *Pkg) VCS() (repo vcs.Repo, err error) {
	if p.repo != nil {
		repo = p.repo
		return
	}
	repoType := p.RepoType()
	repoURL := p.RepoURL()
	repoPath := p.RepoPath()
	switch repoType {
	case vcs.Git:
		repo, err = vcs.NewGitRepo(repoURL, repoPath)
	case vcs.Bzr:
		repo, err = vcs.NewBzrRepo(repoURL, repoPath)
	case vcs.Hg:
		repo, err = vcs.NewHgRepo(repoURL, repoPath)
	case vcs.Svn:
		repo, err = vcs.NewSvnRepo(repoURL, repoPath)
	}
	p.repo = repo
	return
}

// RepoURL creates the repo url from the package import path
func (p *Pkg) RepoURL() string {
	if p.URL != "" {
		return p.URL
	}
	// If it's already installed in vendor or gopath, grab the url from there
	repo := repoFromPath(p.RepoPath(), filepath.Join(gopath, "src", p.Name))
	if repo != nil {
		return repo.Remote()
	}
	// Fallback to resolving the path from the package import path
	// Add more cases as needed/requested
	parts := strings.Split(p.Name, "/")
	switch parts[0] {
	case "github.com":
		return fmt.Sprintf("git@github.com:%s.git", strings.Join(parts[1:2], "/"))
	}
	return ""
}

// RepoType attempts to resolve the repository type of the package by it's name
func (p Pkg) RepoType() vcs.Type {
	// If it's already installed in vendor or gopath, grab the type from there
	repo := repoFromPath(p.RepoPath(), filepath.Join(gopath, "src", p.Name))
	if repo != nil {
		return repo.Vcs()
	}
	// Fallback to resolving the type from the package import path
	// Add more cases as needed/requested
	parts := strings.Split(p.Name, "/")
	switch parts[0] {
	case "github.com":
		return vcs.Git
	}
	return vcs.NoVCS
}

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
