This project is still a work in progress. Don't use it yet.
===========================================================

VGO
===

Lightweight unassuming application vendoring for your go projects. VGO decorates the default behaviour of the Go command to ensure you have reproducable builds for your projects. The intention is to be compatible with the default behaviour of the go command line tools by setting up your project with a vendor directory containing all the dependencies of your project and storing the specific commit hashes of each dependency so builds can reproduced.

The only real requirement is that your GO15VENDOREXPERIMENT environment variable is set to 1. As of Go#1.6 this will be the default setting.

Install
-------

You will need to have your $GOPATH/bin directory available in your $PATH.

```sh
go install github.com/whitecypher/vgo
```

Usage
-----

#### Default

Scans your project to create/update a manifest of all automatically resolved dependencies. Dependencies not yet added will be added, and packages no longer in use are removed.

```sh
vgo
```

#### Sync

Sync installs the dependencies listed in the manifest at the designated reference point. Where no reference point is available in the manifest the last reference compatible with the required version, branch, tag, or commit will be installed. The installed reference point will be stored in the manifest unless otherwise suppressed using the `-dry` option.

```sh
vgo [-dry] sync
```

#### Get

Get a dependency compatible with the optionally specified version, branch, tag, or commit. If the current installed version does not match the required reference it will be updated and the new reference stored in the dependency manifest. When the `-u` flag is provided a dependency will be updated to the latest reference compatible with the stored version, branch, tag, or commit. If a {packagename} with a [#{version}] is given, the `-u` option is implied.

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

e.g. `vgo get github.com/codegangsta/cli#~1.4.1`

#### Remove

Remove a dependency

```sh
vgo rm {packagename}
#or
vgo remove {packagename}
```

e.g. `vgo remove github.com/codegangsta/cli`

#### Catchall

Unmatched actions should fall through to `go` command automatically. This means that `vgo run` will automatically trigger `go run` with all the same rules and options available to you as the standard go commands.

```sh
vgo ...
```

Mindset
-------

-	easy to use in an idiomatic go manner
-	resolve the dependency diamond problem
-	flexible to allow multiple (if not all) use cases
