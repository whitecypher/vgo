# vgo

Lightweight unassuming application vendoring for your go projects. Gapp decorates the default behaviour of the Go compiler and dependency resolver to ensure you have reproducable builds for your project. The intention is to be compatible with the default behaviour of the go get tool by setting up your project with a vendor directory containing all the dependencies of your project and storing the versions of each dependency so builds can reproduced.

The only real requirement is that your GO15VENDOREXPERIMENT environment variable is set to 1

# Build

```sh
go get github.com/whitecypher/vgo
```

# API

Initialize an existing project.
```sh
vgo init
```

Get a dependency compatible with the optionally specified version, tag, or commit. If the current installed version does not match the required compatibility it will be automatically updated. If the -u flag is provided it will update to the latest version matching the required compatibility regardless of the pinned compatible version.
```sh
vgo get [{packagename}[@{version-compatibility}]
```
e.g. `vgo get github.com/codegangsta/cli@~1.4.1`

Remove a dependency
```sh
vgo remove {packagename}
```
e.g. `vgo remove github.com/codegangsta/cli`

Unmatched actions should fall through to `go` command automatically. This means that `vgo run` will automatically trigger `go run` with all the same rules and options available to you as the standard go commands.
```sh
vgo ...
```

# Rationale
Having tried other vendoring packages I've found they all require some odd work arounds or specific GOPATH configuration other than the configurations defined by golang itself.
