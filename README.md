## Introduction

[![GitHub release](https://img.shields.io/github/release/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/KusionStack/kusion)](https://goreportcard.com/report/github.com/KusionStack/kusion)
[![license](https://img.shields.io/github/license/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/blob/main/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/KusionStack/kusion.svg)](https://pkg.go.dev/github.com/KusionStack/kusion)
[![Coverage Status](https://coveralls.io/repos/github/KusionStack/kusion/badge.svg)](https://coveralls.io/github/KusionStack/kusion)

> KusionStack provides the definition and best practice for the native ecology of the cloud, provides the statically typed language and tool support, and provides **Compile to Cloud** technology stack outside the business mirror. Kusion is written in Golang and has attributes of crossing Unix-Like platform.

## 📜 Language

[English](https://github.com/KusionStack/kusion/blob/main/README.md) | [简体中文](https://github.com/KusionStack/kusion/blob/main/README-zh.md)

## ✨ Functional Overview

Kusion's many functions are completed in the form of subcommands. Among them, the more commonly used subcommands include `apply`,`init`, `destroy`,`ls`, `preview`, etc.

- `kusion apply`: Accept the code file written by the KCL language as the input. The output can be YAML files, JSON files, or even execute directly to the K8S Runtime
- `kusion init`: Initialize KCL file structure and base codes for a new project
- `kusion destroy`: Destroy a configuration stack to resource(s) by work directory
- `kusion ls`: List all project and stack information
- `kusion preview`: Preview a series of resource changes within the stack

For a complete command function description, please refer to the [Kusion Command](docs/cmd/en/kusion.md)。

## 🛠️ Installation

### Binary (Cross-platform: windows, linux, mac ...)

To get the binary just download the latest release for your OS/Arch from the [release page](https://github.com/KusionStack/kusion/releases) and put the binary somewhere convenient.

### Kusionup

You can install multiple versions of `kusion` with `kusionup`, and the latest version is installed by default.

```
# First, install kusionup, multiple ways:
# - brew install KusionStack/tap/kusionup
# - curl -sSf https://raw.githubusercontent.com/KusionStack/kusionup/main/scripts/install.sh | bash
# - For other ways, please refer to: https://github.com/KusionStack/kusionup#%EF%B8%8F-installation

# Second, install kusion by kusionup
kusionup install
```

### Build from Source

Starting with Go 1.16+, you can use the `make` command to build a complete `kusion` distribution package suitable for different platforms from the source code.

```
# Build all platforms (darwin, linux, windows)
make build-all

# Build kusion & kcl tool chain for macOS
# make build-local-darwin-all
# Build kusion & kcl tool chain for linux
# make build-local-linux-all
# Build kusion & kcl tool chain for windows
# make build-local-windows-all
```

### Docker

Docker users can use the following commands to pull the latest image of the `kusion`:

```
docker pull kusionstack/kusion:latest
```

## ⚡ Quick Start

Visit [Quick Start](docs/getting-started.md) to understand how to quickly create and apply a KusionStack project.

## 🎖︎ Contribution Guide

Kusion is still in the initial stage, and there are many capabilities that need to be made up, so we welcome everyone to participate in construction with us. Visit the [Contribution Guide](docs/contributing.md) to understand how to participate in the contribution KusionStack project. If you have any questions, please [Submit the Issue](https://github.com/KusionStack/kusion/issues).
