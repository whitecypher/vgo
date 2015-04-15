// Gapp - Go Application Tool
//
//	#
// 	gapp init
//
//	#
// 	gapp get [{packagename[github.com/codegangsta/cli]} [{version-compatibility[~1.4.1]}]
//
//  #
//  gapp update [{packagename[github.com/codegangsta/cli]} [{version-compatibility[~1.4.1]}]
//
//  #
//  gapp remove {packagename[github.com/codegangsta/cli]}
//
//	#
//	gapp install
//
// 	#
// 	gapp run
//
// 	#
//  gapp test
//
// 	#
//  gapp build
//
//  #
//  gapp go
//
package main

import (
	"codegangsta/cli"
	"fmt"
	"os"
	"os/exec"
	_ "time"
	// "golang.org/x/tools/refactor/importgraph"
)

const (
	GAPP_FILE string = "gapp.json"
	LOCK_FILE string = "gapp.lock"
)

// SourcePath contains a dependency source path e.g. "github.com/whitecypher/gapp"
type SourcePath string

// LocalPath contains path used locally for import into the application
//
//  import "{local-path}"
type LocalPath string

// Version compatibily string e.g. "~1.0.0" or "1.*"
type Version string

// DepVerMap contains a mapping of dependency urls to version compatibility strings
//
//  deps := DepVerMap{
//  	"github.com/whitecypher/gapp": "1.*",
//  }
type DepVerMap map[SourcePath]Version

// Author contains information about the application author
type Author struct {
	// Name of author
	Name string `json:"name"`

	// Email of author
	Email string `json:"email"`
}

type Dep struct {
	// Source contains the package remote path
	Source SourcePath `json:"source"`

	// Local contains the local package path used for import
	Local LocalPath `json:"local"`

	// Version contains the compatibily string e.g. "~1.0.0" or "1.*"
	Version Version `json:"version"`
}

// Dist contains information about an application dependency
type Dist struct {
	*Dep

	// Reference contains the version hash or tag
	Reference string `json:"reference"`
}

// Lock contains information about previous installed depencies and their
// corresponding version reference.
type Lock struct {
	// List of distributables installed
	Dists map[LocalPath]Dist `json:"dists"`
}

// Application contains information about the app under development.
// Primarily to store dependencies and version criteria for the application.
// Persisted in gapp.json.
type Application struct {
	// Name of the application
	Name string `json:"name"`

	// Version of the application
	Version Version `json:"version"`

	// Authors
	Authors []Author `json:"authors"`

	// Dependacies
	Deps DepVerMap `json:"deps"`
}

// Gapp entry point
func main() {
	// Get the current working directory
	cwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	// Add currrent working directory to GOPATH
	os.Setenv("GOPATH", fmt.Sprintf("%s:%s", cwd, os.Getenv("GOPATH")))

	// println(os.Getenv("GOPATH"))

	// Set up the CLI commands
	app := cli.NewApp()
	app.Name = "Gapp"
	app.Usage = "Manage your Go project dependencies"
	app.Version = "0.0.0"
	app.Authors = []cli.Author{
		{
			Name:  "Merten van Gerven",
			Email: "merten.vg@gmail.com",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:        "init",
			Usage:       "Initialise the application",
			Description: `Initialise the application with an gapp.json`,
			Action: func(c *cli.Context) {
				println("init: ", c.Args())
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
				println("get: ", c.Args())
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
				println("update: ", c.Args())
			},
		},
		{
			Name:  "remove",
			Usage: "Remove a package from the project dependencies",
			Description: `Remove the specified package from the project dependencies.

ARGUMENTS: repository-url
   repository-url   Repository url to remove e.g. "github.com/whitecypher/gapp"`,
			Action: func(c *cli.Context) {
				println("update: ", c.Args())
			},
		},
		{
			Name:        "install",
			Usage:       "Build and install your app",
			Description: "Install your application to $GOPATH/bin (or $GOROOT/bin and symlink to /usr/local/bin if $GOPATH is empty)",
			Action: func(c *cli.Context) {
				println("install: ", c.Args())
			},
		},
		{
			Name:  "run",
			Usage: "Run the current application",
			Action: func(c *cli.Context) {
				println("run: ", c.Args())
			},
		},
		{
			Name:  "test",
			Usage: "Test the current application",
			Action: func(c *cli.Context) {
				println("test: ", c.Args())
			},
		},
		{
			Name:  "build",
			Usage: "Build the current application",
			Action: func(c *cli.Context) {
				println("build: ", c.Args())
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
				//println("build: ", c.Args())
			},
		},
	}

	// Run the client app
	app.Run(os.Args)
}
