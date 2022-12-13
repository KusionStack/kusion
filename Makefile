SHELL = /bin/bash
PWD:=$(shell pwd)

# todo
KCLVM_BUILDER:=reg.docker.kusionstack.io.com/kusion/kclvm-builder

RUN_IN_DOCKER:=docker run -it --rm
RUN_IN_DOCKER+=-v ~/.ssh:/root/.ssh
RUN_IN_DOCKER+=-v ~/.gitconfig:/root/.gitconfig
RUN_IN_DOCKER+=-v ~/go/pkg/mod:/go/pkg/mod
RUN_IN_DOCKER+=-v ${PWD}:/root/kusion
RUN_IN_DOCKER+=-w /root/kusion ${KCLVM_BUILDER}

GOFORMATER			?= gofumpt
GOFORMATER_VERSION	?= v0.2.0
GOLINTER			?= golangci-lint
GOLINTER_VERSION	?= v1.50.1
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
	go test -gcflags=all=-l -timeout=10m `go list $(SOURCE_PATHS)` ${TEST_FLAGS}

cover:  ## Generates coverage report
	go test -gcflags=all=-l -timeout=10m `go list $(SOURCE_PATHS)` -coverprofile $(COVER_FILE) ${TEST_FLAGS}

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

# todo: fix macOS-arm64 and windows build
build-all: build-local-darwin-all build-local-ubuntu-all build-local-centos-all build-local-darwin-arm64-all ## build-local-windows-all ## Build all platforms (darwin, linux, windows)

build-local-kusion-darwin:  ## Build kusionctl only for macOS
	# Delete old artifacts
	-rm -f ./pkg/version/z_update_version.go
	-rm -rf ./_build/bundles/kusion-darwin/bin/kusion
	# Update version
	go generate ./pkg/version
	# Build kusion
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ./_build/bundles/kusion-darwin/bin/kusion \
		-ldflags="-s -w" \
		./cmd/kusionctl

build-local-darwin:  ## Build kusion tool chain for macOS
	# Delete old artifacts
	-rm -f ./pkg/version/z_update_version.go
	-rm -rf ./_build/bundles/kusion-darwin
	mkdir -p ./_build/bundles/kusion-darwin/bin
	mkdir -p ./_build/bundles/kusion-darwin/kclvm/bin
	# Update version
	go generate ./pkg/version

	# Build kusion
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ./_build/bundles/kusion-darwin/bin/kusion \
		-ldflags="-s -w" \
		./cmd/kusionctl

build-local-darwin-all: build-local-darwin ## Build kusion & kcl tool chain for macOS
	# Install kclvm darwin
	go run ./scripts/install-kclvm \
		--triple=Darwin \
		--mirrors=${KCLVM_URL_BASE_MIRRORS} \
		--outdir=./_build/bundles/kusion-darwin/kclvm

	# Build kcl-go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ${PWD}/_build/bundles/kusion-darwin/kclvm/bin/kcl-go \
		./scripts/install-kcl-go

	# Build kcl-openapi
	cd ./scripts/install-kcl-openapi && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ${PWD}/_build/bundles/kusion-darwin/bin/kcl-openapi

	# chmod +x
	-chmod +x ./_build/bundles/kusion-darwin/bin/kusion
	-chmod +x ./_build/bundles/kusion-darwin/bin/kcl-openapi
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/bin/kcl
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/bin/kclvm
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/bin/kcl-plugin
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/bin/kcl-doc
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/bin/kcl-test
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/bin/kcl-lint
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/bin/kcl-fmt
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/bin/kcl-vet
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/bin/kcl-go
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/bin/kclvm_cli
	-chmod +x ./_build/bundles/kusion-darwin/kclvm/tools/clang/bin/clang

	# Copy docs
	cp -r ./docs ./_build/bundles/kusion-darwin/docs
	
	# README.md
	cp ./README.md ./_build/bundles/kusion-darwin
	# Build tgz
	cd ./_build/bundles/kusion-darwin && tar -zcvf ../kusion-darwin.tgz .
	cd ./_build/bundles && go run ../../scripts/md5file/main.go kusion-darwin.tgz > kusion-darwin.tgz.md5.txt

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
		-ldflags="-s -w" \
		./cmd/kusionctl

