<div align="center">
<p></p><p></p>
<p>
    <img  src="docs/logo.png">
</p>
<h1>Build your Internal Developer Platform more effectively and safely</h1>

[简体中文](https://github.com/KusionStack/kusion/blob/main/README-zh.md)
| [English](https://github.com/KusionStack/kusion/blob/main/README.md)

[Konfig](https://github.com/KusionStack/konfig) | [KCLVM](https://github.com/KusionStack/KCLVM)
| [Kusion](https://github.com/KusionStack/kusion) | [kusionstack.io](https://kusionstack.io/)
| [CNCF Landscape](https://landscape.cncf.io/?selected=kusion-stack)

[![Kusion](https://github.com/KusionStack/kusion/actions/workflows/release.yaml/badge.svg)](https://github.com/KusionStack/kusion/actions/workflows/release.yaml)
[![GitHub release](https://img.shields.io/github/release/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/KusionStack/kusion)](https://goreportcard.com/report/github.com/KusionStack/kusion)
[![Coverage Status](https://coveralls.io/repos/github/KusionStack/kusion/badge.svg)](https://coveralls.io/github/KusionStack/kusion)
[![Go Reference](https://pkg.go.dev/badge/github.com/KusionStack/kusion.svg)](https://pkg.go.dev/github.com/KusionStack/kusion)
[![license](https://img.shields.io/github/license/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/blob/main/LICENSE)
</div>

# Kusion

Kusion is the cloud-native platform engineering engine of [KusionStack](https://github.com/KusionStack), designed to help you build your efficient and secure Internal Developer Platform (IDP). It allows application developers to easily and consistently perform all operational tasks throughout the DevOps lifecycle in one place, across different infrastructures including Kubernetes, cloud platforms, and on-premises environments

## Key Features

- **Platform as Code**: Managing all application operations in one place, in a unified, easy way.
- **Dynamic Configuration Management**: Developers deploy to any environment using one environment-agnostic configuration, while Kusion handles the underlying complexity.
- **Enable Self-Service**: Developers fulfill their own needs by choosing the capabilities provided by the platform engineers in [Kusion Modules](https://www.kusionstack.io/docs/next/kusion/concepts/kusion-module).
- **Built-in Security and Compliance**: Platform engineers standardize application deliveries with security, compliance, and best practices by flexible platform configurations.
- **Open Module Ecosystem**: A growing open module ecosystem covers all stages of the DevOps lifecycle, with extensive connections to various cloud-native infrastructures.
- **Lightweight and Kubernetes-friendly**: As a pure client-side solution with rich APIs, Kusion offers excellent portability and ease of integration. Meanwhile, user-friendly observability and troubleshooting capabilities make Kubernetes operations easier.

<div align="center">

![workflow](docs/workflow.png)
</div>

## Quick Start

### Deploy your first App

Visit [Quick Start](https://www.kusionstack.io/docs/kusion/getting-started/deliver-wordpress) to deliver an
App with one Kusion command

![apply](https://raw.githubusercontent.com/KusionStack/kusionstack.io/main/static/img/docs/user_docs/getting-started/apply.gif)

## Installation

### Homebrew (macOS & Linux)

```shell
brew install KusionStack/tap/kusion
```

### Go Install

```shell
go install kusionstack.io/kusion@latest
```

### Docker

```
docker pull kusionstack/kusion:latest
```

> For more information about installation, please check the [Installation Guide](https://www.kusionstack.io/docs/kusion/getting-started/install-kusion) on KusionStack official website


# Contact Us
- Twitter: [KusionStack](https://twitter.com/KusionStack)
- Slack: [Kusionstack](https://kusionstack.slack.com)
- DingTalk (Chinese): 42753001
- Wechat Group (Chinese)

  <img src="docs/wx_spark.jpg" width="200" height="200"/>


# 🎖︎ Contribution Guide

Kusion is still in the rapid development stage with many capabilities to be enhanced. We welcome everyone to join us and contribute to the collaborative growth.

Visit the [Contribution Guide](docs/contributing.md) to understand how to participate in the contribution KusionStack project. If you have any questions, please [Submit the Issue](https://github.com/KusionStack/kusion/issues).
