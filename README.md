This project is still a work in progress. Don't use it yet.
===========================================================

VGO
===

Lightweight unassuming application vendoring for your go projects. VGO decorates the default behaviour of the Go command to ensure you have reproducable builds for your projects. The intention is to be compatible with the default behaviour of the go command line tools by setting up your project with a vendor directory containing all the dependencies of your project and storing the specific commit hashes of each dependency so builds can reproduced.

The only real requirement is that your GO15VENDOREXPERIMENT environment variable is set to 1. As of Go#1.6 this will be the default setting.

Version tags (e.g. 1.4 or v1.4) are matched according to Semver rules.

Install
-------

You will need to have your $GOPATH/bin directory available in your $PATH.

```sh
go install github.com/whitecypher/vgo
```

Usage
-----

#### Default

Installs the dependencies listed in the manifest at the designated reference point. If no manifest exists, `go in` is implied and run automatically to build dependencies and install them. Where no reference point is available in the manifest the last reference compatible with the required version, branch, tag, or commit will be installed. The installed reference point will be stored in the manifest unless otherwise suppressed using the `-dry` option.

```sh
vgo [-dry]
```

#### In

Scans your project to create/update a manifest of all automatically resolved dependencies. Dependencies not yet added will be added, and packages no longer in use are removed. Result are stored in the vgo manifest file unless executed with the `-dry` option, which may also be inherited from the default `vgo` when `go in` is implied.

```sh
vgo [-dry] in
```

#### Get

Get a dependency compatible with the optionally specified version, branch, tag, or commit. If the current installed reference is not compatible with the required version, branch, tag, or commit it will be updated and the new reference stored in the dependency manifest. This done to ensure manual changes to the manifest will be adhered to when compatibility is compromised. If current reference is compatible (an earlier reference point of the master branch for example) then the stored reference point will be used and the `-u` flag will must be added. When the `-u` flag is provided a dependency will be updated to the latest reference compatible with the stored version, branch, tag, or commit. If a {packagename} with a [#{version|branch|tag|commit}] is given, and differs from that stored in the manifest, the `-u` option is implied.

```sh
vgo get [-u] ./...
#or
vgo get [-u] {packagename}[#{version}]
#or
vgo get [-u] {packagename}[#{branch}]
#or
vgo get [-u] {packagename}[#{tag}]
#or
vgo get [-u] {packagename}[#{commit}]
```

e.g. `vgo get github.com/codegangsta/cli#v1.4.1` or `vgo get github.com/codegangsta/cli#v1.4` or `vgo get github.com/codegangsta/cli#v1`

#### Remove

Remove a dependency

```sh
vgo rm {packagename}
#or
vgo remove {packagename}
```

e.g. `vgo remove github.com/codegangsta/cli`

#### Catchall

Unmatched actions should fall through to `go` command automatically. This means that `vgo run` will automatically trigger `go run` with all the same rules and options available to you as the standard go commands. Vgo will run a `sync` action before deferring to the default go behavior to ensure the action is run on dependable codebase.

```sh
vgo ...
```

Mindset
-------

-	easy to use in an idiomatic go manner
-	resolve the dependency diamond problem
-	flexible to allow multiple (if not all) use cases
