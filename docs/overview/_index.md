---
title: "Kusion 简介"
linkTitle: "简介"
weight: 1
date: 2020-02-17
description: >
  Kusion 基础介绍。
---

## Kusion 项目
Kusion 提供面向云原生生态的定义及最佳实践，提供高级动态配置语言及工具支持，在业务镜像外提供 "Compile to Cloud" 的技术栈。Kusion 使用 Golang 语言编写，并具有跨 Unix-like 平台属性。

关于 Kusion 的理念、功能、实践、场景见文档 《云原生配置系统概述》。目前 Kusion 包括：KCL 语言、K8S 资源模型 SDK、资源状态白盒化框架等 3 大模块。

#### Kusion 命令思路
Kusion 的众多功能采用子命令的形式完成，其中较为重要的子命令包括`apply`、`init`、`delete`、`login`等等，以`apply`为例，`kusion apply`子命令接受 KCL 语言编写的代码文件作为输入，其输出可以是 Yaml 文件、Json 文件、甚至可以直接执行到 K8S Runtime 等等；`kusion init`子命令可以帮助用户快速新建一个 Kusion 工程项目；`kusion delete`子命令可以删除由 KCL 创建的 K8S 资源；`kusion login`控制只有相应权限的用户才能访问服务器资源；`kusion plugin`可以无缝使用 kubectl 已有的插件资源等等……

### KCL 语言

**KCL**（kusion configuration laguange）作为 Kusion 项目的支撑，实现了一门专门为云原生配置编写、逻辑编写而设计的高级动态语言。KCL 为云原生配置系统而设计，但作为一种通用的配置语言不仅限于云原生领域。KCL 吸收了声明式、OOP 编程范式的设计理念，基于此支持对大量不同问题领域配置的开放化建模描述和管理。KCL 希望使用者通过编程方式编写配置，从而更关注人的可读写性。KCL 受动态 PGL Python3 启发，可以视为 Python 的一种方言，所以 K 语言与 Python3 有很多相似之处。

### K8S 资源模型 SDK
KCL 语言作为面向用户的编程界面，对用户输入的代码进行语义上的解析和执行。但一般来说，用户不会从零开始编写 KCL 语言代码，而是利用封装好的 K8S 资源模型 SDK 来实现对资源的快速部署。

这一套模型参考开源模型规范 OAM（Open Application Model）中的 component 及 trait 概念进行了扩展，将语言及工具的玩法总结为一套实践模型 Open Cloud-Native Management Practice，旨在帮助用户通过一套简单的编程模型满足其业务需求。参考 OAM 定义，OCMP 定义了面向业务的 component 模型 server、 job、daemon 供用户使用或扩展，并提供了插件化的附件 blocks 定义，同时支持用户自定义 block。

我们提供了一套完整的 K8S 原生资源的模型，用户借由这些编写好的 KCL 语言模型，可以直接创建想要的 K8S 模型，免去了手工编写 K8S 资源的麻烦。除了 K8S 原生资源模型之外，我们还逐步加入了对 Istio、Argo 等扩展资源模型的支持。不久之后，我们还将支持用户自定义资源模型。

> [K8S 资源模型示例](examples/bookinfo/base/servers.k)

### 资源状态白盒化框架
为了让资源部署这一原本不透明的状态过程变得可视化，Kusion 项目又引入了资源状态白盒化框架。资源状态白盒化框架可以将原本对用户不可见的 K8S 资源状态信息清晰地展示在用户面前，在资源部署的过程中，不需要频繁地刷新命令获取当前资源创建情况，而是动态地获知相关信息。即使 KCL 代码编写有误，资源状态白盒化框架也可以快速地定位问题、展示问题，方便快速调试。

有关 Kusion 的更多信息，请访问 [Kusion Project](http://kusionstack.io/kusion) 页面和 [Quick Startup](docs/quick-start/_index.md) 文档。

## 快速开始

请参考[快速开始](../quick-start)。
