package main

import (
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
	manifestPath = cwd
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
			Name:        "init",
			Usage:       "Initialize vgo project",
			Description: `Scan project for packages, install them if not already vendored and store results into vgo.yaml`,
			Action:      handleInit,
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
	output, err := exec.Command("go", "list", "-json", packageName).CombinedOutput()
	if err != nil {
		fmt.Println("getDepsFromPackage", packageName, err.Error())
		fmt.Println("??", string(output))
		return []string{}
	}
	l := &List{}
	err = json.Unmarshal(output, l)
	if err != nil {
		fmt.Println("getDepsFromPackage", packageName, err.Error())
		return []string{}
	}
	return l.Deps
}

func getImportNameFromPackage(packageName string) string {
	name := packageName
	output, err := exec.Command("go", "list", "-f", "{{ .ImportPath }}", packageName).Output()
	if err == nil {
		name = strings.TrimSpace(string(output))
	}
	// fmt.Println("getImportNameFromPackage", packageName, name)
	return name
}

func resolveDeps(packageName string, findDeps func(string) []string) (deps []string) {
	fmt.Println("resolveDeps", packageName)
	packageName = getImportNameFromPackage(packageName)
	found := []string{}
	for _, dep := range findDeps(packageName) {
		// Skip native packages and vendor packages
		if native.IsNative(dep) {
			continue
		}
		vendorPath := filepath.Join(packageName, "vendor")
		vendored := vendoring && strings.HasPrefix(dep, vendorPath)
		// recurse into subpackages that are not vendored
		if strings.HasPrefix(dep, packageName) && !vendored {
			found = append(found, resolveDeps(dep, findDeps)...)
			continue
		}
		found = append(found, repoName(strings.Trim(strings.TrimPrefix(dep, vendorPath), "/")))
	}
	// Reduce findings to a unique resultset
	has := map[string]bool{}
	for _, dep := range found {
		if _, ok := has[dep]; ok || dep == packageName {
			continue
		}
		has[dep] = true
		deps = append(deps, dep)
	}
	fmt.Println(packageName, "--", strings.Join(deps, " "))
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
