package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
)

var (
	tasks    chan Task
	done     chan bool
	packages map[string]Pkg
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

// Gapp entry point
func main() {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(cwd)

	tasks = make(chan Task, 100)
	done = make(chan bool)

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
				p.ResolveImports()
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

func resolveBaseName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) > 3 {
		parts = parts[0:3]
	}
	return strings.Join(parts, "/")
}
