package main

import (
	"fmt"
	"go/build"
	"strings"

	"github.com/whitecypher/vgo/lib/native"
)

var (
	pkgmap = make(map[string]*Pkg)
	depth  = 0
)

// PkgNotFoundError ...
type PkgNotFoundError string

func (e PkgNotFoundError) Error() string {
	return string(e)
}

// NewPkg ...
func NewPkg(name, dir string, parent *Pkg) *Pkg {
	// , repo *Repo
	p := &Pkg{
		parent: parent,
		Name:   name,
		Dir:    dir,
		// Repo:   repo,
	}
	p.Init()
	return p
}

// Pkg ...
type Pkg struct {
	parent *Pkg

	Name string
	Dir  string
	Repo *Repo
}

// Meta ...
func (p *Pkg) Meta() (bp *build.Package, err error) {
	bp, err = build.Import(p.Name, p.Dir, build.ImportMode(0))
	if bp == nil {
		err = PkgNotFoundError(fmt.Sprintf("Unable to find package %s", p.Name))
		return
	}
	if err != nil {
		return
	}
	return
}

// Init gets the package meta data using the go/build internal package profiler
func (p *Pkg) Init() {
	fmt.Println(strings.Repeat("  ", depth), p.Name)
	var rp *Repo
	if p.parent != nil {
		rp = p.parent.Repo
	}
	p.Repo = NewRepo(p.RepoName(), rp)
	m, _ := p.Meta()

	depth++
	for _, i := range m.Imports {
		if native.IsNative(i) {
			continue
		}
		// fmt.Println(strings.Repeat("  ", depth), i)
		dep := NewPkg(i, installPath, p)
		if dep.RepoName() == p.RepoName() {
			continue
		}
		p.Repo.AddDep(dep.Repo)
	}
	depth--
}

// ImportName ...
func (p *Pkg) ImportName() string {
	// Remove any vendor path prefixes
	vendorParts := strings.Split(p.Name, "/vendor/")
	return vendorParts[len(vendorParts)-1]
}

// SubPath ...
func (p *Pkg) SubPath() string {
	return strings.TrimPrefix(p.ImportName(), p.RepoName())
}

// RepoName ...
func (p *Pkg) RepoName() string {
	// Limit root package name to 3 levels
	parts := strings.Split(p.ImportName(), "/")
	if len(parts) > 3 {
		parts = parts[0:3]
	}
	return strings.Join(parts, "/")
}

// RepoPath ...
func (p *Pkg) RepoPath() string {
	return strings.TrimSuffix(p.Dir, p.SubPath())
}
