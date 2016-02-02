package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
	"github.com/whitecypher/vgo/lib/native"
)

// Pkg ...
type Pkg struct {
	Name   string `yaml:"package,omitempty"`
	Compat Compat `yaml:"compat,omitempty"`
	Ref    string `yaml:"ref,omitempty"`
	Deps   []*Pkg `yaml:"imports,omitempty"`
}

// Compat version compatibility string e.g. "~1.0.0" or "1.*"
type Compat string

// DepVerMap contains a mapping of dependency urls to version compatibility strings
//
// deps := DepVerMap{
//   "github.com/whitecypher/gapp": "1.*",
// }
type DepVerMap map[string]Compat

// List structure from go list -json .
type List struct {
	Dir        string   `json:"Dir"`
	Importpath string   `json:"ImportPath"`
	Name       string   `json:"Name"`
	Target     string   `json:"Target"`
	Root       string   `json:"Root"`
	Gofiles    []string `json:"GoFiles"`
	Imports    []string `json:"Imports"`
	Deps       []string `json:"Deps"`
}

// Gapp entry point
func main() {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(cwd)

	client := cli.NewApp()
	client.Name = "vgo"
	client.Usage = "Vendoring for your go projects"
	client.Version = "0.0.0"
	client.Authors = []cli.Author{
		{
			Name:  "Merten van Gerven",
			Email: "merten.vg@gmail.com",
		},
	}
	client.Commands = []cli.Command{
		{
			Name:        "init",
			Usage:       "Initialize project",
			Description: `Detect imports and initialise the application with a vgo.yaml`,
			Action: func(c *cli.Context) {
				manifestPath := filepath.Join(cwd, "vgo.yaml")
				// _, err := os.Stat(manifestPath)
				// if err != nil {
				// 	fmt.Println(err.Error())
				// 	os.Exit(1)
				// }
				p := &Pkg{}
				p.Load(manifestPath)
				p.ResolveImports(".")
				fmt.Printf("%+v\n", p)
				p.Save(manifestPath)
			},
		},
		{
			Name:        "get",
			Usage:       "Get packages(s)",
			Description: `Get a package (or all added packages) compatible with the provided version.`,
			Action: func(c *cli.Context) {
				fmt.Println("get: ", c.Args())
			},
		},
		{
			Name:        "update",
			Aliases:     []string{"up"},
			Usage:       "Update package(s)",
			Description: `Update a package (or all added packages) compatible with the provided version.`,
			Action: func(c *cli.Context) {
				fmt.Println("update: ", c.Args())
			},
		},
		{
			Name:        "remove",
			Aliases:     []string{"rm"},
			Usage:       "Remove package(s)",
			Description: `Remove the specified package(s)`,
			Action: func(c *cli.Context) {
				fmt.Println("remove: ", c.Args())
			},
		},
	}
	client.Action = func(c *cli.Context) {
		fmt.Println("Catchall: ", c.Args())
	}

	// cmd := exec.Command("goimports", c.Args()...)
	// cmd.Stdout = os.Stdout
	// cmd.Stdin = os.Stdin
	// cmd.Stderr = os.Stderr
	// cmd.Run()

	client.Run(os.Args)
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
func (p *Pkg) ResolveImports(name string) error {
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
			p.ResolveImports(name)
			continue
		}
		// Skip packages already in manifest
		name := resolveBaseName(name)
		if p.HasImport(name) {
			p.ResolveImports(name)
			continue
		}

		dep := &Pkg{Name: name}
		dep.ResolveImports(name)
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

func resolveBaseName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) > 3 {
		parts = parts[0:3]
	}
	return strings.Join(parts, "/")
}
