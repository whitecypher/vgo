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
	p := &Pkg{
		parent: parent,
		Name:   name,
		Dir:    dir,
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
	m, _ := p.Meta()
	p.Name = m.ImportPath

	fmt.Println(strings.Repeat("  ", depth), p.Name)
	var rp *Repo
	if p.parent != nil {
		rp = p.parent.Repo
	}
	p.Repo = NewRepo(p.RepoName(), rp)

	// reload the meta since it might have changed after creating NewRepo
	m, _ = p.Meta()

	// pp := p
	// for err != nil {
	// 	dir := p.VendorPath()
	// 	m, err = build.Import(p.Name, dir, build.ImportMode(0))
	// 	if pp == nil {
	// 		break
	// 	}
	// 	pp = pp.parent
	// }

	depth++

	// if _, ok := err.(*build.NoGoError); !ok {
	// 	Logf("Unable to import package %s in %s with error %s", p.Name, p.Dir, err.Error())
	// 	return
	// }

	for _, i := range m.Imports {
		if native.IsNative(i) {
			continue
		}
		fmt.Println(strings.Repeat("  ", depth), i)
		dep := NewPkg(i, installPath, p)
		if dep.RepoName() == p.RepoName() {
			continue
		}
		p.Repo.AddDep(dep.Repo)
	}

	depth--
}

// VendorPath ...
// func (p *Pkg) VendorPath() string {
// 	if p.parent == nil {
// 		return path.Join(cwd, "vendor")
// 	}
// 	return path.Join(p.parent.VendorPath(), p.Name, "vendor")
// }

// MapDeps ...
// func (p *Pkg) MapDeps(mapper func(parent *Pkg, dependency *Pkg)) {
// 	for _, d := range p.Deps {
// 		mapper(p, d)
// 		d.MapDeps(mapper)
// 	}
// }

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
// func (p *Pkg) Print(w io.Writer, indent string) {
// 	p.print(w, indent, 0)
// }

//
// func (p *Pkg) print(w io.Writer, indent string, depth int) {
// 	w.Write([]byte(strings.Repeat(indent, depth) + p.Name + "\n"))
// 	for _, d := range p.Deps {
// 		d.print(w, indent, depth+1)
// 	}
// }

// Pkgs list
// type Pkgs []*Pkg

// Init all pkgs in list
// func (ps Pkgs) Init() {
// 	for _, p := range ps {
// 		p.Init()
// 	}
// }

// MapDeps ...
// func (ps Pkgs) MapDeps(mapper func(parent *Pkg, dependency *Pkg)) {
// 	for _, p := range ps {
// 		p.MapDeps(mapper)
// 	}
// }
