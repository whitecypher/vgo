package main

import (
	"fmt"
	"os"
	"os/exec"
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
	version      = "0.0.0"
)

func main() {
	if vendoring {
		installPath = path.Join(cwd, "vendor")
	}

	verbose = true

	p := NewPkg(".")
	vgo := cli.App("vgo", "Installs the dependencies listed in the manifest at the designated reference point.\nIf no manifest exists, `go in` is implied and run automatically to build dependencies and install them.")
	dry := vgo.BoolOpt("dry", false, "Prevent updates to manifest for trial runs")
	_ = dry
	vgo.Version("v version", version)
	vgo.Spec = "[--dry]"
	vgo.Before = func() {
		p.LoadManifest()
	}
	vgo.After = func() {
		if !*dry {
			p.SaveManifest()
		} else {
			data, err := yaml.Marshal(p)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(string(data))
		}
	}
	vgo.Action = func() {
		if !p.hasManifest {
			// defer to vgo in
			args := os.Args[1:]
			args = append(args, "in")
			cmd := exec.Command("vgo", args...)
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			cmd.Run()
		} else {
			// just install what's in the manifest
			p.InstallDeps()
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
				m := p.Meta()
				p.Init(m)
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

func repoName(name string) string {
	// Remove any vendor path prefixes
	vendorParts := strings.Split(name, "/vendor/")
	name = vendorParts[len(vendorParts)-1]
	// Limit root package name to 3 levels
	parts := strings.Split(name, "/")
	if len(parts) > 3 {
		parts = parts[0:3]
	}
	return strings.Join(parts, "/")
}

// MustGetwd handles the error returned by Getwd or returns the returns the resulting current working directory path.
func MustGetwd(cwd string, err error) string {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return cwd
}
