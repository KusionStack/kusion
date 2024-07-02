SHELL = /bin/bash
PWD:=$(shell pwd)

KCLVM_BUILDER:=kusionstack/kclvm-builder

RUN_IN_DOCKER:=docker run -it --rm
RUN_IN_DOCKER+=-v ~/.ssh:/root/.ssh
RUN_IN_DOCKER+=-v ~/.gitconfig:/root/.gitconfig
RUN_IN_DOCKER+=-v ~/go/pkg/mod:/go/pkg/mod
RUN_IN_DOCKER+=-v ${PWD}:/root/kusion
RUN_IN_DOCKER+=-w /root/kusion ${KCLVM_BUILDER}

GOFORMATER			?= gofumpt
GOFORMATER_VERSION	?= v0.2.0
GOLINTER			?= golangci-lint
GOLINTER_VERSION	?= v1.56.2
COVER_FILE			?= coverage.out
SOURCE_PATHS		?= ./pkg/...

# TODO: fix mirrors
KCLVM_URL_BASE_MIRRORS:=

.DEFAULT_GOAL := help

CCRED=\033[0;31m
CCEND=\033[0m

help:  ## This help message :)
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# If you encounter an error like "panic: permission denied" on MacOS,
# please visit https://github.com/eisenxp/macos-golink-wrapper to find the solution.
test:  ## Run the tests
	go test -gcflags="all=-l -N" -timeout=10m `go list $(SOURCE_PATHS)` ${TEST_FLAGS}

cover:  ## Generates coverage report
	go test -gcflags="all=-l -N" -timeout=10m `go list $(SOURCE_PATHS)` -coverprofile $(COVER_FILE) ${TEST_FLAGS}

cover-html:  ## Generates coverage report and displays it in the browser
	go tool cover -html=$(COVER_FILE)

format:  ## Format source code
	@which $(GOFORMATER) > /dev/null || (echo "Installing $(GOFORMATER)@$(GOFORMATER_VERSION) ..."; go install mvdan.cc/gofumpt@$(GOFORMATER_VERSION) && echo -e "Installation complete!\n")
	@for path in $(SOURCE_PATHS); do ${GOFORMATER} -l -w -e `echo $${path} | cut -b 3- | rev | cut -b 5- | rev`; done;

lint:  ## Lint, will not fix but sets exit code on error
	@which $(GOLINTER) > /dev/null || (echo "Installing $(GOLINTER)@$(GOLINTER_VERSION) ..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLINTER_VERSION) && echo -e "Installation complete!\n")
	$(GOLINTER) run --deadline=10m $(SOURCE_PATHS)

lint-fix:  ## Lint, will try to fix errors and modify code
	@which $(GOLINTER) > /dev/null || (echo "Installing $(GOLINTER)@$(GOLINTER_VERSION) ..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLINTER_VERSION) && echo -e "Installation complete!\n")
	$(GOLINTER) run --deadline=10m $(SOURCE_PATHS) --fix

doc:  ## Start the documentation server with godoc
	@which godoc > /dev/null || (echo "Installing godoc@latest ..."; go install golang.org/x/tools/cmd/godoc@latest && echo -e "Installation complete!\n")
	godoc -http=:6060

upload:  ## Upload kusion bundles to OSS
	# 执行前先配置 OSS 环境变量 OSS_ACCESS_KEY_ID OSS_ACCESS_KEY_SECRET
	go run ./scripts/oss-upload/main.go

clean:  ## Clean build bundles
	# Delete old artifacts
	-rm -f ./pkg/version/z_update_version.go
	-rm -rf ./_build/bundles

build-all: build-local-darwin-all build-local-linux-all build-local-darwin-arm64-all build-local-windows-all ## Build all platforms (darwin, linux, windows)

build-local-darwin:  ## Build kusion tool chain for macOS
	# Delete old artifacts
	-rm -f ./pkg/version/z_update_version.go
	-rm -rf ./_build/bundles/kusion-darwin
	# Update version
	go generate ./pkg/version
	# Build kusion
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ./_build/bundles/kusion-darwin/bin/kusion \
		-ldflags="-s -w" .

build-local-darwin-all: build-local-darwin ## Build kusion for macOS
	# Copy docs
	cp -r ./docs ./_build/bundles/kusion-darwin/docs
	
	# README.md
	cp ./README.md ./_build/bundles/kusion-darwin
	# Build tgz
	cd ./_build/bundles/kusion-darwin && tar -zcvf ../kusion-darwin.tgz .
	cd ./_build/bundles && go run ../../hack/md5file/main.go kusion-darwin.tgz > kusion-darwin.tgz.md5.txt

build-local-darwin-arm64: ## Build kusion tool chain for macOS arm64
	# Delete old artifacts
	-rm -f ./pkg/version/z_update_version.go
	-rm -rf ./_build/bundles/kusion-darwin-arm64
	mkdir -p ./_build/bundles/kusion-darwin-arm64/bin
	mkdir -p ./_build/bundles/kusion-darwin-arm64/kclvm/bin

	# Update version
	go generate ./pkg/version

	# Build kusion
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
		go build -o ./_build/bundles/kusion-darwin-arm64/bin/kusion \
		-ldflags="-s -w" .

build-local-darwin-arm64-all: build-local-darwin-arm64 ## Build kusion for macOS arm64
	# Copy docs
	cp -r ./docs ./_build/bundles/kusion-darwin-arm64/docs

	# README.md
	cp ./README.md ./_build/bundles/kusion-darwin-arm64
	# Build tgz
	cd ./_build/bundles/kusion-darwin-arm64 && tar -zcvf ../kusion-darwin-arm64.tgz .
	cd ./_build/bundles && go run ../../hack/md5file/main.go kusion-darwin-arm64.tgz > kusion-darwin-arm64.tgz.md5.txt

build-local-linux-in-docker: ## Build kusionctl-linux in docker
	${RUN_IN_DOCKER} make build-local-linux

build-local-linux-all-in-docker: ## Build kusionctl-linux with kcl and kclopenapi in docker
	${RUN_IN_DOCKER} make build-local-linux-all

build-local-linux:  ## Build kusion tool chain for linux
	# Delete old artifacts
	-rm -f ./pkg/version/z_update_version.go
	-rm -rf ./_build/bundles/kusion-linux
	mkdir -p ./_build/bundles/kusion-linux/bin

	# Update version
	go generate ./pkg/version

	# Build kusion
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ./_build/bundles/kusion-linux/bin/kusion \
		-ldflags="-s -w" .

build-local-linux-all: build-local-linux  ## Build kusion for linux
	# Copy docs
	cp -r ./docs ./_build/bundles/kusion-linux/docs

	# Copy README.md
	cp ./README.md ./_build/bundles/kusion-linux

	# Build tgz
	cd ./_build/bundles/kusion-linux && tar -zcvf ../kusion-linux.tgz  .
	cd ./_build/bundles && go run ../../hack/md5file/main.go kusion-linux.tgz > kusion-linux.tgz.md5.txt

build-local-windows:  ## Build kusion tool chain for windows
	# Delete old artifacts
	-rm -f ./pkg/version/z_update_version.go
	-rm -rf ./_build/bundles/kusion-windows
	mkdir -p ./_build/bundles/kusion-windows/bin
	mkdir -p ./_build/bundles/kusion-windows/kclvm/bin
	
	# Update version
	go generate ./pkg/version

	# Build kusion
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ./_build/bundles/kusion-windows/bin/kusion.exe \
		-ldflags="-s -w" .

build-local-windows-all: build-local-windows  ## Build kusion for windows
	# Copy docs
	cp -r ./docs ./_build/bundles/kusion-windows/docs

	# Copy README.md
	cp ./README.md ./_build/bundles/kusion-windows
	# Build zip
	cd ./_build/bundles/kusion-windows && zip -r ../kusion-windows.zip .
	cd ./_build/bundles && go run ../../hack/md5file/main.go kusion-windows.zip > kusion-windows.zip.md5.txt

build-image:  ## Build kusion image
	make build-local-linux-all
	docker build -t kusion/kusion .

sh-in-docker:  ## Run a shell in the docker container of kusion
	${RUN_IN_DOCKER} bash

e2e-test:
	# Run e2e test
	hack/run-e2e.sh $(OSTYPE)

.PHONY: test cover cover-html format lint lint-fix doc build-changelog upload clean build-all build-image build-local-linux build-local-windows build-local-linux-all build-local-windows-all e2e-test
