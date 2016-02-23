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

// Version compatibility string e.g. "~1.0.0" or "1.*"
type Version string

// NewRepo creates and initializes a Repo
func NewRepo(name string, path string) *Repo {
	return &Repo{
		Name: name,
		path: path,
	}
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
	pkg          *Pkg           `yaml:"-"`

	Name         string  `yaml:"name,omitempty"`
	Version      Version `yaml:"ver,omitempty"`
	Reference    string  `yaml:"ref,omitempty"`
	Dependencies []*Repo `yaml:"deps,omitempty"`
	URL          string  `yaml:"url,omitempty"`
}

// FQN resolves the fully qualified package name. This is the equivalent to the name that go uses dependant on it's context.
func (p *Repo) FQN() string {
	if p.IsInGoPath() && !p.IsRoot() {
		return filepath.Join(p.Root().FQN(), "vendor", p.Name)
	}
	if p.Name == "" {
		return "."
	}
	return p.Name
}

// Root returns the topmost package (typically this is the application package)
func (p *Repo) Root() *Repo {
	if p.parent == nil {
		return p
	}
	return p.parent.Root()
}

// IsRoot returns whether the pkg is the root pkg
func (p *Repo) IsRoot() bool {
	return p.parent == nil
}

// IsInGoPath returns whether project and all vendored packages are contained in the $GOPATH
func (p *Repo) IsInGoPath() bool {
	if p.parent != nil {
		return p.parent.IsInGoPath()
	}
	return strings.HasPrefix(p.path, gosrcpath)
}

// Init ...
func (p *Repo) Init() {
	p.pkg = NewPkg(p.FQN(), p.path)
	p.pkg.Print(os.Stdout)
	// p.Lock()
	// p.path = meta.Dir
	// if p.IsInGoPath() {
	// 	p.Name = repoName(meta.ImportPath)
	// }
	// p.Unlock()
	//
	// wg := sync.WaitGroup{}
	// for _, i := range resolveImportsRecursive(p.Name, meta.Imports) {
	// 	name := repoName(i)
	// 	// Skip subpackages
	// 	if strings.HasPrefix(name, p.Name) {
	// 		continue
	// 	}
	//
	// 	// Reuse packages already added to the project
	// 	dep := p.Find(name)
	// 	if dep == nil {
	// 		dep = NewRepo(name)
	// 		dep.parent = p
	// 		p.Lock()
	// 		p.Dependencies = append(p.Dependencies, dep)
	// 		p.Unlock()
	//
	// 		wg.Add(1)
	// 		go func() {
	// 			dep.Install()
	// 			wg.Done()
	// 		}()
	// 	} else {
	// 		// check the version compatibility. We might need to create a broken diamond here.
	// 	}
	// }
	// wg.Wait()
	// p.InstallDeps()
}

// LoadManifest ...
func (p *Repo) LoadManifest() error {
	p.hasManifest = false
	if len(p.manifestFile) == 0 {
		p.manifestFile = "vgo.yaml"
	}
	data, err := ioutil.ReadFile(filepath.Join(p.path, p.manifestFile))
	if err != nil {
		return err
	}
	p.Lock()
	err = yaml.Unmarshal(data, p)
	p.Unlock()
	if err != nil {
		return err
	}
	p.hasManifest = true
	p.updateDepsParents()
	return nil
}

// updateDepsParents resolves the parent (caller) pkg for all dependencies recursively
func (p *Repo) updateDepsParents() {
	for _, d := range p.Dependencies {
		d.Lock()
		d.parent = p
		d.Unlock()
		d.updateDepsParents()
	}
}

