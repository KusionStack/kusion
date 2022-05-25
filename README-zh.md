## 简介

[![GitHub release](https://img.shields.io/github/release/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/KusionStack/kusion)](https://goreportcard.com/report/github.com/KusionStack/kusion)
[![license](https://img.shields.io/github/license/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/blob/main/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/KusionStack/kusion.svg)](https://pkg.go.dev/github.com/KusionStack/kusion)
[![Coverage Status](https://coveralls.io/repos/github/KusionStack/kusion/badge.svg)](https://coveralls.io/github/KusionStack/kusion)

> KusionStack 提供面向云原生生态的定义及最佳实践，提供静态类型配置语言及工具支持，在业务镜像外提供 **Compile to Cloud** 的技术栈。Kusion 使用 Golang 语言编写，并具有跨 Unix-like 平台属性。

## 📜 语言

[English](https://github.com/KusionStack/kusion/blob/main/README.md) | [简体中文](https://github.com/KusionStack/kusion/blob/main/README-zh.md)

## ✨ 功能概述

Kusion 的众多功能采用子命令的形式完成，其中较为常用的子命令包括 `apply`、`init`、`destroy`、`ls`、`preview` 等。

- `kusion apply`：接受 KCL 语言编写的代码文件作为输入，其输出可以是 Yaml 文件、JSON 文件，甚至可以直接执行到 K8S Runtime；
- `kusion init`：可以帮助用户快速新建一个 Kusion 项目；
- `kusion destroy`：可以删除由 KCL 创建的 K8S 资源；
- `kusion ls`：列出当前目录或指定目录中的所有 Project 和 Stack 信息；
- `kusion preview`：预览 Stack 中的一系列资源更改；

完整的命令功能说明，详见 [Kusion 命令](docs/cmd/en/kusion.md)。

## 🛠️ 安装

### 二进制安装（跨平台: windows, linux, mac ...）

从二进制安装，只需从 `kusion` 的 [发布页面](https://github.com/KusionStack/kusion/releases) 下载对应平台的二进制文件，然后将二进制文件放在命令行能访问到的目录中即可。

### Kusinoup

你可以通过 `kusionup` 安装多个 `kusion` 版本，默认会安装最新版。

```
brew install KusionStack/tap/kusionup
kusionup install
```

### 从源码构建

使用 Go 1.16+ 版本，你可以通过 `make` 指令从源码构建适应于不同平台的完整 `kusion` 发布包：

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

Docker 用户可以用以下命令拉取 `kusion` 的镜像：

```
docker pull kusionstack/kusion:latest
```

## ⚡ 快速开始

访问[快速开始](docs/getting-started.md)了解如何快速创建并应用一个 Kusion 项目。

## 🎖︎ 贡献指南

Kusion 仍处在初级阶段，有很多能力需要补全，所以我们欢迎所有人参与进来与我们一起共建。
访问[贡献指南](docs/contributing.md)了解如何参与到贡献 Kusion 项目中。
如有任何疑问欢迎[提交 Issue](https://github.com/KusionStack/kusion/issues)。
