package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/whitecypher/vgo/lib/native"
)

var (
	gopath       = os.Getenv("GOPATH")
	installPath  = filepath.Join(gopath, "src")
	cwd          = MustGetwd(os.Getwd())
	manifestPath = filepath.Join(cwd, "vgo.yaml")
	vendoring    = os.Getenv("GO15VENDOREXPERIMENT") == "1"
)

func main() {
	if vendoring {
		installPath = path.Join(cwd, "vendor")
	}

	verbose = true

	client := cli.NewApp()
	client.Name = "vgo"
	client.Usage = "Vendoring for your go projects"
	client.Version = "0.0.0"
	// client.Authors = []cli.Author{
	// 	{
	// 		Name:  "Merten van Gerven",
	// 		Email: "merten.vg@gmail.com",
	// 	},
	// }
	client.Commands = []cli.Command{
		{
			Name:        "pin",
			Usage:       "Initialize vgo project",
			Description: `Scan project for packages, install them if not already vendored and store results into vgo.yaml`,
			Action:      handlePin,
		},
		{
			Name: "dry",
			// Aliases:     []string{"dr"},
			Usage:       "Scan for packages(s)",
			Description: `Scan project for used packages and print results to the terminal. Runs recursively into packages already installed if possible.`,
			Action:      handleDryRun,
		},
		{
			Name:        "install",
			Usage:       "Install packages(s)",
			Description: `Install packages at the stored revision or compatibility`,
			Action:      handleInstall,
		},
		{
			Name:        "get",
			Usage:       "Get packages(s)",
			Description: `Get a specified package (or all specified packages) compatible with the provided version.`,
			Action:      handleGet,
		},
		{
			Name: "update",
			// Aliases:     []string{"up"},
			Usage:       "Update package(s)",
			Description: `Update a package (or all added packages) compatible with the provided version.`,
			Action:      handleUpdate,
		},
		{
			Name: "remove",
			// Aliases:     []string{"rm"},
			Usage:       "Remove package(s)",
			Description: `Remove the specified package(s)`,
			Action:      handleRemove,
		},
		{
			Name:        "clean",
			Usage:       "Remove package(s)",
			Description: `Remove the specified package(s)`,
			Action:      handleClean,
		},
	}
	client.Action = handleDefault

	// cmd := exec.Command("goimports", c.Args()...)
	// cmd.Stdout = os.Stdout
	// cmd.Stdin = os.Stdin
	// cmd.Stderr = os.Stderr
	// cmd.Run()

	client.Run(os.Args)
}

func repoName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) > 3 {
		parts = parts[0:3]
	}
	return strings.Join(parts, "/")
}

func getDepsFromPackage(packageName string) []string {
	b := &bytes.Buffer{}
	cmd := exec.Command("go", "list", "-json", packageName)
	cmd.Stdout = b
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		return []string{}
	}
	l := &List{}
	// fmt.Println(string(b.Bytes()))
	err = json.Unmarshal(b.Bytes(), l)
	if err != nil {
		fmt.Println(err.Error())
		return []string{}
	}
	return l.Deps
}

func getImportNameFromPackage(packageName string) string {
	b := &bytes.Buffer{}
	cmd := exec.Command("go", "list", "-f", "{{ .ImportPath }}", packageName)
	cmd.Stdout = b
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
	return strings.TrimSpace(string(b.Bytes()))
}

func resolveDeps(packageName string, findDeps func(string) []string) (deps []string) {
	packageName = getImportNameFromPackage(packageName)
	foundDeps := findDeps(packageName)
	// fmt.Println("==", packageName)
	// fmt.Printf("%+v\n", foundDeps)
	for _, dep := range foundDeps {
		// fmt.Println("--", dep)
		// Skip native packages and vendor packages
		if native.IsNative(dep) {
			continue
		}
		vendorPath := filepath.Join(packageName, "vendor")
		vendored := vendoring && strings.HasPrefix(dep, vendorPath)
		// fmt.Println(dep, vendorPath, vendored)
		// recurse into subpackages that are not vendored
		if strings.HasPrefix(dep, packageName) && !vendored {
			deps = append(deps, resolveDeps(dep, findDeps)...)
			continue
		}
		deps = append(deps, strings.Trim(strings.TrimPrefix(dep, vendorPath), "/"))
	}
	return
}

// MustGetwd handles the error returned by Getwd or returns the returns the resulting current working directory path.
func MustGetwd(cwd string, err error) string {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return cwd
}