build-local-darwin-arm64-all: build-local-darwin-arm64 ## Build kusion & kcl tool chain for macOS arm64
	# Install kclvm darwin
	go run ./scripts/install-kclvm \
		--triple=Darwin-arm64 \
		--mirrors=${KCLVM_URL_BASE_MIRRORS} \
		--outdir=./_build/bundles/kusion-darwin-arm64/kclvm

	# Build kcl-go
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
		go build -o ${PWD}/_build/bundles/kusion-darwin-arm64/kclvm/bin/kcl-go \
		./scripts/install-kcl-go

	# Build kcl-openapi
	cd ./scripts/install-kcl-openapi && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
		go build -o ${PWD}/_build/bundles/kusion-darwin-arm64/bin/kcl-openapi

	# chmod +x
	-chmod +x ./_build/bundles/kusion-darwin-arm64/bin/kusion
	-chmod +x ./_build/bundles/kusion-darwin-arm64/bin/kcl-openapi
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/bin/kcl
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/bin/kclvm
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/bin/kcl-plugin
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/bin/kcl-doc
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/bin/kcl-test
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/bin/kcl-lint
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/bin/kcl-fmt
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/bin/kcl-vet
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/bin/kcl-go
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/bin/kclvm_cli
	-chmod +x ./_build/bundles/kusion-darwin-arm64/kclvm/tools/clang/bin/clang

	# Copy docs
	cp -r ./docs ./_build/bundles/kusion-darwin-arm64/docs

	# README.md
	cp ./README.md ./_build/bundles/kusion-darwin-arm64
	# Build tgz
	cd ./_build/bundles/kusion-darwin-arm64 && tar -zcvf ../kusion-darwin-arm64.tgz .
	cd ./_build/bundles && go run ../../scripts/md5file/main.go kusion-darwin-arm64.tgz > kusion-darwin-arm64.tgz.md5.txt

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
		-ldflags="-s -w" \
		./cmd/kusionctl

build-local-linux-all:
	@echo -e "$(CCRED)**** The use of build-local-linux-alle is deprecated. Use build-local-ubuntu-all or build-local-centos-all instead. ****$(CCEND)"
	$(MAKE) build-local-ubuntu-all

build-local-ubuntu-all: build-local-linux  ## Build kusion & kcl tool chain for ubuntu
	# Install kclvm ubuntu
	go run ./scripts/install-kclvm \
		--triple=ubuntu \
		--mirrors=${KCLVM_URL_BASE_MIRRORS} \
		--outdir=./_build/bundles/kusion-ubuntu/kclvm

	# Build kcl-go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ${PWD}/_build/bundles/kusion-ubuntu/kclvm/bin/kcl-go \
		./scripts/install-kcl-go

	# Build kcl-openapi
	cd ./scripts/install-kcl-openapi && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ${PWD}/_build/bundles/kusion-ubuntu/bin/kcl-openapi

	# chmod +x
	-chmod +x ./_build/bundles/kusion-linux/bin/kusion
	-chmod +x ./_build/bundles/kusion-ubuntu/bin/kcl-openapi
	-chmod +x ./_build/bundles/kusion-ubuntu/kclvm/bin/kcl
	-chmod +x ./_build/bundles/kusion-ubuntu/kclvm/bin/kclvm
	-chmod +x ./_build/bundles/kusion-ubuntu/kclvm/bin/kcl-plugin
	-chmod +x ./_build/bundles/kusion-ubuntu/kclvm/bin/kcl-doc
	-chmod +x ./_build/bundles/kusion-ubuntu/kclvm/bin/kcl-test
	-chmod +x ./_build/bundles/kusion-ubuntu/kclvm/bin/kcl-lint
	-chmod +x ./_build/bundles/kusion-ubuntu/kclvm/bin/kcl-fmt
	-chmod +x ./_build/bundles/kusion-ubuntu/kclvm/bin/kcl-vet
	-chmod +x ./_build/bundles/kusion-ubuntu/kclvm/bin/kcl-go
	-chmod +x ./_build/bundles/kusion-ubuntu/kclvm/bin/kclvm_cli
	# linux use native clang and need not chmod

	# Copy docs
	cp -r ./docs ./_build/bundles/kusion-ubuntu/docs

	# Copy README.md
	cp ./README.md ./_build/bundles/kusion-ubuntu

	# Copy kusion
	cp -r ./_build/bundles/kusion-linux/bin ./_build/bundles/kusion-ubuntu

	# Build tgz
	cd ./_build/bundles/kusion-ubuntu && tar -zcvf ../kusion-ubuntu.tgz  .
	cd ./_build/bundles && go run ../../scripts/md5file/main.go kusion-ubuntu.tgz > kusion-ubuntu.tgz.md5.txt

