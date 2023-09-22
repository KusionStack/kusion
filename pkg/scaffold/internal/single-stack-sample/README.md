# 前言
> 本 README.md 包括配置代码仓库目录/文件说明及如何本地使用 Kusion+Minikube 进行测试

## 快速开始
```bash
$ cd dev
$ kusion apply
 ✔︎  Generating Spec in the Stack dev...
Stack: dev  ID                                                  Action
* ├─       v1:Namespace:helloworld                             Create
* ├─       v1:Service:helloworld:helloworld-dev-nginx-private  Create
* └─       apps/v1:Deployment:helloworld:helloworld-dev-nginx  Create

? Do you want to apply these diffs? yes
Start applying diffs ...
 SUCCESS  Create v1:Namespace:helloworld success
 SUCCESS  Create v1:Service:helloworld:helloworld-dev-nginx-private success
 SUCCESS  Create apps/v1:Deployment:helloworld:helloworld-dev-nginx success
Create apps/v1:Deployment:helloworld:helloworld-dev-nginx success [3/3] ████████████████████████ 100% | 0s
Apply complete! Resources: 3 created, 0 updated, 0 deleted.

$ kubectl port-forward svc/helloworld-dev-nginx-private -n helloworld 30000:80

$ curl -s http://127.0.0.1:30000 | grep title
<title>Welcome to nginx!</title>

$ kusion destroy
```

## 目录和文件说明
```bash
.
├── README.md                   // 说明文档
├── dev                         // 环境目录
│   ├── kcl.mod                 // KCL Package 声明文件
│   ├── kcl.mod.lock            // KCL Package Lock 文件
│   └── main.k                  // 应用在当前环境的配置清单
│   └── stack.yaml              // Stack 元信息
└── project.yaml	            // Project 元信息
```
