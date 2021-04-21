# protoc-gen-sanitize (PGS)

PGS is a protoc plugin to generate sanitization methods from protobuf messages.

This project uses [protoc-gen-star](https://github.com/lyft/protoc-gen-star) to ease code generation.

See `./proto/asset.proto` for example on how to use it.

## Tests

Do a `make test` to test and view a code generated example.

## Debug

To debug in vscode (not working well right know but you can try), edit the `test` task in the Makefile to give the path of the `protoc-gen-sanitize` script (at the root of this project) instead of the protoc-gen-sanitize binary in the `./bin` dir.
