<div align="center">
<p></p><p></p>
<p>
    <img  src="docs/logo.png">
</p>

<h1 style="font-size: 1.5em;">
    Intent-Driven Platform Orchestrator
</h1>

<p align="center">
  <a href="https://www.kusionstack.io/docs/" target="_blank"><b>🌐 官网</b></a> •
  <a href="https://www.kusionstack.io/docs/getting-started/getting-started-with-kusion-cli/deliver-quickstart" target="_blank"><b>⚡️ 快速开始</b></a> •
  <a href="https://www.kusionstack.io/docs/" target="_blank"><b>📚 文档</b></a> •
  <a href="https://github.com/orgs/KusionStack/discussions" target="_blank"><b>💬 讨论</b></a><br>
  <a href="https://github.com/KusionStack/kusion/blob/main/README.md" target="_blank">[English]</a>
  [中文]
</p>

[![Kusion](https://github.com/KusionStack/kusion/actions/workflows/release.yaml/badge.svg)](https://github.com/KusionStack/kusion/actions/workflows/release.yaml)
[![GitHub release](https://img.shields.io/github/release/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/KusionStack/kusion)](https://goreportcard.com/report/github.com/KusionStack/kusion)
[![Go Reference](https://pkg.go.dev/badge/github.com/KusionStack/kusion.svg)](https://pkg.go.dev/github.com/KusionStack/kusion)
[![license](https://img.shields.io/github/license/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/blob/main/LICENSE)

</div>

Kusion 是一个意图驱动的[平台编排器](https://internaldeveloperplatform.org/platform-orchestrators/)，它位于[内部开发者平台 (IDP)](https://internaldeveloperplatform.org/what-is-an-internal-developer-platform/)的核心。通过 Kusion，你可以启用以应用为中心的开发，你的开发者只需要编写单一的应用配置 - [AppConfiguration](https://www.kusionstack.io/docs/concepts/appconfigurations)，无需提供特定于环境的值，即可定义工作负载和所有资源依赖，Kusion 确保为应用运行提供一切所需。

Kusion 帮助负责创建应用的应用开发者以及负责维护应用运行的基础设施的平台工程师。这些角色在你的组织中可能重叠或不同，但 Kusion 旨在为任何负责这些任务的从业者减轻工作负担。

## Kusion 如何工作？

作为一个平台编排器，Kusion 使您能够解决通常与 Day 0 和 Day 1 关联的挑战。平台工程师和应用工程师都可以从 Kusion 中获益。

Kusion 有两个关键工作流程：
1. **Day 0 - 设置模块和工作空间：** 平台工程师为部署应用及其底层基础设施创建共享模块，并为目标着陆区定义工作空间。这些标准化的共享模块编写了包括安全、合规和财务在内的组织中各利益相关者的要求。
   Kusion 模块抽象了底层基础设施工具的复杂性，使应用开发者能够使用自助模式部署他们的应用程序。
   
2. **Day 1 - 设置应用程序：** 应用开发者利用平台工程师创建的工作空间和模块来部署应用及其支持的基础设施。平台团队维护工作空间和模块，这允许应用开发者专注于在标准化的基础设施上使用可重复的过程构建应用。

## 快速开始

本指南将涵盖：
1. 安装 Kusion CLI。
2. 使用 Kusion 将应用部署到 Kubernetes。

### 安装

#### Homebrew (macOS & Linux)

```shell
# tap formula repository Kusionstack/tap
brew tap KusionStack/tap

# install Kusion 
brew install KusionStack/tap/kusion
```

#### Powershell

```shell
# install Kusion latest version
powershell -Command "iwr -useb https://www.kusionstack.io/scripts/install.ps1 | iex"
```

> 有关安装的更多信息，请参考[安装指南](https://www.kusionstack.io/docs/getting-started/getting-started-with-kusion-cli/install-kusion)以获取更多选项。

### 部署

要部署应用程序，您可以运行 `kusion apply` 命令。
> 要快速启动并运行 Kusion，请参阅[快速开始指南](https://www.kusionstack.io/docs/getting-started/getting-started-with-kusion-cli/deliver-quickstart)。
>

![apply](https://raw.githubusercontent.com/KusionStack/kusionstack.io/main/static/img/docs/user_docs/getting-started/kusion_apply_quickstart.gif)


## 联系方式

如果您有任何问题，欢迎通过以下方式联系我们：
- [Slack](https://kusionstack.slack.com) | [加入](https://join.slack.com/t/kusionstack/shared_invite/zt-2drafxksz-VzCZZwlraHP4xpPeh_g8lg)
- [钉钉群](https://page.dingtalk.com/wow/dingtalk/act/en-home)：`42753001`（中文）
- 微信群（中文）：添加微信小助手，拉你进用户群

  <img src="docs/wx_spark.jpg" width="200" height="200"/>

## 贡献

如果您有兴趣贡献，在**提交 Pull Request 前**请参阅[贡献指南](CONTRIBUTING.md)。

## 许可证

Kusion 根据 Apache 2.0 许可证授权。有关详细信息，请见 [LICENSE](LICENSE) 文件。
