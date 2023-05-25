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

Kusion is the platform engineering engine of [KusionStack](https://github.com/KusionStack). It delivers intentions
described in [Konfig](https://github.com/KusionStack/konfig) to Kubernetes, Clouds and Customized Infrastructure
resources

## Key Features

- **Application-Centric**: Managing all application operations in one place, in a unified way
- **Enable Self-Service**: Developers fulfill their own needs by using the capabilities provided by the platform
- **Shift left security**: Guarantee security at the earliest stages to make operation more confidence
- **Kubernetes-friendly**: Provide friendly observability and troubleshooting capabilities to make Kubernetes operations
  easier.


<div align="center">

![arch](docs/arch.png)
</div>

## Quick Start

### Deploy your first App

Visit [Quick Start](https://kusionstack.io/docs/user_docs/getting-started/usecases/deliver-first-project) to deliver an
App with one Kusion command

### Demo Video

[![Wordpress Demo](http://img.youtube.com/vi/QHzKKsoKLQ0/maxresdefault.jpg)](http://www.youtube.com/watch?v=QHzKKsoKLQ0)

## Installation

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

```
docker pull kusionstack/kusion:latest
```

> For more information about installation, please check the [Installation Guide](https://kusionstack.io/docs/user_docs/getting-started/install) on KusionStack official website


# Contact Us
- Twitter: [KusionStack](https://twitter.com/KusionStack)
- Slack: [Kusionstack](https://join.slack.com/t/kusionstack/shared_invite/zt-19lqcc3a9-_kTNwagaT5qwBE~my5Lnxg)
- DingTalk (Chinese): 42753001
- Wechat Group (Chinese)

  <img src="docs/wx_spark.jpg" width="200" height="200"/>


# 🎖︎ Contribution Guide

Kusion is still in the initial stage, and there are many capabilities that need to be made up, so we welcome everyone to participate in construction with us. Visit the [Contribution Guide](docs/contributing.md) to understand how to participate in the contribution KusionStack project. If you have any questions, please [Submit the Issue](https://github.com/KusionStack/kusion/issues).
