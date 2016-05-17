package main

// VersionType enum type
type VersionType int

// VersionType constants
const (
	VersionTypeRef VersionType = iota
	VersionTypeBranch
	VersionTypeTag
	VersionTypeSemVer
)

// VersionFromString creates a Version instance from a string
func VersionFromString(v string) Version {
	return Version{
		Kind: VersionTypeSemVer,
		Ref: v,
	}
}

// Version compatibility string e.g. "~1.0.0" or "1.*"
type Version struct {
	Kind VersionType
	Ref string
}
//
// // IsCompatibleWith checks if given version is compatible
// func (v Version) IsCompatibleWith(t Version) bool {
// 	if v.Kind == t.Kind && v.Kind == VersionTypeSemVer {
// 		vRef := strings.Trim(v.Ref, "vV")
// 		tRef := strings.Trim(t.Ref, "vV")
// 		_ = vRef
// 		_ = tRef
// 	} else {
// 		return v.Ref == t.Ref
// 	}
// 	return false
// }

func (v Version) String() string {
	return v.Ref
}

// Marshal YAML

// UnmarshalYAML
