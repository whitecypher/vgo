# Gapp
Go application dependency and build management simplified

# Build
```sh
GOPATH=$(pwd):$GOPATH go install
```

# API

Clone a repository and symlink to GOPATH using package name from gapp.json

	gapp clone {git@github.com/whitecypher/gapp}

Initialize an existing project with

	gapp init [package/name [version]]

Bump the version number. Arguments can be on of "major", "minor", "patch"

	gapp bump {major|minor|patch]

Add a package to the project

	gapp get [package/url]
	gapp add {package/url}

Install project dependencies

	gapp install

Update one or all installed packages to the latest compatible version

	gapp update [package/url|package/name]

Run the application

	gapp run

Build the application

	gapp build

Test project, or package with optional recursive flag to include subpackages in tests.
Each package must be run independantly since coverage reports cannot natively be run on
multipackage tests.

	gapp test [-r] [package/name]
	gapp test [--recursive] [package/name]

Unmatched actions should fall through to `go` command automatically

	gapp ...
