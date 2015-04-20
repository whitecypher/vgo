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
