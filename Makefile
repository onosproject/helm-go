export CGO_ENABLED=0
export GO111MODULE=on

.PHONY: build

build: # @HELP build the Go binaries and run all validations (default)
build: test

generate-client: # @HELP generate k8s client interfaces and implementations
generate-client:
	go run github.com/onosproject/helm-go/cmd/generate-client ./build/client/client.yaml ./pkg/kubernetes

test: # @HELP run the unit tests and source code validation
test: linters license_check build deps
	go test github.com/onosproject/helm-go/pkg/...
	go test github.com/onosproject/helm-go/cmd/...

coverage: # @HELP generate unit test coverage data
coverage: build deps license_check
	#./build/bin/coveralls-coverage


linters: # @HELP examines Go source code and reports coding problems
	golangci-lint run

deps: # @HELP ensure that the required dependencies are in place
	go build -v ./...
	bash -c "diff -u <(echo -n) <(git diff go.mod)"
	bash -c "diff -u <(echo -n) <(git diff go.sum)"


license_check: # @HELP examine and ensure license headers exist
	@if [ ! -d "../build-tools" ]; then cd .. && git clone https://github.com/onosproject/build-tools.git; fi
	./../build-tools/licensing/boilerplate.py -v --rootdir=${CURDIR}


all: build test

help:
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST) \
    | sort \
    | awk ' \
        BEGIN {FS = ": *# *@HELP"}; \
        {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}; \
    '
