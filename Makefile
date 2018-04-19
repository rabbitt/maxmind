SHELL = bash
TOOLING ?= golang.org/x/tools/cmd/goimports github.com/kardianos/govendor
FORMAT_FILES ?= $(find . -name '*.go' -a -not -regex '.+/vendor/.+')

.ONESHELL:

default: maxmind

maxmind: prep
	@go build -i -o bin/maxmind .

# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@go list -f '{{.Dir}}' ./... | grep -v /vendor/
		| grep -v '.*github.com/rabbitt/maxmind$$'
		| xargs govendor tool vet -all

# prep runs `go generate` to build the dynamically generated
# source files.
prep: format-check
	@echo "==> Rebuilding ffjson stubs..."
	@govendor generate $$(go list ./... | grep -v /vendor/)

# bootstrap the build by downloading additional tools
bootstrap:
	@for tool in $(TOOLING) ; do
		echo "Installing/Updating $$tool"
		go get -u $$tool
	done
	govendor sync

format:
	@echo "==> formating code according to gofmt requirements..."
	gofmt -w $$(find . -name '*.go' -a -not -regex '.+/vendor/.+')

format-check:
	@echo "==> Checking that code complies with gofmt requirements..."
	declare -a files
	files=( $$(gofmt -l $$(find . -name '*.go' -a -not -regex '.+/vendor/.+' | tr ' ' '\n')) )
	if [[ $${#files[@]} -ge 1 ]]; then
		echo "Found $${#files[@]} files that need reformatting:"
		for ((i = 0; i < $${#files[@]}; i++)); do
			echo -e "\t$${files[$$i]}"
		done
		echo "You can reformat all of them using: make format"
	fi

.PHONY: bin default prep test vet bootstrap format format-check
