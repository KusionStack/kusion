<div align="center">
<p></p><p></p>
<p>
    <img  src="docs/logo.png">
</p>
<h1>Codify, Collaborate, Automate modern App delivery across Kubernetes and Clouds</h1>

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

Kusion is the engine of [KusionStack](https://github.com/KusionStack) for parsing user's intentions described in [Konfig](https://github.com/KusionStack/konfig) and making them effect in infrastructures.

## Key Features

- **App Whole Lifecycle Management**: Manage App from the first code to production-ready with [Kusion](https://github.com/KusionStack/kusion) and [Konfig](https://github.com/KusionStack/konfig)
- **Team cooperation**: App Dev, SRE and Platform Dev cooperate efficiently in this codify platform 
- **Hybrid Runtime**: Orchestrate hybrid runtime resources like Terraform and Kubernetes in a unified way
- **Vendor Agnostic**: Write once, render dynamically, deliver to any cloud

<div align="center">

![arch](docs/arch.png)
</div>

## Quick Start
Visit [Quick Start](https://kusionstack.io/docs/user_docs/getting-started/usecase) to deliver an App with one Kusion command

![apply](https://kusionstack.io/assets/images/apply-30acfe738fbda046d76f2996b2bf51b5.gif)


## Installation

### Kusionup

You can install multiple versions of `kusion` with `kusionup`, and the latest version is installed by default.
#### Install Kusionup
```
# Homebrew
brew install KusionStack/tap/kusionup
```
```
# cURL
curl -sSf https://raw.githubusercontent.com/KusionStack/kusionup/main/scripts/install.sh | bash
```
#### Install Kusion
```
# visit https://github.com/KusionStack/kusionup for more details
kusionup install
```

### Binary

Download the latest release for your OS/Arch from the [release page](https://github.com/KusionStack/kusion/releases) and put the binary somewhere convenient.

### Docker

Docker users can use the following commands to pull the latest image of the `kusion`:

```
docker pull kusionstack/kusion:latest
```

## Deploy your first App
Deploy your first App with one Kusion command. Please visit this [use case](https://kusionstack.io/docs/user_docs/getting-started/usecase) for more details

# Contact Us
- Twitter: [KusionStack](https://twitter.com/KusionStack)
- Slack: [Kusionstack](https://join.slack.com/t/kusionstack/shared_invite/zt-19lqcc3a9-_kTNwagaT5qwBE~my5Lnxg)
- DingTalk (Chinese): 42753001
- Wechat Group (Chinese)

  <img src="docs/wx_spark.jpg" width="200" height="200"/>


# 🎖︎ Contribution Guide

Kusion is still in the initial stage, and there are many capabilities that need to be made up, so we welcome everyone to participate in construction with us. Visit the [Contribution Guide](docs/contributing.md) to understand how to participate in the contribution KusionStack project. If you have any questions, please [Submit the Issue](https://github.com/KusionStack/kusion/issues).
