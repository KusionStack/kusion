# 前言

> 本 README.md 包括配置代码仓库目录/文件说明及如何本地使用 Kusion+Minikube 进行测试

## 快速开始

```bash
$ cd dev
$ kusion apply
SUCCESS  Compiling in stack dev...

Stack: dev    Provider                          Type             Name    Plan
      * ├─  kubernetes                  v1:Namespace        http-echo  Create
      * ├─  kubernetes            apps/v1:Deployment     http-echodev  Create
      * ├─  kubernetes                    v1:Service    apple-service  Create
      * └─  kubernetes  networking.k8s.io/v1:Ingress  example-ingress  Create

✔ yes
Start applying diffs......
SUCCESS  Creating Namespace/http-echo
SUCCESS  Creating Deployment/http-echodev
SUCCESS  Creating Service/apple-service
SUCCESS  Creating Ingress/example-ingress

Creating Ingress/example-ingress [4/4] ████████████████████████████████ 100% | 0s

$ minikube service apple-service -n http-echo --url
http://192.168.99.102:30206

$ curl http://192.168.99.102:30206/apple                         
apple

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
