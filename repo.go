package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Masterminds/vcs"
	"gopkg.in/yaml.v2"
)

var repos = make(map[string]*Repo)

// Version compatibility string e.g. "~1.0.0" or "1.*"
type Version string

// NewRepo creates and initializes a Repo
func NewRepo(name string, path string) *Repo {
	if r, ok := repos[name]; ok {
		return r
	}
	r := &Repo{
		Name: name,
		path: path,
	}
	repos[name] = r
	return r
}

// Repo ...
type Repo struct {
	sync.RWMutex `yaml:"-"`

	meta         *build.Package `yaml:"-"`
	repo         vcs.Repo       `yaml:"-"`
	parent       *Repo          `yaml:"-"`
	hasManifest  bool           `yaml:"-"`
	manifestFile string         `yaml:"-"`
	installed    bool           `yaml:"-"`
	path         string         `yaml:"-"`

	Name         string  `yaml:"name,omitempty"`
	Version      Version `yaml:"ver,omitempty"`
	Reference    string  `yaml:"ref,omitempty"`
	Dependencies []*Repo `yaml:"deps,omitempty"`
	URL          string  `yaml:"url,omitempty"`
	UsedPkgs     Pkgs    `yaml:"-"`
}

// HasDep checks for the existence of a dependancy
func (r *Repo) HasDep(dep *Repo) bool {
	for _, d := range r.Dependencies {
		if d == dep {
			return true
		}
	}
	return false
}

// AddDep ...
func (r *Repo) AddDep(dep *Repo) {
	if r.HasDep(dep) {
		return
	}
	r.Dependencies = append(r.Dependencies, dep)
}

// FQN resolves the fully qualified package name. This is the equivalent to the name that go uses dependant on it's context.
func (r *Repo) FQN() string {
	if r.IsInGoPath() && !r.IsRoot() {
		return filepath.Join(r.Root().FQN(), "vendor", r.Name)
	}
	if r.Name == "" {
		return "."
	}
	return r.Name
}

// Root returns the topmost package (typically this is the application package)
func (r *Repo) Root() *Repo {
	if r.parent == nil {
		return r
	}
	return r.parent.Root()
}

// IsRoot returns whether the pkg is the root pkg
func (r *Repo) IsRoot() bool {
	return r.parent == nil
}

// IsInGoPath returns whether project and all vendored packages are contained in the $GOPATH
func (r *Repo) IsInGoPath() bool {
	if r.parent != nil {
		return r.parent.IsInGoPath()
	}
	return strings.HasPrefix(r.path, gosrcpath)
}

// LoadManifest ...
func (r *Repo) LoadManifest() error {
	r.hasManifest = false
	if len(r.manifestFile) == 0 {
		r.manifestFile = "vgo.yaml"
	}
	data, err := ioutil.ReadFile(filepath.Join(r.path, r.manifestFile))
	if err != nil {
		return err
	}
	r.Lock()
	err = yaml.Unmarshal(data, r)
	r.Unlock()
	if err != nil {
		return err
	}
	r.hasManifest = true
	r.updateDepsParents()
	return nil
}

// updateDepsParents resolves the parent (caller) pkg for all dependencies recursively
func (r *Repo) updateDepsParents() {
	for _, d := range r.Dependencies {
		d.Lock()
		d.parent = r
		d.Unlock()
		d.updateDepsParents()
	}
}

// Find looks for a package in it's dependencies or parents dependencies recursively
func (r *Repo) Find(name string) *Repo {
	for _, d := range r.Dependencies {
		if (*d).Name == name {
			return d
		}
	}
	if r.parent != nil {
		return (*r.parent).Find(name)
	}
	return nil
}

// SaveManifest ...
func (r *Repo) SaveManifest() error {
	data, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(r.path, r.manifestFile), data, os.FileMode(0644))
	if err != nil {
		return err
	}
	return nil
}

// Install the package
func (r *Repo) Install() error {
	if r.parent == nil {
		// don't touch the current working directory
		return nil
	}
	repo, err := r.VCS()
	if repo == nil {
		return fmt.Errorf("Could not resolve repo for %s with error %s", r.Name, err)
	}
	r.Lock()
	r.installed = repo.CheckLocal()
	r.path = repo.LocalPath()
	if !r.installed {
		Logf("Installing %s", r.Name)
		err = repo.Get()
		if err != nil {
			Logf("Failed to install %s with error %s, %s", r.Name, err.Error(), r.path)
		}
	}
	r.Unlock()
	r.Checkout()
	return err
}

