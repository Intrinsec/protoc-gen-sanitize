// Package main hosts the `protoc-gen-sanitize` plugin.
//
// This file's post-processor was lifted from
// https://github.com/lyft/protoc-gen-star/pull/96 and can be deleted once
// protoc-gen-star ships a release including that PR.
package main

import (
	"strings"

	pgs "github.com/lyft/protoc-gen-star"
	"golang.org/x/tools/imports"
)

type goImports struct {}

// GoImports returns a PostProcessor that run goimports on any files ending in ".go"
func GoImports() pgs.PostProcessor { return goImports{} }

func (g goImports) Match(a pgs.Artifact) bool {
	var n string

	switch a := a.(type) {
	case pgs.GeneratorFile:
		n = a.Name
	case pgs.GeneratorTemplateFile:
		n = a.Name
	case pgs.CustomFile:
		n = a.Name
	case pgs.CustomTemplateFile:
		n = a.Name
	default:
		return false
	}

	return strings.HasSuffix(n, ".go")
}

func (g goImports) Process(in []byte) ([]byte, error) {
	// We do not want to give a filename here, ever
	return imports.Process("", in, nil)
}
