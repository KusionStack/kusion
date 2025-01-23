<div align="center">
<p></p><p></p>
<p>
    <img  src="docs/logo.png">
</p>

<h1 style="font-size: 1.5em;">
    Intent-Driven Platform Orchestrator
</h1>

<p align="center">
  <a href="https://www.kusionstack.io/docs/" target="_blank"><b>🌐 Website</b></a> •
  <a href="https://www.kusionstack.io/docs/getting-started/getting-started-with-kusion-cli/deliver-quickstart" target="_blank"><b>⚡️ Quick Start</b></a> •
  <a href="https://www.kusionstack.io/docs/" target="_blank"><b>📚 Docs</b></a> •
  <a href="https://kusion.kusionstack.io/" target="_blank"><b>📚 Landing Page</b></a> •
  <a href="https://github.com/orgs/KusionStack/discussions" target="_blank"><b>💬 Discussions</b></a><br>
  [English] 
  <a href="https://github.com/KusionStack/kusion/blob/main/README-zh.md" target="_blank">[中文]</a>
</p>

[![Kusion](https://github.com/KusionStack/kusion/actions/workflows/release.yaml/badge.svg)](https://github.com/KusionStack/kusion/actions/workflows/release.yaml)
[![GitHub release](https://img.shields.io/github/release/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/KusionStack/kusion)](https://goreportcard.com/report/github.com/KusionStack/kusion)
[![Go Reference](https://pkg.go.dev/badge/github.com/KusionStack/kusion.svg)](https://pkg.go.dev/github.com/KusionStack/kusion)
[![license](https://img.shields.io/github/license/KusionStack/kusion.svg)](https://github.com/KusionStack/kusion/blob/main/LICENSE)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/kusion)](https://artifacthub.io/packages/helm/kusionstack/kusion)
[![CNCF](https://shields.io/badge/CNCF-Sandbox%20project-blue?logo=linux-foundation&style=flat)](https://landscape.cncf.io/?item=provisioning--automation-configuration--kusionstack)
[![Gitpod Ready-to-Code](https://img.shields.io/badge/Gitpod-Ready--to--Code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/KusionStack/kusion)
[![Twitter Follow](https://img.shields.io/twitter/follow/KusionStack?style=social)](https://twitter.com/KusionStack)
[![Medium](https://img.shields.io/badge/@kusionstack-black?style=flat&logo=medium&logoColor=white&link=https://medium.com/@kusionstack)](https://medium.com/@kusionstack)
[![Slack](https://img.shields.io/badge/slack-kusion-blueviolet?logo=slack)](https://cloud-native.slack.com/archives/C07U0395UG0)


<a href="https://www.producthunt.com/posts/kusion?embed=true&utm_source=badge-featured&utm_medium=badge&utm_souce=badge-kusion" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=800331&theme=dark" alt="Kusion - Application&#0032;delivery&#0032;made&#0032;simple | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>

</div>

## What is Kusion?

Kusion is an intent-driven [Platform Orchestrator](https://internaldeveloperplatform.org/platform-orchestrators/), which sits at the core of an [Internal Developer Platform (IDP)](https://internaldeveloperplatform.org/what-is-an-internal-developer-platform/). With Kusion you can enable app-centric development, your developers only need to write a single application specification - [AppConfiguration](https://www.kusionstack.io/docs/concepts/appconfigurations). AppConfiguration defines the workload and all resource dependencies without needing to supply environment-specific values, Kusion ensures it provides everything needed for the application to run.

Kusion helps app developers who are responsible for creating applications and the platform engineers responsible for maintaining the infrastructure the applications run on. These roles may overlap or align differently in your organization, but Kusion is intended to ease the workload for any practitioner responsible for either set of tasks.

<div align="center">

![workflow](docs/overview.jpg)
</div>

## How does Kusion work?

As a Platform Orchestrator, Kusion enables you to address challenges often associated with Day 0 and Day 1. Both platform engineers and application engineers can benefit from Kusion.

There are two key workflows for Kusion:

1. **Day 0 - Set up the modules and workspaces:** Platform engineers create shared modules for deploying applications and their underlying infrastructure, and workspace definitions for the target landing zone. These standardized, shared modules codify the requirements of stakeholders across the organization including security, compliance, and finance.

	Kusion modules abstract the complexity of underlying infrastructure tooling, enabling app developers to deploy their applications using a self-service model.
	
	<div align="center">

	![workflow](docs/platform_workflow.jpg)
	</div>
	
2. **Day 1 - Set up the application:** Application developers leverage the workspaces and modules created by the platform engineers to deploy applications and their supporting infrastructure. The platform team maintains the workspaces and modules, which allows application developers to focus on building applications using a repeatable process on standardized infrastructure.

	<div align="center">

	![workflow](docs/app_workflow.jpg)
	</div>


## Introducing Kusion Server with a Developer Portal

Starting With Kusion v0.14.0, we are officially introducing Kusion Server with a Developer Portal.

Kusion Server runs as a long-running service, providing the same set of functionalities as the Kusion CLI, with additional capabilities to manage application metadata and visualized application resource graphs. 

Kusion Server manages instances of Projects, Stacks, Workspaces, Runs, etc. centrally via a Developer Portal and a set of RESTful APIs for other systems to integrate with.

## Quick Start With Kusion Server with a Developer Portal

To start with a Kusion Server, please follow the [QuickStart Guide for Kusion Server](https://www.kusionstack.io/docs/getting-started/getting-started-with-kusion-server/install-kusion).

## Quick Start With Kusion CLI

This guide will cover:

1. Install Kusion CLI.
2. Deploy an application to Kubernetes with Kusion.

### Install

#### Homebrew (macOS & Linux)

```shell
# tap formula repository Kusionstack/tap
brew tap KusionStack/tap

# install Kusion 
brew install KusionStack/tap/kusion
```

#### Powershell

```
# install Kusion latest version
powershell -Command "iwr -useb https://www.kusionstack.io/scripts/install.ps1 | iex"
```

> For more information about CLI installation, please refer to the [CLI Installation Guide](https://www.kusionstack.io/docs/getting-started/getting-started-with-kusion-cli/install-kusion) for more options.

### Deploy

To deploy an application, you can run the `kusion apply` command.

> To rapidly get Kusion up and running, please refer to the [Quick Start Guide](https://www.kusionstack.io/docs/getting-started/getting-started-with-kusion-cli/deliver-quickstart).

![apply](https://raw.githubusercontent.com/KusionStack/kusionstack.io/main/static/img/docs/user_docs/getting-started/kusion_apply_quickstart.gif)

## Case Studies

Check out these case studies on how Kusion can be useful in production:

- Jan 2025: [Configuration Management at Ant Group: Generated Manifest & Immutable Desired State](https://blog.kusionstack.io/configuration-management-at-ant-group-generated-manifest-immutable-desired-state-3c50e363a3fb)
- Jan 2024: [Modeling application delivery using AppConfiguration](https://blog.kusionstack.io/modeling-application-delivery-using-appconfiguration-d291830de8f1)

## Contact

If you have any questions, feel free to reach out to us in the following ways:

- [Slack](https://kusionstack.slack.com) | [Join](https://join.slack.com/t/kusionstack/shared_invite/zt-2drafxksz-VzCZZwlraHP4xpPeh_g8lg)
- [DingTalk Group](https://page.dingtalk.com/wow/dingtalk/act/en-home): `42753001`  (Chinese)
- WeChat Group (Chinese): Add the WeChat assistant to bring you into the user group.

  <img src="docs/wx_spark.jpg" width="200" height="200"/>

## Contributing

If you're interested in contributing, please refer to the [Contributing Guide](./CONTRIBUTING.md) **before submitting 
a pull request**.

## License

Kusion is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.

## OpenSSF Best Practice Badge
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9586/badge)](https://www.bestpractices.dev/projects/9586)
