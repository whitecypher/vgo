package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Masterminds/vcs"
	"github.com/whitecypher/vgo/lib/native"
	"gopkg.in/yaml.v2"
)

// Version compatibility string e.g. "~1.0.0" or "1.*"
type Version string

// NewPkg creates and initializes a Pkg
func NewPkg(name string) *Pkg {
	return &Pkg{
		Name: name,
	}
}

// Pkg ...
type Pkg struct {
	sync.Mutex `yaml:"-"`

	meta         *build.Package `yaml:"-"`
	repo         vcs.Repo       `yaml:"-"`
	parent       *Pkg           `yaml:"-"`
	hasManifest  bool           `yaml:"-"`
	manifestFile string         `yaml:"-"`
	installed    bool           `yaml:"-"`
	path         string         `yaml:"-"`
	installPath  string         `yaml:"-"`

	Name         string  `yaml:"pkg,omitempty"`
	Version      Version `yaml:"ver,omitempty"`
	Reference    string  `yaml:"ref,omitempty"`
	Dependencies []*Pkg  `yaml:"deps,omitempty"`
	URL          string  `yaml:"url,omitempty"`
}

// Meta get the package meta data using the go/build internal package profiler
func (p *Pkg) Meta() *build.Package {
	if p.meta != nil {
		return p.meta
	}
	m, err := build.Import(p.FQN(), cwd, build.ImportMode(0))
	if err != nil {
		Logf("Unable to import package %s with error %s", p.FQN(), err.Error())
	}
	p.meta = m
	return m
}

// FQN resolves the fully qualified package name. This is the equivalent to the name that go uses dependant on it's context.
func (p Pkg) FQN() string {
	// if p.IsInGoPath() && p.parent != nil {
	// 	return filepath.Join(p.parent.FQN(), "vendor", p.Name)
	// }
	if p.Name == "" {
		return "."
	}
	return p.Name
}

// Root returns the topmost package (typically this is the application package)
func (p *Pkg) Root() *Pkg {
	if p.parent == nil {
		return p
	}
	return p.parent.Root()
}

// IsInGoPath returns whether project and all vendored packages are contained in the $GOPATH
func (p Pkg) IsInGoPath() bool {
	if p.parent != nil {
		return p.parent.IsInGoPath()
	}
	return strings.HasPrefix(p.path, gosrcpath)
}

