// Gapp - Go Application Tool2

//  # Clone a repository and symlink to GOPATH using package name from gapp.json
//  gapp clone {git@github.com/whitecypher/gapp}

//  # Initialize an existing project with
//  gapp init [package/name [version]]

//  gapp bump {major|minor|fix]

//  gapp get [package/url]
//  gapp add {package/url}

//  gapp install
//  gapp update [package/url|package/name]

//  gapp run

//  gapp build

//  # Test project, or package with optional recursive flag to include subpackages in tests.
//  # Each package must be run independantly since coverage reports cannot natively be run on
//  # multipackage tests.
//  gapp test [-r] [package/name]
//  gapp test [--recursive] [package/name]

//  # Unmatched actions should fall through to `go` command automatically
//  gapp ...
//
//
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
	_ "whitecypher/gapp/testutils"
	// "golang.org/x/tools/refactor/importgraph"
)

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
				fmt.Println("init: ", c.Args())
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
				//println("build: ", c.Args())
			},
		},
	}

	// Run the client app
	app.Run(os.Args)
}
