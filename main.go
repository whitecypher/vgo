package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

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
	version      = "0.0.0"
)

func main() {
	if vendoring {
		installPath = path.Join(cwd, "vendor")
	}

	verbose = true

	r := NewRepo(".", cwd)
	vgo := cli.App("vgo", "Installs the dependencies listed in the manifest at the designated reference point.\nIf no manifest exists, `go in` is implied and run automatically to build dependencies and install them.")
	dry := vgo.BoolOpt("dry", false, "Prevent updates to manifest for trial runs")
	vgo.Version("v version", version)
	vgo.Spec = "[--dry]"
	vgo.Before = func() {
		r.LoadManifest()
	}
	vgo.After = func() {
		if !*dry {
			r.SaveManifest()
		} else {
			data, err := yaml.Marshal(r)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(string(data))
		}
	}
	vgo.Action = func() {
		if !r.hasManifest {
			p := NewPkg(r.FQN(), r.path)
			p.Print(os.Stdout, "  ")
			p.MapDeps(PackageRepoMapper)
		}
		r.InstallDeps()

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
				p := NewPkg(r.FQN(), r.path)
				p.Print(os.Stdout, "  ")
				p.MapDeps(PackageRepoMapper)
				r.InstallDeps()
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

	vgo.Run(os.Args)
}