func resolveImportsRecursive(path string, imports []string) []string {
	r := []string{}
	for _, i := range imports {
		// Skip native packages
		if native.IsNative(i) {
			continue
		}
		// check subpackages for dependencies
		if strings.HasPrefix(i, path) { //  && !strings.HasPrefix(i, filepath.Join(path, "vendor"))
			m, err := build.Import(i, cwd, build.ImportMode(0))
			if err != nil {
				fmt.Println(err.Error())
			} else {
				r = append(r, resolveImportsRecursive(i, m.Imports)...)
			}
		}
		name := repoName(i)
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

// Init ...
func (p *Pkg) Init(meta *build.Package) {
	p.Lock()
	p.path = meta.Dir
	if p.IsInGoPath() {
		p.Name = repoName(meta.ImportPath)
	}
	p.Unlock()

	// i := resolveImportsRecursive(p.Name, meta.Imports)
	// js, _ := json.MarshalIndent(i, "", "  ")
	// fmt.Println(string(js))
	// os.Exit(0)

	wg := sync.WaitGroup{}
	for _, i := range resolveImportsRecursive(p.Name, meta.Imports) {
		name := repoName(i)
		// Skip subpackages
		if strings.HasPrefix(name, p.Name) {
			continue
		}

		// Reuse packages already added to the project
		dep := p.Find(name)
		if dep == nil {
			fmt.Println("adding", name)
			dep = NewPkg(name)
			dep.parent = p
			p.Lock()
			p.Dependencies = append(p.Dependencies, dep)
			p.Unlock()
		} else {
			// check the version compatibility. We might need to create a broken diamond here.
		}
		wg.Add(1)
		go func() {
			dep.Install()
			wg.Done()
		}()
	}
	wg.Wait()
	// p.InstallDeps()
}

// LoadManifest ...
func (p *Pkg) LoadManifest() error {
	p.hasManifest = false
	if len(p.manifestFile) == 0 {
		p.manifestFile = "vgo.yaml"
	}
	fmt.Println("Loading manifest", filepath.Join(p.path, p.manifestFile))
	data, err := ioutil.ReadFile(filepath.Join(p.path, p.manifestFile))
	if err != nil {
		fmt.Println("Unable to load manifest with error", err.Error())
		return err
	}
	p.Lock()
	err = yaml.Unmarshal(data, p)
	p.Unlock()
	if err != nil {
		fmt.Println("Unable to unmarshal manifest with error", err.Error())
		return err
	}
	p.hasManifest = true
	p.updateDepsParents()
	return nil
}

// updateDepsParents resolves the parent (caller) pkg for all dependencies recursively
func (p *Pkg) updateDepsParents() {
	for _, d := range p.Dependencies {
		d.Lock()
		d.parent = p
		d.Unlock()
		d.updateDepsParents()
	}
}

// Find looks for a package in it's dependencies or parents dependencies recursively
func (p Pkg) Find(name string) *Pkg {
	fmt.Println("find", name)
	for _, d := range p.Dependencies {
		fmt.Println("find - check against", d.Name)
		if (*d).Name == name {
			return d
		}
	}
	if p.parent != nil {
		return (*p.parent).Find(name)
	}
	return nil
}

// SaveManifest ...
func (p Pkg) SaveManifest() error {
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(p.path, p.manifestFile), data, os.FileMode(0644))
	if err != nil {
		return err
	}
	return nil
}

// Install the package
func (p *Pkg) Install() error {
	if p.parent == nil {
		// don't touch the current working directory
		return nil
	}
	repo, err := p.VCS()
	if repo == nil {
		fmt.Println(err)
		return fmt.Errorf("Could not resolve repo for %s with error %s", p.Name, err)
	}
	p.Lock()
	p.installed = repo.CheckLocal()
	p.path = repo.LocalPath()
	if !p.installed {
		Logf("Installing %s", p.Name)
		err = repo.Get()
		if err != nil {
			Logf("Failed to install %s with error %s", p.Name, err.Error())
		}
	}
	p.Unlock()
	p.Checkout()
	p.LoadManifest()
	if !p.hasManifest {
		p.Init(p.Meta())
	}
	return err
}

// InstallDeps install package dependencies
func (p *Pkg) InstallDeps() (err error) {
	wg := sync.WaitGroup{}
	for _, dep := range p.Dependencies {
		d := dep
		wg.Add(1)
		go func() {
			err = d.Install()
			if err != nil {
				Logf("Package %s could not be installed with error", err.Error())
			}
			wg.Done()
		}()

	}
	wg.Wait()
	return
}

// RepoPath path to the package
func (p Pkg) RepoPath() string {
	dir := p.installPath
	if len(dir) == 0 {
		dir = installPath
	}
	return path.Join(dir, p.Name)
}

// Checkout switches the package version to the commit nearest maching the Compat string
func (p *Pkg) Checkout() error {
	if p.parent == nil {
		// don't touch the current working directory
		return nil
	}
	repo, err := p.VCS()
	if err != nil {
		return err
	}
	p.Lock()
	version := p.Version
	if p.Reference != "" {
		version = Version(p.Reference)
	}
	p.installed = repo.CheckLocal()
	if p.installed {
		if v, err := repo.Version(); err == nil && p.Reference == v {
			p.Unlock()
			Logf("%s OK %s", p.Reference, p.Name)
			return nil
		}
		err = repo.UpdateVersion(string(version))
		if err != nil {
			p.Unlock()
			return err
		}
	}
	p.Reference, err = repo.Version()
	p.path = repo.LocalPath()
	p.Unlock()
	Logf("Switching %s to %s", p.Name, version)
	p.LoadManifest()
	if !p.hasManifest {
		p.Init(p.Meta())
	}
	return err
}

// VCS resolves the vcs.Repo for the Pkg
func (p *Pkg) VCS() (repo vcs.Repo, err error) {
	p.Lock()
	defer p.Unlock()
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
func (p Pkg) RepoURL() string {
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

// MarshalYAML implements yaml.Marsheler to prevent duplicate storage of nested packages with vgo.yaml
func (p Pkg) MarshalYAML() (interface{}, error) {
	copy := p
	if copy.hasManifest && copy.parent != nil {
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
