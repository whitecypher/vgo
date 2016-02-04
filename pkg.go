package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/whitecypher/vgo/lib/native"
	"gopkg.in/yaml.v2"
)

// Compat version compatibility string e.g. "~1.0.0" or "1.*"
type Compat string

// Pkg ...
type Pkg struct {
	sync.Mutex `yaml:"-"`

	Name   string `yaml:"package,omitempty"`
	Compat Compat `yaml:"compat,omitempty"`
	Ref    string `yaml:"ref,omitempty"`
	Deps   []*Pkg `yaml:"imports,omitempty"`
	Bin    string `yaml:"cvs,omitempty"`
	URL    string `yaml:"url,omitempty"`
}

// Load ...
func (p *Pkg) Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	yaml.Unmarshal(data, p)
	addToPackagesMap(p)
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
func (p *Pkg) ResolveImports(wg *sync.WaitGroup) error {
	defer wg.Done()

	name := p.Name
	if len(name) == 0 || strings.HasSuffix(cwd, p.Name) {
		name = "."
	}

	p.Install()
	p.Checkout()

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
		packageLock.RLock()
		sp, ok := packages[name]
		packageLock.RUnlock()
		if ok {
			wg.Add(1)
			go sp.ResolveImports(wg)
			continue
		}

		dep := &Pkg{Name: name, Compat: Compat("master")}
		addToPackagesMap(dep)
		wg.Add(1)
		go dep.ResolveImports(wg)
		p.Lock()
		p.Deps = append(p.Deps, dep)
		p.Unlock()
	}
	return nil
}

// Install the package
func (p *Pkg) Install() error {
	if p.Name == "." || strings.HasSuffix(cwd, p.Name) {
		return nil
	}

	p.Lock()
	defer p.Unlock()

	p.ResolveCVS()

	dir := path.Join(installPath, resolveBaseName(p.Name))
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		fmt.Println("Install", dir)

		cmd := exec.Command(p.Bin, "clone", p.URL, dir)
		// cmd.Dir = installPath
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stdout
		err := cmd.Run()
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

// Checkout switches the package version to the commit nearest maching the Compat string
func (p *Pkg) Checkout() error {
	if p.Name == "." || strings.HasSuffix(cwd, p.Name) {
		return nil
	}

	p.Lock()
	defer p.Unlock()

	p.ResolveCVS()

	dir := path.Join(installPath, resolveBaseName(p.Name))
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		fmt.Println("Checkout", dir)

		cmd := exec.Command(p.Bin, "checkout", "-f", string(p.Compat))
		cmd.Dir = dir
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// ResolveCVS resolves the CVS properties (Bin, URL) for the package
func (p *Pkg) ResolveCVS() (bin, url string) {
	parts := strings.Split(p.Name, "/")
	service := parts[0]
	repo := strings.Join(parts[1:], "/")

	if len(p.Bin) == 0 {
		switch service {
		case "github.com":
			p.Bin = "git"
		}
	}

	if len(url) == 0 {
		switch service {
		case "github.com":
			p.URL = fmt.Sprintf("git@github.com:%s.git", repo)
		}
	}

	return p.Bin, p.URL
}
