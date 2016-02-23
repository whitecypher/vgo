package main

import (
	"encoding/json"
	"go/build"
	"io"

	"github.com/whitecypher/vgo/lib/native"
)

// NewPkg ...
func NewPkg(name, dir string) *Pkg {
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
	Deps []*Pkg
}

// Init gets the package meta data using the go/build internal package profiler
func (p *Pkg) Init() {
	m, err := build.Import(p.Name, p.Dir, build.ImportMode(0))
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok {
			Logf("Unable to import package %s with error %s", p.Name, err.Error())
		}
	}
	p.Name = m.ImportPath
	p.Dir = m.Dir
	// fmt.Printf("%+v", m.Imports)
	for _, i := range m.Imports {
		if native.IsNative(i) {
			continue
		}
		p.Deps = append(p.Deps, NewPkg(i, p.Dir))
	}
}

// Print ...
func (p *Pkg) Print(w io.Writer) {
	js, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(js)
}
