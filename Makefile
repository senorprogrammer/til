.PHONY: build help install lint test uninstall

# Set go modules to on and use GoCenter for immutable modules
export GO111MODULE = on
export GOPROXY = https://proxy.golang.org,direct

# Determines the path to this Makefile
THIS_FILE := $(lastword $(MAKEFILE_LIST))

APP=til

## build: builds a local version
build:
	go build -o bin/${APP}
	@echo "Done building"

## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## install: installs a local version of the app
install:
	@echo "Installing ${APP}..."
	@go clean
	@go install -ldflags="-s -w"
	$(eval INSTALLPATH = $(shell which ${APP}))
	@echo "${APP} installed into ${INSTALLPATH}"

## lint: runs a number of code quality checks against the source code
lint:
	golangci-lint run

## test: runs the test suite
test: build
	go test ./...

## uninstall: uninstals a locally-installed version
uninstall:
	@rm ~/go/bin/${APP}