// Find looks for a package in it's dependencies or parents dependencies recursively
func (p *Repo) Find(name string) *Repo {
	for _, d := range p.Dependencies {
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
func (p *Repo) SaveManifest() error {
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
func (p *Repo) Install() error {
	if p.parent == nil {
		// don't touch the current working directory
		return nil
	}
	repo, err := p.VCS()
	if repo == nil {
		return fmt.Errorf("Could not resolve repo for %s with error %s", p.Name, err)
	}
	p.Lock()
	p.installed = repo.CheckLocal()
	p.path = repo.LocalPath()
	if !p.installed {
		Logf("Installing %s", p.Name)
		err = repo.Get()
		if err != nil {
			Logf("Failed to install %s with error %s, %s", p.Name, err.Error(), p.path)
		}
	}
	p.Unlock()
	p.Checkout()
	return err
}

// InstallDeps install package dependencies
func (p *Repo) InstallDeps() (err error) {
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

// RelativePath returns the path to package relative to the root package
func (p *Repo) RelPath() string {
	return strings.TrimPrefix(cwd, p.path)
}

// Checkout switches the package version to the commit nearest maching the Compat string
func (p *Repo) Checkout() error {
	if p.parent == nil {
		// don't touch the current working directory
		return nil
	}
	repo, err := p.VCS()
	if err != nil {
		return err
	}
	if repo.IsDirty() {
		Logf("Skipping checkout for %s. Dependency is dirty.", p.Name)
	}
	p.Lock()
	version := p.Version
	if p.Reference != "" {
		version = Version(p.Reference)
	}
	p.installed = repo.CheckLocal()
	if p.installed {
		v := string(version)
		if repo.IsReference(v) {
			Logf("OK %s", p.Name)
			p.Unlock()
			return nil
		}
		err = repo.UpdateVersion(v)
		if err != nil {
			p.Unlock()
			Logf("Checkout failed with error %s", err.Error())
			return err
		}
	}
	p.Reference, err = repo.Version()
	p.path = repo.LocalPath()
	p.Unlock()
	p.LoadManifest()
	if !p.hasManifest {
		p.parent.Init()
	}
	return err
}

// VCS resolves the vcs.Repo for the Repo
func (p *Repo) VCS() (repo vcs.Repo, err error) {
	p.Lock()
	defer p.Unlock()
	if p.repo != nil {
		repo = p.repo
		return
	}
	repoType := p.RepoType()
	repoURL := p.RepoURL()
	repoPath := p.path
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
func (p *Repo) RepoURL() string {
	if p.URL != "" {
		return p.URL
	}
	// If it's already installed in vendor or gopath, grab the url from there
	repo := repoFromPath(p.path, filepath.Join(gopath, "src", p.Name))
	if repo != nil {
		return repo.Remote()
	}
	// Fallback to resolving the path from the package import path
	// Add more cases as needed/requested
	parts := strings.Split(p.Name, "/")
	switch parts[0] {
	case "github.com":
		return fmt.Sprintf("git@github.com:%s.git", strings.Join(parts[1:3], "/"))
	case "golang.org":
		return fmt.Sprintf("git@github.com:golang/%s.git", parts[2])
	case "gopkg.in":
		nameParts := strings.Split(parts[2], ".")
		name := strings.Join(nameParts[:len(nameParts)-1], ".")
		p.Version = Version(nameParts[len(nameParts)-1])
		return fmt.Sprintf("git@github.com:%s/%s.git", parts[1], name)
	}
	return ""
}

// RepoType attempts to resolve the repository type of the package by it's name
func (p *Repo) RepoType() vcs.Type {
	// If it's already installed in vendor or gopath, grab the type from there
	repo := repoFromPath(p.path, filepath.Join(p.path, "vendor", p.Name), filepath.Join(gopath, "src", p.Name))
	if repo != nil {
		return repo.Vcs()
	}
	// Fallback to resolving the type from the package import path
	// Add more cases as needed/requested
	parts := strings.Split(p.Name, "/")
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
func (p *Repo) MarshalYAML() (interface{}, error) {
	p.RLock()
	copy := *p
	p.RUnlock()
	if copy.hasManifest && copy.parent != nil {
		copy.Dependencies = []*Repo{}
	}
	return copy, nil
}
