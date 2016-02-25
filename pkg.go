package main

import (
	"go/build"
	"io"
	"strings"

	"github.com/whitecypher/vgo/lib/native"
)

var pkgmap = make(map[string]*Pkg)

// NewPkg ...
func NewPkg(name, dir string) *Pkg {
	if p, ok := pkgmap[name]; ok {
		return p
	}
	p := &Pkg{
		Name: name,
		Dir:  dir,
	}
	p.Init()
	return p
}

// Pkg ...
type Pkg struct {
	Name string
	Dir  string
	Deps Pkgs
}

// Init gets the package meta data using the go/build internal package profiler
func (p *Pkg) Init() {
	m, err := build.Import(p.Name, p.Dir, build.ImportMode(0))
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok {
			Logf("Unable to import package %s with error %s", p.Name, err.Error())
		}
	}
	p.Dir = m.Dir
	pkgmap[p.Name] = p
	for _, i := range m.Imports {
		if native.IsNative(i) {
			continue
		}
		p.Deps = append(p.Deps, NewPkg(i, p.Dir))
	}
}

// MapDeps ...
func (p *Pkg) MapDeps(mapper func(parent *Pkg, dependency *Pkg)) {
	for _, d := range p.Deps {
		mapper(p, d)
		d.MapDeps(mapper)
	}
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

// Print ...
func (p *Pkg) Print(w io.Writer, indent string) {
	p.print(w, indent, 0)
}

func (p *Pkg) print(w io.Writer, indent string, depth int) {
	w.Write([]byte(strings.Repeat(indent, depth) + p.Name + "\n"))
	for _, d := range p.Deps {
		d.print(w, indent, depth+1)
	}
}

// Pkgs list
type Pkgs []*Pkg

// Init all pkgs in list
func (ps Pkgs) Init() {
	for _, p := range ps {
		p.Init()
	}
}

// MapDeps ...
func (ps Pkgs) MapDeps(mapper func(parent *Pkg, dependency *Pkg)) {
	for _, p := range ps {
		p.MapDeps(mapper)
	}
}
