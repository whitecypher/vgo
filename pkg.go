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

	repo         vcs.Repo `yaml:"-"`
	parent       *Pkg     `yaml:"-"`
	hasManifest  bool     `yaml:"-"`
	manifestFile string   `yaml:"-"`
	installed    bool     `yaml:"-"`

	Name         string  `yaml:"pkg,omitempty"`
	Version      Version `yaml:"ver,omitempty"`
	Reference    string  `yaml:"ref,omitempty"`
	Dependencies []*Pkg  `yaml:"deps,omitempty"`
	URL          string  `yaml:"url,omitempty"`
}

// Load ...
func (p *Pkg) Load(path string) error {
	if p.manifestFile == "" {
		p.manifestFile = "vgo.yaml"
	}
	data, err := ioutil.ReadFile(filepath.Join(path, p.manifestFile))
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
	if p.manifestFile == "" {
		p.manifestFile = "vgo.yaml"
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(path, p.manifestFile), data, os.FileMode(0644))
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

	// find a better way of doing this
	path := name
	if path != "." {
		path = fmt.Sprintf("./vendor/%s", path)
	}
	var deps []string
	deps = append(deps, resolveDeps(path, getDepsFromPackage)...)
	for _, name := range deps {
		name := repoName(name)
		// Skip packages already in manifest
		existing := p.Find(name)
		if existing != nil {
			// check the version for compatibility to try and share packages as much as possible
			continue
		}

		dep := &Pkg{Name: name, parent: p}
		p.Lock()
		p.Dependencies = append(p.Dependencies, dep)
		p.Unlock()
		wg.Add(1)
		dep.ResolveImports(wg, install)
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
	if repo == nil {
		return fmt.Errorf("Could not resolve repo for %s", p.Name)
	}
	Logf("Installing %s", p.Name)
	if !repo.CheckLocal() {
		err = repo.Get()
		if err != nil {
			return err
		}
	}
	p.installed = true
	p.Reference, err = repo.Version()
	p.Load(repo.LocalPath())
	return err
}

// RepoPath path to the package
func (p *Pkg) RepoPath() string {
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
	Logf("Switching %s to %s", p.Name, version)
	err = repo.UpdateVersion(string(version))
	if err != nil {
		return err
	}
	p.Reference, err = repo.Version()
	p.Load(repo.LocalPath())
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
		return fmt.Sprintf("git@github.com:%s.git", strings.Join(parts[1:3], "/"))
	}
	return ""
}

// RepoType attempts to resolve the repository type of the package by it's name
func (p *Pkg) RepoType() vcs.Type {
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

// MarshalYAML implements yaml.Marsheler to prevent duplicate storage of nested packages with vgo.yaml
func (p *Pkg) MarshalYAML() (interface{}, error) {
	copy := *p
	if copy.hasManifest {
		copy.Dependencies = []*Pkg{}
	}
	return copy, nil
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
