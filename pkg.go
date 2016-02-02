package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/whitecypher/vgo/lib/native"
	"gopkg.in/yaml.v2"
)

// Compat version compatibility string e.g. "~1.0.0" or "1.*"
type Compat string

// Pkg ...
type Pkg struct {
	Name   string `yaml:"package,omitempty"`
	Compat Compat `yaml:"compat,omitempty"`
	Ref    string `yaml:"ref,omitempty"`
	Deps   []*Pkg `yaml:"imports,omitempty"`
}

// Load ...
func (p *Pkg) Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	yaml.Unmarshal(data, p)
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
func (p *Pkg) ResolveImports() error {
	name := p.Name
	if len(name) == 0 {
		name = "."
	}
	fmt.Println(name)

	b := &bytes.Buffer{}
	cmd := exec.Command("go", "list", "-json", name)
	cmd.Stdout = b
	cmd.Run()

	l := &List{}
	err := json.Unmarshal(b.Bytes(), l)
	if err != nil {
		return err
	}

	ignoreMap := map[string]bool{}
	for _, name := range native.Packages() {
		ignoreMap[name] = true
	}

	for _, name := range l.Deps {
		// Skip native packages
		if _, ok := ignoreMap[name]; ok {
			continue
		}
		// Skip subpackages
		basePath := resolveBaseName(l.Importpath)
		if strings.HasPrefix(name, basePath) {
			continue
		}
		// Skip packages already in manifest
		name := resolveBaseName(name)
		if p.HasImport(name) {
			continue
		}

		dep := &Pkg{Name: name}
		dep.ResolveImports()
		p.Deps = append(p.Deps, dep)
	}
	return nil
}

// HasImport ...
func (p *Pkg) HasImport(name string) bool {
	for _, i := range p.Deps {
		if name == i.Name {
			return true
		}
	}
	return false
}
