<div align="center">
<p></p><p></p>
<p>
    <img  src="docs/logo.png">
</p>
<h1>面向异构资源的自服务应用发布平台</h1>

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

Kusion 是 [KusionStack](https://github.com/KusionStack) 的平台工程引擎，可以根据用户在 [Konfig](https://github.com/KusionStack/konfig) 中描述的运维意图对 Kubernetes、IaaS 云资源和自定义基础设施运维
## 核心能力

- **应用全生命周期管理**: 结合 [Kusion](https://github.com/KusionStack/kusion) 与 [Konfig](https://github.com/KusionStack/konfig) 实现从应用第一行配置代码到生产可用的全生命周期管理
- **自服务能力**: 为 App Dev 提供自服务能力，帮助他们与 SRE 和 Platform Dev 在代码化的平台上高效的合作
- **异构资源运维**: 以统一的方式运维 Kubernetes、Terraform 和自定义基础设施等多种运行时的资源
- **厂商无关**: 一次编写，动态渲染，多云运行

<div align="center">

![arch](docs/arch.png)
</div>

## 快速开始

参考 [快速开始](https://kusionstack.io/docs/user_docs/getting-started/usecase) 通过一条 Kusion 命令拉起一个应用

![apply](https://kusionstack.io/assets/images/apply-1cc90f7fe294b3b1414b4dd3a27a2d2b.gif)



## 安装

### Homebrew (macOS & Linux)

```shell
brew install KusionStack/tap/kusion
```

### Go Install

```shell
go install kusionstack.io/kusion@latest
```

### Docker

```shell
docker pull kusionstack/kusion:latest
```

> 有关安装的更多信息，请查看 KusionStack 官网的[安装指南](https://kusionstack.io/docs/user_docs/getting-started/install)。

## 部署第一个应用

一键部署你的一个应用，详情请参考 [use case](https://kusionstack.io/docs/user_docs/getting-started/usecase)

# 🎖︎ 贡献指南

Kusion 仍处在初级阶段，有很多能力需要补全，所以我们欢迎所有人参与进来与我们一起共建。
访问[贡献指南](docs/contributing.md)了解如何参与到贡献 Kusion 项目中。
如有任何疑问欢迎[提交 Issue](https://github.com/KusionStack/kusion/issues)。
