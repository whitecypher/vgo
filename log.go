package main

import "fmt"

var (
	verbose = false
)

// Log to console
func Log(message string) {
	fmt.Println(message)
}

// VerboseLog verbose message to console
func VerboseLog(message string) {
	if !verbose {
		return
	}
	Log(message)
}

// Logf verbose message to console
func Logf(message string, args ...interface{}) {
	Log(fmt.Sprintf(message, args...))
}

// VerboseLogf verbose message to console
func VerboseLogf(message string, args ...interface{}) {
	VerboseLog(fmt.Sprintf(message, args...))
}
