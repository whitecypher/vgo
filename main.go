package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/codegangsta/cli"
)

var (
	packages    = map[string]*Pkg{}
	packageLock = &sync.RWMutex{}
	installPath = os.Getenv("GOPATH")
	cwd         = MustGetwd(os.Getwd())
)

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

// MustGetwd handles the error returned by Getwd or returns the returns the resulting current working directory path.
func MustGetwd(cwd string, err error) string {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return cwd
}

// Gapp entry point
func main() {
	if os.Getenv("GO15VENDOREXPERIMENT") == "1" {
		installPath = path.Join(cwd, "vendor")
	}

	fmt.Println(installPath)

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
			Name:        "get",
			Usage:       "Get packages(s)",
			Description: `Get a specified package (or all specified packages) compatible with the provided version.`,
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
		args := c.Args()
		if len(args) == 0 {
			handleDefault(c)
			return
		}
		fmt.Println("Catchall: ", c.Args())
	}

	// cmd := exec.Command("goimports", c.Args()...)
	// cmd.Stdout = os.Stdout
	// cmd.Stdin = os.Stdin
	// cmd.Stderr = os.Stderr
	// cmd.Run()

	client.Run(os.Args)
}

func handleDefault(c *cli.Context) {
	manifestPath := filepath.Join(cwd, "vgo.yaml")
	p := &Pkg{}

	p.Load(manifestPath)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go p.ResolveImports(wg)
	wg.Wait()

	p.Save(manifestPath)

	fmt.Printf("%+v\n", p)
}

func resolveBaseName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) > 3 {
		parts = parts[0:3]
	}
	return strings.Join(parts, "/")
}

func addToPackagesMap(p *Pkg) {
	packageLock.Lock()
	packages[resolveBaseName(p.Name)] = p
	packageLock.Unlock()

	for _, sp := range p.Deps {
		addToPackagesMap(sp)
	}
}
