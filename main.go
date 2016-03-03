package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/jawher/mow.cli"
)

var (
	gopath       = os.Getenv("GOPATH")
	gosrcpath    = filepath.Join(gopath, "src")
	installPath  = gosrcpath
	cwd          = MustGetwd(os.Getwd())
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
	name := strings.Trim(strings.TrimPrefix(cwd, gosrcpath), "/")
	r := NewRepo(name, nil)
	vgo := cli.App("vgo", "Installs the dependencies listed in the manifest at the designated reference point.\nIf no manifest exists, `go in` is implied and run automatically to build dependencies and install them.")
	dry := vgo.BoolOpt("dry", false, "Prevent updates to manifest for trial runs")
	vgo.Version("v version", version)
	vgo.Spec = "[--dry]"
	vgo.Before = func() {
		r.LoadManifest()
	}
	vgo.After = func() {
		if !*dry {
			err := r.SaveManifest()
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
	}
	vgo.Action = func() {
		if r.hasManifest {
			r.InstallDeps()
			return
		}
		if len(r.Main) > 0 {
			for _, m := range r.Main {
				NewPkg(path.Join(name, m), cwd, nil)
			}
		} else {
			NewPkg(name, cwd, nil)
		}

		// pass command through to go - need to find another way to do this
		// if len(os.Args) > 1 {
		// 	args := os.Args[1:]
		// 	if args[0] == "--dry" {
		// 		args = args[1:]
		// 	}
		// cmd := exec.Command("go", args...)
		// cmd.Stdout = os.Stdout
		// cmd.Stdin = os.Stdin
		// cmd.Stderr = os.Stderr
		// cmd.Run()
		// }
	}
	vgo.Command(
		"in",
		"Scans your project to create/update a manifest of all automatically resolved dependencies",
		func(cmd *cli.Cmd) {
			cmd.Action = func() {
				if len(r.Main) > 0 {
					for _, m := range r.Main {
						NewPkg(path.Join(name, m), cwd, nil)
					}
				} else {
					NewPkg(name, cwd, nil)
				}
			}
		},
	)
	vgo.Command(
		"get",
		"Get a dependency compatible with the optionally specified version, branch, tag, or commit",
		func(cmd *cli.Cmd) {
			update := cmd.BoolOpt("u", false, "Update the package to the latest compatible reference")
			paths := cmd.StringsArg("PKG", []string{}, "Package to be installed {package/import/path#compat}")
			cmd.Action = func() {
			}
			_ = update
			_ = paths
		},
	)
	vgo.Command(
		"rm",
		"Scans your project to create/update a manifest of all automatically resolved dependencies",
		func(cmd *cli.Cmd) {
			paths := cmd.StringsArg("PKG", []string{}, "Package to be removed {package/import/path}")
			cmd.Action = func() {
			}
			_ = paths
		},
	)
	vgo.Command(
		"add-main",
		"Add a main (entrypoint) package to the project manifest",
		func(cmd *cli.Cmd) {
			paths := cmd.StringsArg("PATH", []string{}, "Relative path to main package (entrypoint)")
			cmd.Action = func() {
				for _, path := range *paths {
					r.AddMain(path)
				}
			}
			_ = paths
		},
	)
	vgo.Command(
		"rm-main",
		"Remove a main (entrypoint) package from the project manifest",
		func(cmd *cli.Cmd) {
			paths := cmd.StringsArg("PATH", []string{}, "Relative path to main package (entrypoint)")
			cmd.Action = func() {
				for _, path := range *paths {
					r.AddMain(path)
				}
			}
			_ = paths
		},
	)

	vgo.Run(os.Args)
}
