SHELL := /bin/bash

empty :=
space := $(empty) $(empty)
NAME := sanitize
PACKAGE := github.com/intrinsec/protoc-gen-$(NAME)

# protoc-gen-go parameters for properly generating the import path for PGV
GO_IMPORT_SPACES := M$(NAME)/$(NAME).proto=${PACKAGE}/$(NAME),\
	Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,\
	Mgoogle/protobuf/duration.proto=github.com/golang/protobuf/ptypes/duration,\
	Mgoogle/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct,\
	Mgoogle/protobuf/timestamp.proto=github.com/golang/protobuf/ptypes/timestamp,\
	Mgoogle/protobuf/wrappers.proto=github.com/golang/protobuf/ptypes/wrappers,\
	Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor
GO_IMPORT:=$(subst $(space),,$(GO_IMPORT_SPACES))

# protoc bundles the well-known types under <protoc>/../include. Fixtures that
# import google/protobuf/*.proto need this on the include path.
PROTOC_INCLUDE := $(shell dirname $(shell which protoc))/../include

.PHONY: build
build: bin/protoc-gen-$(NAME)

.PHONY: install
install: $(NAME)/$(NAME).pb.go
	@go install -mod=vendor -v .

$(NAME)/$(NAME).pb.go: bin/protoc-gen-go $(NAME)/$(NAME).proto
	@cd $(NAME) && protoc -I . \
		--plugin=protoc-gen-go=$(shell pwd)/bin/protoc-gen-go \
		--experimental_allow_proto3_optional \
		--go_opt=paths=source_relative \
		--go_out="${GO_IMPORT}:." $(NAME).proto

bin/protoc-gen-go:
	@GOBIN=$(shell pwd)/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@latest


bin/protoc-gen-$(NAME): $(NAME)/$(NAME).pb.go $(wildcard *.go)
	@GOBIN=$(shell pwd)/bin go install -mod=vendor .

PROTO_FIXTURES := $(shell find tests -name '*.proto' 2>/dev/null)

.PHONY: generate
generate: bin/protoc-gen-go bin/protoc-gen-$(NAME)
	@protoc -I . -I $(PROTOC_INCLUDE) --plugin=protoc-gen-go=$(shell pwd)/bin/protoc-gen-go --go_out="." $(PROTO_FIXTURES)
	@protoc -I . -I $(PROTOC_INCLUDE) --plugin=protoc-gen-$(NAME)=$(shell pwd)/bin/protoc-gen-$(NAME) --$(NAME)_out=. $(PROTO_FIXTURES)

.PHONY: test
test: generate
	@cat tests/entity.pb.$(NAME).go
	@cd tests && go test -mod=vendor -v -coverprofile=cover.out -covermode=atomic .
	@go tool cover -func=tests/cover.out | tail -1

.PHONY: lint
lint: generate
	@golangci-lint run ./...

.PHONY: vuln
vuln:
	@govulncheck ./...

.PHONY: vendor-check
vendor-check:
	@go mod tidy
	@go mod vendor
	@git diff --exit-code -- go.mod go.sum vendor/ \
		|| (echo "vendor/ out of date — run 'go mod tidy && go mod vendor'"; exit 1)

.PHONY: sbom
sbom:
	@syft dir:. -o cyclonedx-json=sbom.cdx.json

.PHONY: license-check
license-check:
	@set -o pipefail; GOFLAGS="-tags=codegenruntime" go-licenses report ./... 2>/dev/null \
	  | awk -F',' 'NR==FNR{ok[$$1]=1;next} {if(!ok[$$3]){print "DISALLOWED:",$$1,$$3;bad=1}} END{exit bad?1:0}' \
	    .license-allowlist.txt -


.PHONY: clean
clean:
	@rm -fv tests/*.pb.$(NAME).go


.PHONY: distclean
distclean: clean
	@rm -fv bin/protoc-gen-go bin/protoc-gen-$(NAME) $(NAME)/$(NAME).pb.go
