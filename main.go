// Copyright 2021-2022 Intrinsec. All rights reserved.

package main

import (
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
)

func main() {

	sanitizeModule := Sanitize()

	pgs.Init(pgs.DebugEnv("DEBUG_PG_SAN")).
		RegisterModule(sanitizeModule).
		RegisterPostProcessor(pgsgo.GoFmt()).
		RegisterPostProcessor(GoImports()).
		Render()
	sanitizeModule.ExitCheck()
}
