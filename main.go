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
	gosrcpath    = filepath.Join(gopath, "src")
	installPath  = gosrcpath
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

// ListResult structure from go list -json .
type ListResult struct {
	ImportPath string   `json:"import_path"`
	Deps       []string `json:"deps"`
}

func getDepsFromPackage(packageName string) []ListResult {
	path := "."
	if strings.Trim(packageName, ".") != "" {
		path = packageName
	}
	tmpl := `{ "import_path":"{{ .ImportPath }}", "deps":["{{ join .Deps "\",\"" }}"], "test_deps":["{{ join .TestImports "\",\"" }}"]}`
	output, err := exec.Command("go", "list", "-f", tmpl, path).Output()
	if err != nil {
		fmt.Println("getDepsFromPackage 1", packageName, err.Error())
		// fmt.Println("??", string(output))
		return []ListResult{}
	}
	var l []ListResult
	// Combine results into a json array of values
	b := &bytes.Buffer{}
	b.WriteString("[")
	b.Write(bytes.Join(bytes.Split(bytes.TrimSpace(output), []byte("\n")), []byte(",")))
	b.WriteString("]")
	// fmt.Println(b.String())
	err = json.Unmarshal(b.Bytes(), &l)
	if err != nil {
		fmt.Println("getDepsFromPackage 2", packageName, err.Error())
		return []ListResult{}
	}
	return l
}

func resolveDeps(packageName string, findDeps func(string) []ListResult) (deps []string) {
	found := findDeps(packageName)
	if len(found) > 0 {
		packageName = found[0].ImportPath
	}
	vendorPath := filepath.Join(packageName, "vendor")
	has := map[string]bool{}
	for _, lr := range found {
		// Iterate dependencies to extract unique items
		for _, dep := range lr.Deps {
			// Skip native packages and vendor packages
			if native.IsNative(dep) {
				continue
			}
			// Skip subpackages except if vendored
			vendorPkg := vendoring && strings.HasPrefix(dep, vendorPath)
			if strings.HasPrefix(dep, packageName) && !vendorPkg {
				continue
			}
			dep := repoName(dep)
			// Skip packages already found
			if _, ok := has[dep]; ok || dep == packageName {
				continue
			}
			has[dep] = true
			deps = append(deps, dep)
		}
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
