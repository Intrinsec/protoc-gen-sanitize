# protoc-gen-sanitize (PGS)

PGS is a protoc plugin to generate sanitization methods from protobuf messages.

This project uses [protoc-gen-star](https://github.com/lyft/protoc-gen-star) to ease code generation.

## Parameters

The plugin can take `strict` as a parameter to return an error if at least one of the following is found:
    - protobuf strings without explicit sanitization options (to enable or disable it)
    - sanitization is defined on a disabled message.

This parameter permits to stop CI/CD or build process.

It can be given as `--sanitize_out=strict:...` or on the dedicated command line flag `--sanitize_opt=strict`.

## Tests

Do a `make test` to test and view a code generated example.

## Examples

In the `tests` directory there are various samples of `.proto` files to show usage examples.

## Debug

To debug in vscode (not working well right know but you can try), edit the `test` task in the Makefile to give the path of the `protoc-gen-sanitize` script (at the root of this project) instead of the protoc-gen-sanitize binary in the `./bin` dir.
