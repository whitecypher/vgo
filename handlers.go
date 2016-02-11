package main

import (
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
)

func resolvePkg() *Pkg {
	p := &Pkg{}
	p.Init()
	p.LoadManifest()
	return p
}

func handleDefault(c *cli.Context) {
	// pass command through to go
}

func handleDryRun(c *cli.Context) {
	p := resolvePkg()
	p.ResolveImports()
	data, err := yaml.Marshal(p)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(data))
}

func handleInit(c *cli.Context) {
	p := resolvePkg()
	p.ResolveImports()
	p.SaveManifest()
}

func handleInstall(c *cli.Context) {
	p := resolvePkg()
	p.ResolveImports()
	p.InstallDeps()
	p.SaveManifest()
}

func handleGet(c *cli.Context) {
	fmt.Println("get: ", c.Args())
}

func handleUpdate(c *cli.Context) {
	fmt.Println("update: ", c.Args())
}

func handleRemove(c *cli.Context) {
	fmt.Println("remove: ", c.Args())
}

func handleClean(c *cli.Context) {
	fmt.Println("remove: ", c.Args())
}
