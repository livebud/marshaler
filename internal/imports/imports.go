package imports

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Import struct {
	Path string
	Name string
}

type Imports []*Import

func (i *Imports) Import(path string) (name string, err error) {
	name = assumedName(path)
	for _, imp := range *i {
		if imp.Name == name {
			if imp.Path != path {
				return "", fmt.Errorf("imports: name %s already used for %q", name, imp.Path)
			}
			return name, nil
		}
	}
	// Add the name
	*i = append(*i, &Import{
		Path: path,
		Name: name,
	})
	return name, nil
}

// assumedName returns the assumed name for the import path. It's pulled from:
// https://cs.opensource.google/go/x/tools/+/refs/tags/v0.6.0:internal/imports/fix.go;l=1144
func assumedName(importPath string) string {
	base := path.Base(importPath)
	if strings.HasPrefix(base, "v") {
		if _, err := strconv.Atoi(base[1:]); err == nil {
			dir := path.Dir(importPath)
			if dir != "." {
				base = path.Base(dir)
			}
		}
	}
	base = strings.TrimPrefix(base, "go-")
	if i := strings.IndexFunc(base, notIdentifier); i >= 0 {
		base = base[:i]
	}
	return base
}

// notIdentifier reports whether ch is an invalid identifier character.
func notIdentifier(ch rune) bool {
	return !('a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' ||
		'0' <= ch && ch <= '9' ||
		ch == '_' ||
		ch >= utf8.RuneSelf && (unicode.IsLetter(ch) || unicode.IsDigit(ch)))
}
