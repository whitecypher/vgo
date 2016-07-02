package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
)

var (
	gopath       = os.Getenv("GOPATH")
	gosrcpath    = filepath.Join(gopath, "src")
	installPath  = gosrcpath
	cwd          = MustGetwd()
	manifestPath = cwd
	vendoring    = os.Getenv("GO15VENDOREXPERIMENT") == "1"
	ingopath     = strings.HasPrefix(cwd, gosrcpath)
	version      = "0.0.0"
)

func main() {
	if vendoring {
		installPath = path.Join(cwd, "vendor")
	}

	if !ingopath {
		fmt.Println("Your project isn't in the gopath. We haven't tested this with VGO yet so we recommend you move you project into your gopath.")
		os.Exit(1)
	}

	verbose = true
	name, err := filepath.Rel(gosrcpath, cwd)
	if err != nil {
		name = filepath.Base(cwd)
	}
	r := NewRepo(name, nil, resolveManifestFilePath(cwd))

	vgo := cli.NewApp()
	vgo.Name = "vgo"
	vgo.Usage = "Installs the dependencies listed in the manifest at the designated reference point.\n   If no manifest exists, use `vgo discover` to resolve dependencies and create one."
	vgo.Version = version
	vgo.EnableBashCompletion = true
	vgo.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "dry",
			Usage: "Prevent updates to manifest for trial runs",
		},
	}
	vgo.Before = func(c *cli.Context) (err error) {
		r.LoadManifest()
		return
	}
	vgo.After = func(c *cli.Context) (err error) {
		if !c.Bool("dry") {
			err = r.SaveManifest()
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			// r.Print("  ", os.Stdout)
			data, err := yaml.Marshal(r)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(string(data))
		}
		return err
	}
	// client.Authors = []cli.Author{
	// 	{
	// 		Name:  "Merten van Gerven",
	// 		Email: "merten.vg@gmail.com",
	// 	},
	// }
	vgo.Commands = []cli.Command{
		{
			Name:        "discover",
			Usage:       "Discover dependencies",
			Description: `Scan project for packages, install them if not already vendored and store results into vgo.yaml`,
			Action: func(c *cli.Context) {
				if len(r.Main) > 0 {
					for _, m := range r.Main {
						NewPkg(path.Join(name, m), cwd, nil)
					}
				} else {
					NewPkg(name, cwd, nil)
				}
			},
		},
		{
			Name:        "get",
			Usage:       "Get a dependency",
			Description: `Get a dependency compatible with the optionally specified version, branch, tag, or commit`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "u",
					Usage: "Update the package to the latest compatible reference",
				},
			},
			Action: func(c *cli.Context) {
			},
		},
		{
			Name:        "remove",
			Aliases:     []string{"rm"},
			Usage:       "Remove a dependency",
			Description: `Remove one or more dependencies matching the given paths`,
			Action: func(c *cli.Context) {
			},
		},
		{
			Name: "main",
			// Aliases:     []string{"up"},
			Usage: "Manage (add/remove) application entry points",
			// Description: ``,
			Subcommands: []cli.Command{
				{
					Name:        "add",
					Usage:       "Add an entrypoint",
					Description: `Add a main (entrypoint) package to the project manifest`,
					Action: func(c *cli.Context) {
						paths := c.Args()
						for _, path := range paths {
							r.AddMain(path)
						}
					},
				},
				{
					Name:        "remove",
					Aliases:     []string{"rm"},
					Usage:       "Remove an entrypoint",
					Description: `Remove a main (entrypoint) package from the project manifest`,
					Action: func(c *cli.Context) {
						paths := c.Args()
						for _, path := range paths {
							r.RemoveMain(path)
						}
					},
				},
			},
		},
	}
	vgo.Action = func(c *cli.Context) {
		if r.hasManifest {
			r.InstallDeps()
		} else {
			Log("No manifest found. Running discover task.")
			if len(r.Main) > 0 {
				for _, m := range r.Main {
					NewPkg(path.Join(name, m), cwd, nil)
				}
			} else {
				NewPkg(name, cwd, nil)
			}
		}
		// pass command through to go
		args := c.Args()
		if len(args) > 0 {
			if args[0] == "--dry" {
				args = args[1:]
			}
			cmd := exec.Command("go", args...)
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}
	vgo.Run(os.Args)
}
