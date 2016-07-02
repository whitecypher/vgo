package version

import (
    "regexp"
)

// Type enum type
type Type int

// VersionType constants
const (
	TypeNone Type = iota
	TypeRef
	TypeNamed
	TypeSemVer
)

var (
    noVersion = Version{
    	Kind: TypeNone,
    }
    regexSemver = regexp.MustCompile("^v?[.0-9]+(!?-[.0-9A-Za-z]+)(!?\\+[.0-9A-Za-z]+)$")
)


// NoVersion returns a Version of TypeNone
func NoVersion() Version {
	return noVersion
}

// FromString creates a Version instance from a string
func FromString(v string) Version {
    k := resolveKindFromString(v)
	return Version{
		Kind: k,
		Ref: v,
	}
}

func resolveKindFromString(v string) Type {
    if len(v) == 0 {
        return TypeNone
    }
    if regexSemver.

    return
}

// Version compatibility string e.g. "1.0.0" or "1"
type Version struct {
	Kind Type
	Ref string
}

// IsCompatibleWith checks if given version is compatible
func (v Version) IsCompatibleWith(t Version) bool {
	if v.Kind == t.Kind && v.Ref == t.Ref {
        return true
	}

	return false
}

func (v Version) String() string {
	return v.Ref
}

// Marshal YAML

// UnmarshalYAML
