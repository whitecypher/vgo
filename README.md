# VGO

Lightweight unassuming application vendoring for your go projects. VGO decorates the default behaviour of the Go command to ensure you have reproducable builds for your projects. The intention is to be compatible with the default behaviour of the go command line tools by setting up your project with a vendor directory containing all the dependencies of your project and storing the specific commit hashes of each dependency so builds can reproduced.

The only real requirement is that your GO15VENDOREXPERIMENT environment variable is set to 1. As of Go@1.6 this will be the default setting.

## Install

You will need to have your $GOPATH/bin directory available in your $PATH.
```sh
go install github.com/whitecypher/vgo
```

## Usage

#### Init
Initialize an existing project with a manifest of all automatically resolved dependencies.
```sh
vgo init
```

#### Get
Get a dependency compatible with the optionally specified version, branch, tag, or commit. If the current installed version does not match the required reference it will be updated and the new reference stored in the dependency manifest.
```sh
vgo get ./...
#or
vgo get {packagename}[@{version}]
#or
vgo get {packagename}[@{branch}]
#or
vgo get {packagename}[@{tag}]
#or
vgo get {packagename}[@{commit}]
```
e.g. `vgo get github.com/codegangsta/cli@~1.4.1`

#### Remove
Remove a dependency
```sh
vgo rm {packagename}
#or
vgo remove {packagename}
```
e.g. `vgo remove github.com/codegangsta/cli`

#### Clean
Clean your dependency tree. This automatically scans your application for dependencies and removes any unused vendors from your manifest and vendor directory. While doing this 
```sh
vgo clean
```

#### Vendor
IDEA: Include your vendor dir in your commits. This should prevent vendored repositories from being added to you project as submodules so all source files will be committed along with your own application code.
```sh
vgo vend
```

### Catchall
Unmatched actions should fall through to `go` command automatically. This means that `vgo run` will automatically trigger `go run` with all the same rules and options available to you as the standard go commands.
```sh
vgo ...
```

## Mindset
* easy to use in an idiomatic go manner
* resolve the dependency diamond problem
* flexible to allow multiple (if not all) use cases
