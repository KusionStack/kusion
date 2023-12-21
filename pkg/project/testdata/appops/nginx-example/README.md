# 前言

> 本 README.md 包括配置代码仓库目录/文件说明及如何本地使用 Kusion+Minikube 进行测试

## 快速开始

```bash
$ cd dev
$ kusion apply
SUCCESS  Compiling in stack dev...

Stack: dev    Provider                Type              Name    Plan
      * ├─  kubernetes        v1:Namespace     nginx-example  Create
      * ├─  kubernetes          v1:Service     nginx-example  Create
      * └─  kubernetes  apps/v1:Deployment  nginx-exampledev  Create

✔ yes
Start applying diffs......
SUCCESS  Creating Namespace/nginx-example
SUCCESS  Creating Service/nginx-example
SUCCESS  Creating Deployment/nginx-exampledev

Creating Deployment/nginx-exampledev [3/3] ████████████████████████████████ 100% | 0s

$ minikube service nginx-example -n nginx-example --url
http://192.168.99.102:30201

$ curl -s http://192.168.99.102:30201 | grep '<title>'   # Or visit http://192.168.99.102:30201 in browser
<title>Welcome to nginx!</title>

$ kusion destroy
```

## 目录和文件说明

```bash
.
├── base                        // 各环境通用配置
│   ├── base.k                  // 应用的环境通用配置
├── prod                        // 环境目录
│   └── ci-test                 // ci 测试目录，放置测试脚本和数据
│     └── settings.yaml         // 测试数据和编译文件配置
│     └── stdout.golden.yaml    // 期望的 YAML，可通过 make 更新
│   └── kcl.yaml                // 当前 Stack 的多文件编译配置
│   └── main.k                  // 应用在当前环境的配置清单
│   └── stack.yaml              // Stack 元信息
└── project.yaml	            // Project 元信息
└── README.md                   // 说明文档
```
