<div align="center">
<p></p><p></p>
<p>
    <img  src="docs/logo.png">
</p>
<h1>更高效更安全地构建你的内部研发平台</h1>

[简体中文](https://github.com/KusionStack/kusion/blob/main/README-zh.md) | [English](https://github.com/KusionStack/kusion/blob/main/README.md)

[Konfig](https://github.com/KusionStack/konfig) | [KCLVM](https://github.com/KusionStack/KCLVM) | [Kusion](https://github.com/KusionStack/kusion) | [kusionstack.io](https://kusionstack.io/) | [CNCF Landscape](https://landscape.cncf.io/?selected=kusion-stack)

[![Kusion](https://github.com/KusionStack/kusion/actions/workflows/release.yaml/badge.svg)](https://github.com/KusionStack/kusion/actions/workflows/release.yaml)
[![GitHub release](https://img.shields.io/github/release/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/KusionStack/kusion)](https://goreportcard.com/report/github.com/KusionStack/kusion)
[![Coverage Status](https://coveralls.io/repos/github/KusionStack/kusion/badge.svg)](https://coveralls.io/github/KusionStack/kusion)
[![Go Reference](https://pkg.go.dev/badge/github.com/KusionStack/kusion.svg)](https://pkg.go.dev/github.com/KusionStack/kusion)
[![license](https://img.shields.io/github/license/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/blob/main/LICENSE)
</div>

# Kusion

Kusion 是 [KusionStack](https://github.com/KusionStack)
的平台工程引擎，可以根据用户在 [Konfig](https:/github.com/KusionStack/konfig) 中描述的运维意图对
Kubernetes、IaaS云资源和自定义基础设施运维

## 核心能力

- **以应用为中心**: 在同一个地方、以统一的方式，管理应用所需要的所有的运维操作
- **研发自服务**: 研发可以利用平台提供的能力，自助实现自己的需求
- **风险左移**: 在运维前期保证安全性，使操作更有信心。
- **Kubernetes 友好**: 提供友好的可观测性与排障能力，让 Kubernetes 运维更简单


<div align="center">

![arch](docs/arch.png)
</div>

## 快速开始

### 部署第一个应用

参考 [Quick Start](https://kusionstack.io/docs/user_docs/getting-started/usecases/deliver-first-project) 通过 Kusion
部署您的第一个应用

### 演示视频

[![Wordpress Demo](http://img.youtube.com/vi/QHzKKsoKLQ0/maxresdefault.jpg)](http://www.youtube.com/watch?v=QHzKKsoKLQ0)

## 安装

### Homebrew (macOS & Linux)

```shell
brew install KusionStack/tap/kusion
```

### Scoop (Windows)

```bash
scoop add bucket KusionStack https://github.com/KusionStack/scoop-bucket.git
scoop install KusionStack/kusion
```

### Go Install

```shell
go install kusionstack.io/kusion@latest
```

### Docker

```shell
docker pull kusionstack/kusion:latest
```

> 有关安装的更多信息，请查看 KusionStack 官网的[安装指南](https://kusionstack.io/zh-CN/docs/user_docs/getting-started/install)。

# 🎖︎ 贡献指南

Kusion 仍处在初级阶段，有很多能力需要补全，所以我们欢迎所有人参与进来与我们一起共建。
访问[贡献指南](docs/contributing.md)了解如何参与到贡献 Kusion 项目中。
如有任何疑问欢迎[提交 Issue](https://github.com/KusionStack/kusion/issues)。
