---
title: "快速部署 Kusion 项目"
linkTitle: "快速部署 Kusion 项目"
date: 2020-02-17
weight: 2
description: >
 本文将介绍如何快速部署 Kusion 项目。
---

## 下载，安装和卸载
Kusion 二进制文件可用于 Mac 和 Linux。
下载地址： http://gitlab./.net/kusion/artifacts

### 安装
> 您可以通过如下命令安装 KusionCtl
> wget -q https://code.kusionstack.io/kusion/tree/master/scripts/install.sh -O -

通过如下命令验证安装结果
> kusion -v
> kcl -V

安装成功后，将 '~/.kusion/bin'，'~/.kusion/kclvm/bin' 加入到 $PATH 环境变量中。

### 更新
> 您可以通过如下命令更新 KusionCtl
> wget -q https://code.kusionstack.io/kusion/tree/master/scripts/upgrade.sh -O -

### 卸载
> 您可以通过如下命令卸载 KusionCtl
> wget -q https://code.kusionstack.io/kusion/tree/master/scripts/uninstall.sh -O -

卸载成功后，将 '~/.kusion/bin'，'~/.kusion/kclvm/bin' 从 $PATH 环境变量中移除。

## 编写资源声明
### 使用kusion_model
使用 kusion models 定义您的云原生基础设施，参考[最佳实践范例](docs/samples/_index.md)

## 生成 yaml 文件
### 命令说明
```
kusion apply -f <filePath>
```
### 使用示例
[embed: 2-guestbook-yaml-components.mp4](https://intranetproxy.kusionstack.io/skylark/lark/0/2020/mp4/122885/1579237956731-42620b92-af44-4006-9f9f-cd4648927762.mp4)

## 提交到 k8s 集群
### 命令说明
```
kusion apply --target runtime --timeout 20s --print-trace -f <filePath>
```
### 使用示例
[embed: 2-guestbook-k8s.mp4](https://intranetproxy.kusionstack.io/skylark/lark/0/2020/mp4/122885/1579238064285-5c4a7a61-e418-48e5-982a-406ac52a502e.mp4)

