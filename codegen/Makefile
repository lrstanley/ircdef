.DEFAULT_GOAL := run-codegen

GOPATH := $(shell go env | grep GOPATH | sed 's/GOPATH="\(.*\)"/\1/')
PATH := $(GOPATH)/bin:$(PATH)
export $(PATH)

# enable Go 1.11.x module support.
export GO111MODULE=on

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

fetch: ## Fetches the necessary dependencies to build.
	go mod download
	go mod tidy
	go mod vendor

clean-gen: ## Clean all generated files and folders in the main package.
	/bin/rm -rfv ../*_gen.go
	/usr/bin/find ../ -mindepth 1 -maxdepth 1 -type d ! -name codegen ! -name vendor ! -name .git -exec rm -rfv "{}" \;

run-codegen: fetch clean-gen ## Generates the code.
	go run *.go --pkg-path "github.com/lrstanley/ircdef" --git.repo "https://github.com/ircdocs/irc-defs.git" --git.branch "gh-pages" --output "../"