build-local-centos-all: build-local-linux  ## Build kusion & kcl tool chain for linux
	# Install kclvm linux
	go run ./scripts/install-kclvm \
		--triple=centos \
		--mirrors=${KCLVM_URL_BASE_MIRRORS} \
		--outdir=./_build/bundles/kusion-centos/kclvm

	# Build kcl-go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ${PWD}/_build/bundles/kusion-centos/kclvm/bin/kcl-go \
		./scripts/install-kcl-go

	# Build kcl-openapi
	cd ./scripts/install-kcl-openapi && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ${PWD}/_build/bundles/kusion-centos/bin/kcl-openapi

	# chmod +x
	-chmod +x ./_build/bundles/kusion-linux/bin/kusion
	-chmod +x ./_build/bundles/kusion-centos/bin/kcl-openapi
	-chmod +x ./_build/bundles/kusion-centos/kclvm/bin/kcl
	-chmod +x ./_build/bundles/kusion-centos/kclvm/bin/kclvm
	-chmod +x ./_build/bundles/kusion-centos/kclvm/bin/kcl-plugin
	-chmod +x ./_build/bundles/kusion-centos/kclvm/bin/kcl-doc
	-chmod +x ./_build/bundles/kusion-centos/kclvm/bin/kcl-test
	-chmod +x ./_build/bundles/kusion-centos/kclvm/bin/kcl-lint
	-chmod +x ./_build/bundles/kusion-centos/kclvm/bin/kcl-fmt
	-chmod +x ./_build/bundles/kusion-centos/kclvm/bin/kcl-vet
	-chmod +x ./_build/bundles/kusion-centos/kclvm/bin/kcl-go
	-chmod +x ./_build/bundles/kusion-centos/kclvm/bin/kclvm_cli
	# linux use native clang and need not chmod

	# Copy docs
	cp -r ./docs ./_build/bundles/kusion-centos/docs

	# Copy README.md
	cp ./README.md ./_build/bundles/kusion-centos

	# Copy kusion
	cp -r ./_build/bundles/kusion-linux/bin ./_build/bundles/kusion-centos

	# Build tgz
	cd ./_build/bundles/kusion-centos && tar -zcvf ../kusion-centos.tgz  .
	cd ./_build/bundles && go run ../../scripts/md5file/main.go kusion-centos.tgz > kusion-centos.tgz.md5.txt

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
		-ldflags="-s -w" \
		./cmd/kusionctl

build-local-windows-all: build-local-windows  ## Build kusion & kcl tool chain for windows
	# Install kclvm windows
	go run ./scripts/install-kclvm \
		--triple=windows \
		--mirrors=${KCLVM_URL_BASE_MIRRORS} \
		--outdir=./_build/bundles/kusion-windows/kclvm

	# Build kcl-go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ${PWD}/_build/bundles/kusion-windows/kclvm/kcl-go \
		./scripts/install-kcl-go

	# Build kcl-openapi
	cd ./scripts/install-kcl-openapi && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
		go build -o ${PWD}/_build/bundles/kusion-windows/kcl-openapi

	# Copy docs
	cp -r ./docs ./_build/bundles/kusion-windows/docs

	# Copy README.md
	cp ./README.md ./_build/bundles/kusion-windows
	# Build zip
	cd ./_build/bundles/kusion-windows && zip -r ../kusion-windows.zip .
	cd ./_build/bundles && go run ../../scripts/md5file/main.go kusion-windows.zip > kusion-windows.zip.md5.txt

build-image:  ## Build kusion image
	make build-local-linux-all
	docker build -t kusion/kusion .

sh-in-docker:  ## Run a shell in the docker container of kusion
	${RUN_IN_DOCKER} bash

e2e-test:
	# Run e2e test
	hack/run-e2e.sh

.PHONY: test cover cover-html format lint lint-fix doc build-changelog upload clean build-all build-image build-local-linux build-local-windows build-local-linux-all build-local-windows-all e2e-test
