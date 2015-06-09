// Gapp - Go Application Tool2
//
//  # Initialize an existing project with
//  gapp init
//
//  # Test project, or package with optional recursive flag to include subpackages in tests.
//  # Optional deps flag may also be added to also include tests for dependencies.
//  # Each package must be run independantly since coverage reports cannot natively be run on
//  # multipackage tests.
//  gapp test [--recursive (-r)] [--deps (-d)] [package/name]
//
//	# Add a dependency and version
// 	gapp get [{packagename[github.com/codegangsta/cli]} [{tag|branch|commit-hash|version-compatibility[~1.4.1]}]
//
//  # Update a dependency
//  gapp update [{packagename[github.com/codegangsta/cli]} [{tag|branch|commit-hash|version-compatibility[~1.4.1]}]
//
//  # Remove a dependency
//  gapp remove {packagename[github.com/codegangsta/cli]}
//
//	# Install the stored dependencies
//	gapp install
//
// 	# Run the application after installing dependencies
// 	gapp run
//
// 	# Run the tests after installing dependencies
//  gapp test
//
// 	# Build the application after installing dependencies
//  gapp build
//
//  # Run a go command through gapp
//  gapp go
//
//  # Unmatched actions should fall through to `go` command automatically
//  gapp ...
//
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/codegangsta/cli"
)

const (
	GAPP_FILE string = "gapp.json"
	LOCK_FILE string = "gapp.lock"
)

// SourcePath contains a dependency source path e.g. "github.com/whitecypher/gapp"
type sourcePath string

// LocalPath contains path used locally for import into the application
//
//  import "{local-path}"
type localPath string

// Version compatibily string e.g. "~1.0.0" or "1.*"
type version string

// DepVerMap contains a mapping of dependency urls to version compatibility strings
//
//  deps := DepVerMap{
//  	"github.com/whitecypher/gapp": "1.*",
//  }
type depVerMap map[sourcePath]version

// Author contains information about the application author
type author struct {
	// Name of author
	Name string `json:"name"`
	// Email of author
	Email string `json:"email"`
}

// Package contains information about an application dependency
type pkg struct {
	// Source contains the package remote path
	Source sourcePath `json:"source"`
	// Local contains the local package path used for import
	Local localPath `json:"local"`
	// Version contains the compatibily string e.g. "~1.0.0" or "1.*"
	Version version `json:"version"`
}

// Dist contains information about an installed application dependency
type dist struct {
	*pkg

	// Reference contains the version hash or tag
	Reference string `json:"reference"`
}

// Lock contains information about previous installed depencies and their
// corresponding version reference.
type lock struct {
	// List of distributables installed
	Dists map[localPath]dist `json:"dists"`
}

// Application contains information about the app under development.
// Primarily to store dependencies and version criteria for the application.
// Persisted in gapp.json.
type app struct {
	// Name of the application
	Name string `json:"name"`

	// Version of the application
	Description string `json:"description"`

	// Authors
	Authors []author `json:"authors"`

	// Dependacies
	Deps depVerMap `json:"packages"`
}

// Gapp entry point
func main() {
	twinkle.Twinkle()

	// Get the current working directory
	cwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	fmt.Println(cwd)

	// if !isGappProject(cwd) {
	// 	fmt.Println("Is this supposed to be a gapp project? Use `gapp init` to make it a project")
	// 	os.Exit(25)
	// }

	// Add currrent working directory to GOPATH
	os.Setenv("GOPATH", fmt.Sprintf("%s/vendor:%s", cwd, os.Getenv("GOPATH")))

	// println(os.Getenv("GOPATH"))

	// Set up the CLI commands
	client := cli.NewApp()
	client.Name = "Gapp"
	client.Usage = "Manage your Go application dependencies"
	client.Version = "0.0.0"
	client.Authors = []cli.Author{
		{
			Name:  "Merten van Gerven",
			Email: "merten.vg@gmail.com",
		},
	}
	client.Commands = []cli.Command{
		{
			Name:        "init",
			Usage:       "Initialise the application",
			Description: `Initialise the application with an gapp.json`,
			Action: func(c *cli.Context) {
				initialize(c.Args())
			},
		},
		{
			Name:  "get",
			Usage: "Retrieve a packages source files",
			Description: `Retrieve a package (or all added packages) from remote repositories.

   If a repository-url is provided it will be added to the dependencies listed in gapp.json and
   retrieved into your projects 'src' directory. When a version is given, only tags matching the 
   version requirements will be retrieved.

ARGUMENTS: [repository-url [version]]
   repository-url   Optional repository url to add to project e.g. "github.com/whitecypher/gapp"
   version          Optional version compatibility string e.g. "~1.0.0" or "1.0.*"`,
			Action: func(c *cli.Context) {
				fmt.Println("get: ", c.Args())
			},
		},
		{
			Name:  "update",
			Usage: "Update a single package, or all packages",
			Description: `Update a specific package (or all added packages).

   If a repository-url is provided and exists within the dependencies list it will updated to the
   latest compatible version. If not found within the dependencies list it be added in gapp.json and
   retrieved into your projects 'src' directory. When a version is given, gapp.json will be updated 
   and only tags matching the version requirements will be retrieved.

ARGUMENTS: [repository-url [version]]
   repository-url   Optional repository url to update e.g. "github.com/whitecypher/gapp"
   version          Optional Revision to the version compatibility string e.g. "~1.0.0" or "1.0.*"`,
			Action: func(c *cli.Context) {
				fmt.Println("update: ", c.Args())
			},
		},
		{
			Name:  "remove",
			Usage: "Remove a package from the project dependencies",
			Description: `Remove the specified package from the project dependencies.

ARGUMENTS: repository-url
   repository-url   Repository url to remove e.g. "github.com/whitecypher/gapp"`,
			Action: func(c *cli.Context) {
				fmt.Println("update: ", c.Args())
			},
		},
		{
			Name:        "install",
			Usage:       "Build and install your app",
			Description: "Install your application to $GOPATH/bin (or $GOROOT/bin and symlink to /usr/local/bin if $GOPATH is empty)",
			Action: func(c *cli.Context) {
				fmt.Println("install: ", c.Args())
			},
		},
		{
			Name:  "run",
			Usage: "Run the current application",
			Action: func(c *cli.Context) {
				fmt.Println("run: ", c.Args())
			},
		},
		{
			Name:  "test",
			Usage: "Test the current application",
			Action: func(c *cli.Context) {
				fmt.Println("test: ", c.Args())
			},
		},
		{
			Name:  "build",
			Usage: "Build the current application",
			Action: func(c *cli.Context) {
				fmt.Println("build: ", c.Args())
			},
		},
		{
			Name:  "go",
			Usage: "Run go tools through gapp",
			Action: func(c *cli.Context) {
				cmd := exec.Command("go", c.Args()...)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr
				cmd.Run()
			},
		},
		{
			Name:  "goimports",
			Usage: "Run goimports through gapp",
			Action: func(c *cli.Context) {
				cmd := exec.Command("goimports", c.Args()...)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr
				cmd.Run()
			},
		},
	}

	// Run the client app
	client.Run(os.Args)
}
