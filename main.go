// Copyright 2021-2022 Intrinsec. All rights reserved.

package main

import (
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
)

// version and commit are populated at build time by goreleaser via
// `-ldflags "-X main.version=... -X main.commit=..."`. Unused at the
// protoc protocol level (plugins have no `--version` flag), but kept so
// `strings <binary> | grep main.version` reveals the build provenance.
var (
	version = "dev"
	commit  = "none"
)

func main() {
	_, _ = version, commit

	sanitizeModule := Sanitize()

	pgs.Init(pgs.DebugEnv("DEBUG_PG_SAN")).
		RegisterModule(sanitizeModule).
		RegisterPostProcessor(pgsgo.GoFmt()).
		RegisterPostProcessor(GoImports()).
		Render()
	sanitizeModule.ExitCheck()
}
