# Kusion

![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

TODO: logo...

## 概述

Kusion 提供面向云原生生态的定义及最佳实践，提供高级动态配置语言及工具支持，
在业务镜像外提供 **Compile to Cloud** 的技术栈。Kusion 使用 Golang 语言编写，
并具有跨 Unix-like 平台属性。

## 设计理念

Kusion 项目主要由 3 大模块构成：
- [KCL 语言](docs/kcl.md)
- [K8S 资源模型 SDK](docs/k8s-model-sdk.md)
- [资源状态白盒化框架](docs/white-box.md)

KCL 语言作为面向用户的编程界面，对用户输入的代码进行语义上的解析和执行。
但一般来说，用户不会从零开始编写 KCL 语言代码，而是利用封装好的 K8S 资源模型 SDK 来实现对资源的快速部署。
为了让资源部署这一原本不透明的状态过程变得可视化，Kusion 项目又引入了资源状态白盒化框架。

## 功能概述

Kusion 的众多功能采用子命令的形式完成，其中较为常用的子命令包括 `apply`、`init`、`destroy`、`ls`、`preview` 等。

- `kusion apply`：接受 KCL 语言编写的代码文件作为输入，其输出可以是 Yaml 文件、Json 文件，甚至可以直接执行到 K8S Runtime；
- `kusion init`：可以帮助用户快速新建一个 Kusion 项目；
- `kusion destroy`：可以删除由 KCL 创建的 K8S 资源；
- `kusion ls`：列出当前目录或指定目录中的所有 Project 和 Stack 信息；
- `kusion preview`：预览 Stack 中的一系列资源更改；

完整的命令功能说明，详见[Kusion 命令](docs/cmd/en/kusion.md)。

## 快速开始

访问[快速开始](docs/getting-started.md)了解如何快速创建并应用一个 Kusion 项目。

## 贡献指南

Kusion 仍处在初级阶段，有很多能力需要补全，所以我们欢迎所有人参与进来与我们一起共建。
访问[贡献指南](docs/contributing.md)了解如何参与到贡献 Kusion 项目中。
如有任何疑问欢迎[提交 Issue](https://github.com/KusionStack/kusion/issues)。

## License

Apache License Version 2.0
