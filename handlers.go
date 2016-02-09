package main

import (
	"fmt"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
)

func handleDefault(c *cli.Context) {
	// pass command through to go
}

func handleDryRun(c *cli.Context) {
	p := &Pkg{}
	p.Load(manifestPath)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go p.ResolveImports(wg, false)
	wg.Wait()
	data, err := yaml.Marshal(p)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(data))
}

func handleInit(c *cli.Context) {
	p := &Pkg{}
	p.Load(manifestPath)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go p.ResolveImports(wg, true)
	wg.Wait()
	p.Save(manifestPath)
}

func handleInstall(c *cli.Context) {
	p := &Pkg{}
	p.Load(manifestPath)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go p.ResolveImports(wg, true)
	wg.Wait()
	p.Save(manifestPath)
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
