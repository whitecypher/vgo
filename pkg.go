package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Masterminds/vcs"
	"gopkg.in/yaml.v2"
)

// Version compatibility string e.g. "~1.0.0" or "1.*"
type Version string

// GoListPackage model for output from go list -json
type GoListPackage struct {
	Dir           string // directory containing package sources
	ImportPath    string // import path of package in dir
	ImportComment string // path in import comment on package statement
	Name          string // package name
	Doc           string // package documentation string
	Target        string // install path
	Shlib         string // the shared library that contains this package (only set when -linkshared)
	Goroot        bool   // is this package in the Go root?
	Standard      bool   // is this package part of the standard Go library?
	Stale         bool   // would 'go install' do anything for this package?
	Root          string // Go root or Go path dir containing this package

	// Source files
	GoFiles        []string // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles       []string // .go sources files that import "C"
	IgnoredGoFiles []string // .go sources ignored due to build constraints
	CFiles         []string // .c source files
	CXXFiles       []string // .cc, .cxx and .cpp source files
	MFiles         []string // .m source files
	HFiles         []string // .h, .hh, .hpp and .hxx source files
	SFiles         []string // .s source files
	SwigFiles      []string // .swig files
	SwigCXXFiles   []string // .swigcxx files
	SysoFiles      []string // .syso object files to add to archive

	// Cgo directives
	CgoCFLAGS    []string // cgo: flags for C compiler
	CgoCPPFLAGS  []string // cgo: flags for C preprocessor
	CgoCXXFLAGS  []string // cgo: flags for C++ compiler
	CgoLDFLAGS   []string // cgo: flags for linker
	CgoPkgConfig []string // cgo: pkg-config names

	// Dependency information
	Imports []string // import paths used by this package
	Deps    []string // all (recursively) imported dependencies

	TestGoFiles  []string // _test.go files in package
	TestImports  []string // imports from TestGoFiles
	XTestGoFiles []string // _test.go files outside package
	XTestImports []string // imports from XTestGoFiles
}

// Pkg ...
type Pkg struct {
	sync.Mutex `yaml:"-"`

	repo         vcs.Repo `yaml:"-"`
	parent       *Pkg     `yaml:"-"`
	hasManifest  bool     `yaml:"-"`
	manifestFile string   `yaml:"-"`
	installed    bool     `yaml:"-"`
	path         string   `yaml:"-"`

	Name         string  `yaml:"pkg,omitempty"`
	Version      Version `yaml:"ver,omitempty"`
	Reference    string  `yaml:"ref,omitempty"`
	Dependencies []*Pkg  `yaml:"deps,omitempty"`
	URL          string  `yaml:"url,omitempty"`
}

// FQN resolves the fully qualified package name. This is the equivalent to the name that go uses dependant on it's context.
func (p Pkg) FQN() string {
	if p.IsInGoPath() && p.parent != nil {
		return filepath.Join(p.parent.FQN(), "vendor", p.Name)
	}
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

// Init attempts to detect package information
func (p *Pkg) Init() {
	output, err := exec.Command("go", "list", "-json", p.FQN()).Output()
	if err != nil {
		return
	}
	var model GoListPackage
	err = json.Unmarshal(output, &model)
	if err != nil {
		return
	}
	p.Lock()
	p.path = model.Dir
	p.manifestFile = "vgo.yaml"
	if p.IsInGoPath() {
		p.Name = repoName(model.ImportPath)
	}
	p.Unlock()
}

// LoadManifest ...
func (p *Pkg) LoadManifest() error {
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

// ResolveImports ...
func (p *Pkg) ResolveImports() error {
	wg := &sync.WaitGroup{}
	for _, name := range resolveDeps(p.FQN(), getDepsFromPackage) {
		name := repoName(name)
		// Skip packages already in manifest
		dep := p.Find(name)
		if dep != nil {
			// check the version for compatibility to try and share packages as much as possible
		} else {
			dep = &Pkg{Name: name, parent: p}
			dep.Init()
			dep.Lock()
			dep.Dependencies = append(dep.Dependencies, dep)
			dep.Unlock()
		}

		wg.Add(1)
		go dep.ResolveImportsAsync(wg)
	}

	wg.Wait()
	return nil
}

// ResolveImportsAsync runs a ResolveImports asynchronously
func (p *Pkg) ResolveImportsAsync(wg *sync.WaitGroup) {
	p.ResolveImports()
	wg.Done()
}

// Install the package
func (p *Pkg) Install() error {
	fmt.Println("install")
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
	p.InstallDeps()
	return err
}

// InstallDeps install package dependencies
func (p *Pkg) InstallDeps() (err error) {
	for _, dep := range p.Dependencies {
		err = dep.Install()
		if err != nil {
			Logf("Package %s could not be installed with error", err.Error())
		}
	}
	return
}

// RepoPath path to the package
func (p Pkg) RepoPath() string {
	return path.Join(installPath, p.Name)
}

// Checkout switches the package version to the commit nearest maching the Compat string
func (p *Pkg) Checkout() error {
	fmt.Println("checkout")
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
	p.ResolveImports()
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
