# Kusion

## Kusion 项目

Kusion 提供面向云原生生态的定义及最佳实践，提供高级动态配置语言及工具支持，在业务镜像外提供 "Compile to Cloud" 的技术栈。Kusion 使用 Golang 语言编写，并具有跨 Unix-like 平台属性。

### Kusion 命令思路

Kusion 的众多功能采用子命令的形式完成，其中较为重要的子命令包括`run`、`init`、`delete`、`login`和`plugin`等等，以`run`为例，`kusion run`子命令接受 KCL 语言编写的代码文件作为输入，其输出可以是 Yaml 文件、Json 文件、甚至可以直接执行到 K8S Runtime 等等；`kusion init`子命令可以帮助用户快速新建一个 Kusion 工程项目；`kusion delete`子命令可以删除由 KCL 创建的 K8S 资源；`kusion login`控制只有相应权限的用户才能访问服务器资源；`kusion plugin`可以无缝使用 kubectl 已有的插件资源等等……

> [Kusion 命令](docs/cmd/en/kusion.md)

### KCL 语言

**KCL**（Kusion Configuration Laguange）作为 Kusion 项目的支撑，实现了一门专门为云原生配置编写、逻辑编写而设计的高级动态语言。KCL 为云原生配置系统而设计，但作为一种通用的配置语言不仅限于云原生领域。KCL 吸收了声明式、OOP 编程范式的设计理念，基于此支持对大量不同问题领域配置的开放化建模描述和管理。KCL 希望使用者通过编程方式编写配置，从而更关注人的可读写性。KCL 受动态 PGL Python3 启发，可以视为 Python 的一种方言，所以 KCL 与 Python3 有很多相似之处。

### K8S 资源模型 SDK

KCL 语言作为面向用户的编程界面，对用户输入的代码进行语义上的解析和执行。但一般来说，用户不会从零开始编写 KCL 语言代码，而是利用封装好的 K8S 资源模型 SDK 来实现对资源的快速部署。

这一套模型参考开源模型规范 OAM（Open Application Model）中的 component 及 trait 概念进行了扩展，将语言及工具的玩法总结为一套实践模型 Open Cloud-Native Management Practice，旨在帮助用户通过一套简单的编程模型满足其业务需求。参考 OAM 定义，OCMP 定义了面向业务的 component 模型 server、job、daemon 供用户使用或扩展，并提供了插件化的附件 blocks 定义，同时支持用户自定义 block。

我们提供了一套完整的 K8S 原生资源的模型，用户借由这些编写好的 KCL 语言模型，可以直接创建想要的 K8S 模型，免去了手工编写 K8S 资源的麻烦。除了 K8S 原生资源模型之外，我们还逐步加入了对 Istio、Argo 等扩展资源模型的支持。不久之后，我们还将支持用户自定义资源模型。

> [K8S 资源模型示例](examples/bookinfo/base/servers.k)

### 资源状态白盒化框架

为了让资源部署这一原本不透明的状态过程变得可视化，Kusion 项目又引入了资源状态白盒化框架。资源状态白盒化框架可以将原本对用户不可见的 K8S 资源状态信息清晰地展示在用户面前，在资源部署的过程中，不需要频繁地刷新命令获取当前资源创建情况，而是动态地获知相关信息。即使 KCL 代码编写有误，资源状态白盒化框架也可以快速地定位问题、展示问题，方便快速调试。

## 功能列表

+ 代码编写
    + 提供动态声明式配置语言 KCL，用于云原生资源配置编写
    + 提供了设施齐全的 SDK
        + 提供面向云原生微服务类业务的顶层模型抽象和扩展机制
        + 提供直接编写 Kubernetes 模型的便利
+ 文件导出
    + 支持从 KCL 语言文件导出 Yaml、Json 文件
+ K 文件执行
    + 支持 KCL 语言执行模型到 Kubernetes 运行时
+ 状态可视化
    + 支持 Kubernetes 模型状态可视化
        + 支持 Deployment、Service、Pod 的状态追踪
        + 模型生效阶段实时配置状态变化显示
        + 支持执行结果的汇总输出
+ 工具支持
    + 支持一键初始化项目框架模板
    + 支持 Yaml 到 KCL 代码转换
    + 无缝支持 Kubectl 的插件机制

## 快速开始

访问[快速开始文档](docs/quick-start/_index.md) 了解如何快速创建一个 Kusion 项目。

## 项目示例

提供了一个使用 KCL 语言描述的 Kubernetes 的[最佳实践范例](docs/samples/_index.md)。


## 本地构建
Golang 环境配置
// todo

## 想要贡献

Kusion 仍处在初级阶段，有很多能力需要补全，所以我们欢迎所有人参与进来与我们一起共建。
访问该链接了解如何参与到贡献 Kusion 项目中。
如有任何疑问欢迎[提交 Issue](http://kusionstack.io/kusion)。

### 格式化代码
提交前请先格式化代码，格式化工具我们采用更严格的 [gofumpt](https://github.com/mvdan/gofumpt)，它是 goimports 和 gofmt 的超集；

`gofumpt` 安装:
```bash
go install mvdan.cc/gofumpt@latest
```

或者你可以将 `gofumpt` [配置在 IDE](https://github.com/mvdan/gofumpt#installation) 中，在保存时自动执行格式化

### 代码检查
本仓库的代码检查（linter）工具为 [golangci-lint]()，配置文件在 `./golangci.yml`，你可以手动运行代码检查或者将 `golangci-lint` 配置到 IDE 中实时检查代码风格；

#### 手动运行
```
make lint
```

#### 配置到 IDE 中
VSCode:

在 settings.json 添加以下配置：
```
"go.lintTool":"golangci-lint",
"go.lintFlags": [
  "--fast"
]
```

NOTE: `golangci-lint` 会自动检查项目目录下的 `.golangci.yml` 配置，你不需要将它配置到 VSCode Settings 中；

Goland:
1. Install plugin
2. Add File Watcher using existing golangci-lint template.
3. If your version of GoLand does not have the golangci-lint File Watcher template you can configure your own and use arguments run --disable=typecheck $FileDir$.

其它配置可见：https://golangci-lint.run/usage/integrations/