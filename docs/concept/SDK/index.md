---
title: "K8S 资源模型 SDK"
linkTitle: "K8S 资源模型 SDK"
date: 2020-02-17
weight: 5
description: >
  为了避免用户从零开发 KCL 语言代码，Kusion 提供了与 OpenAPI Spec 表达含义完全一致的资源模型。
---

为了避免用户从零开发 KCL 语言代码，Kusion 提供了与 OpenAPI Spec 表达含义完全一致的资源模型，并在这些模型的基础之上进行了再一次的封装，使其更接近用户的直观感受。
Kusion 不仅参考 OAM 对 SDK 进行了扩展以方便用户编写代码，还额外提供了 builder、mixin 等多种编程模式，优化执行机制的体验，方便用户以任意的方式对资源进行组装。

## 资源模型完备
Kusion 支持复用 Kubernetes、Istio 等三百多个经过良好设计的核心模型，对于绝大部分需求可以通过扩展已有模型的方式实现，简单灵活，也避免了翻译类 operator 的存在。对于已有模型无法支持的情况， Kusion 也支持通过自定义模型，即对应于 Kubernetes 中的 CRD 概念。

## OCMP 实践模型
Kusion 参考开源模型规范 OAM（Open Application Model）中的 component 及 trait 概念进行了扩展，将语言及工具的玩法总结为一套实践模型 Open Cloud-Native Management Practice，旨在帮助用户通过一套简单的编程模型满足其业务需求。参考 OAM 定义，OCMP 定义了面向业务的 component 模型 server、 job、daemon 供用户使用或扩展，并提供了插件化的附件 blocks 定义，同时支持用户自定义 block。目前 Kusion 提供了 server 模型满足试点需求，后续将提供更多模型。

## 编写资源声明
kusion 面向不同需求的用户提供了三种声明编写方式：适合日常运维使用的 model 模式、适合有较强自定义需求的 builder 模式 和 mixin 模式。

