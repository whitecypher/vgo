package main

import "fmt"

// Task ...
type Task interface {
	Run() error
}

// InstallTask ...
type InstallTask struct {
	Package        Pkg
	DestinationDir string
}

// Run implements Task interface
func (t InstallTask) Run() error {
	fmt.Printf("Cloning '%s' into '%s'", t.Package.Name, t.DestinationDir)

	t.Package.ResolveImports()
	return nil
}
