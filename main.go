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
	"sync"

	"github.com/codegangsta/cli"
	"github.com/whitecypher/vgo/lib/native"
)

var (
	packages     = map[string]*Pkg{}
	packageLock  = &sync.RWMutex{}
	gopath       = os.Getenv("GOPATH")
	installPath  = filepath.Join(gopath, "src")
	cwd          = MustGetwd(os.Getwd())
	manifestPath = filepath.Join(cwd, "vgo.yaml")
)

func main() {
	if os.Getenv("GO15VENDOREXPERIMENT") == "1" {
		installPath = path.Join(cwd, "vendor")
	}

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

func addToPackagesMap(p *Pkg) {
	packageLock.Lock()
	packages[repoName(p.Name)] = p
	packageLock.Unlock()

	for _, sp := range p.Dependencies {
		addToPackagesMap(sp)
	}

	// packageLock.RLock()
	// fmt.Printf("%+v\n----\n", packages)
	// packageLock.RUnlock()
}

func resolveDeps(packageName string) (deps []string) {
	b := &bytes.Buffer{}
	cmd := exec.Command("go", "list", "-json", packageName)
	cmd.Stdout = b
	cmd.Run()
	l := &List{}
	err := json.Unmarshal(b.Bytes(), l)
	if err != nil {
		// TODO: add logging for this error
		return
	}
	ignoreMap := map[string]bool{}
	for _, name := range native.Packages() {
		ignoreMap[name] = true
	}
	for _, dep := range l.Deps {
		// Skip native packages and vendor packages
		if _, ok := ignoreMap[dep]; ok {
			// fmt.Printf("Skipping native package %s\n", name)
			continue
		}

		dep = strings.TrimPrefix(dep, filepath.Join(l.Importpath, "vendor"))

		// recurse into subpackages
		if strings.HasPrefix(dep, l.Importpath) {
			deps = append(deps, resolveDeps(dep)...)
			continue
		}

		// fmt.Println(dep, filepath.Join(l.Importpath, "vendor"))

		deps = append(deps, strings.Trim(dep, "/"))
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