// InstallDeps install package dependencies
func (r *Repo) InstallDeps() (err error) {
	wg := sync.WaitGroup{}
	for _, dep := range r.Dependencies {
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

// RelPath returns the path to package relative to the root package
func (r *Repo) RelPath() string {
	return strings.TrimPrefix(cwd, r.path)
}

// Checkout switches the package version to the commit nearest maching the Compat string
func (r *Repo) Checkout() error {
	if r.parent == nil {
		// don't touch the current working directory
		return nil
	}
	repo, err := r.VCS()
	if err != nil {
		return err
	}
	if repo.IsDirty() {
		Logf("Skipping checkout for %s. Dependency is dirty.", r.Name)
	}
	r.Lock()
	version := r.Version
	if r.Reference != "" {
		version = Version(r.Reference)
	}
	r.installed = repo.CheckLocal()
	if r.installed {
		v := string(version)
		if repo.IsReference(v) {
			Logf("OK %s", r.Name)
			r.Unlock()
			return nil
		}
		err = repo.UpdateVersion(v)
		if err != nil {
			r.Unlock()
			Logf("Checkout failed with error %s", err.Error())
			return err
		}
	}
	r.Reference, err = repo.Version()
	r.path = repo.LocalPath()
	r.Unlock()
	r.LoadManifest()
	if !r.hasManifest {
		r.UsedPkgs.Init()
		r.UsedPkgs.MapDeps(PackageRepoMapper)
	}
	r.InstallDeps()
	return err
}

// VCS resolves the vcs.Repo for the Repo
func (r *Repo) VCS() (repo vcs.Repo, err error) {
	r.Lock()
	defer r.Unlock()
	if r.repo != nil {
		repo = r.repo
		return
	}
	repoType := r.RepoType()
	repoURL := r.RepoURL()
	repoPath := r.path
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
	r.repo = repo
	return
}

// RepoURL creates the repo url from the package import path
func (r *Repo) RepoURL() string {
	if r.URL != "" {
		return r.URL
	}
	// If it's already installed in vendor or gopath, grab the url from there
	repo := repoFromPath(r.path, filepath.Join(gopath, "src", r.Name))
	if repo != nil {
		return repo.Remote()
	}
	// Fallback to resolving the path from the package import path
	// Add more cases as needed/requested
	parts := strings.Split(r.Name, "/")
	switch parts[0] {
	case "github.com":
		return fmt.Sprintf("git@github.com:%s.git", strings.Join(parts[1:3], "/"))
	case "golang.org":
		return fmt.Sprintf("git@github.com:golang/%s.git", parts[2])
	case "gopkg.in":
		nameParts := strings.Split(parts[2], ".")
		name := strings.Join(nameParts[:len(nameParts)-1], ".")
		r.Version = Version(nameParts[len(nameParts)-1])
		return fmt.Sprintf("git@github.com:%s/%s.git", parts[1], name)
	}
	return ""
}

// RepoType attempts to resolve the repository type of the package by it's name
func (r *Repo) RepoType() vcs.Type {
	// If it's already installed in vendor or gopath, grab the type from there
	repo := repoFromPath(r.path, filepath.Join(r.path, "vendor", r.Name), filepath.Join(gopath, "src", r.Name))
	if repo != nil {
		return repo.Vcs()
	}
	// Fallback to resolving the type from the package import path
	// Add more cases as needed/requested
	parts := strings.Split(r.Name, "/")
	switch parts[0] {
	case "github.com":
		return vcs.Git
	case "golang.org":
		return vcs.Git
	case "gopkg.in":
		return vcs.Git
	}
	return vcs.NoVCS
}

// MarshalYAML implements yaml.Marsheler to prevent duplicate storage of nested packages with vgo.yaml
func (r *Repo) MarshalYAML() (interface{}, error) {
	r.RLock()
	copy := *r
	r.RUnlock()
	if copy.Name == "." {
		copy.Name = ""
	}
	if copy.hasManifest && copy.parent != nil {
		copy.Dependencies = []*Repo{}
	}
	return copy, nil
}
