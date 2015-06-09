# Gapp

Go application dependency management simplified. Gapp wraps the default behaviour of the Go compiler and dependency resolver to ensure you have reproducable builds for your project. The intention is to be compatible with the default behaviour of the go get tool by setting up your $GOPATH with the additional needed paths for you project structure and running additional checks and version switches while dependencies are being resolved.

# Build

```sh
go install
```

# API

Initialize an existing project.
```sh
gapp init
```

Test your project and anything contained in your project src directory. Providing a package name will test only that package from either your project src directory, vendor/src directory, or your default $GOPATH, in this order.
```sh
gapp test [package/name]
```

Test your project and all it's dependencies including your project src directory and vendor packages. Providing a package name will test only that package from either your project src directory, vendor/src directory, or your default $GOPATH, in this order.
```sh
gapp test-all [package/name]
```

Add a dependency and version
```sh
gapp get [{packagename[github.com/codegangsta/cli]} [{version-compatibility[~1.4.1]}]
```

Update a dependency version. If the dependency doesn't exist update will perform a gapp get
```sh
gapp update [{packagename[github.com/codegangsta/cli]} [{version-compatibility[~1.4.1]}]
```

Remove a dependency
```sh
gapp remove {packagename[github.com/codegangsta/cli]}
```

Install the stored dependencies
```sh
gapp install
```

Run the application after installing dependencies
```sh
gapp run
```

Run the tests after installing dependencies
```sh
gapp test
```

Build the application after installing dependencies
```sh
gapp build
```

Run a go command through gapp
```sh
gapp go
```

Unmatched actions should fall through to `go` command automatically
```sh
gapp ...
```

# Rationale
