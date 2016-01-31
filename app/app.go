package app

// SourcePath contains a dependency source path e.g. "github.com/whitecypher/gapp"
type sourcePath string

// LocalPath contains path used locally for import into the application
//
//  import "{local-path}"
type localPath string

// Version compatibily string e.g. "~1.0.0" or "1.*"
type version string
