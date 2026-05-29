//go:build codegenruntime

// File codegenruntime tracks codegen-output runtime dependencies that the
// Go module graph would not otherwise see. `sanitizer.go` emits Go source
// referencing `github.com/microcosm-cc/bluemonday` at codegen time, so the
// dependency must remain in `go.mod` (and `vendor/`) for downstream test
// packages that compile the generated output. Uses a distinct build tag
// (`codegenruntime`) to avoid colliding with the `tools` tag used by
// `lyft/protoc-gen-star`.
package main

import (
	_ "github.com/microcosm-cc/bluemonday"
	_ "google.golang.org/protobuf/types/known/timestamppb"
)
