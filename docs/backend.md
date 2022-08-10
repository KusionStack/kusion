# backend
kusion state backends定义state存储位置，默认情况下，kusion使用local类型存储state在本地磁盘上，对于团队协作项目，state可存储在远程服务上，允许多人使用

## backend 配置
### 配置文件

kusion 通过 project.yaml 中 backend 配置储存，例如
```
backend:
  storageType: local
  config:
    path: kusion_state.json
```
* storageType - 声明储存类型
* config - 声明对应存储类型所需参数
### 命令行配置
```
kusion apply --backend-type local --backend-config path=kusion-state.json
```
### 合并配置
当配置文件中 config 和命令行中 --backend-config 同时配置时，整个配置合并配置文件和命令行配置，例如
```
backend:
  storageType: local
  config:
    path: kusion_state.json
```
```
kusion apply --backend-config path-state=kusion-state.json
```
合并后 backend config 为
```
backend:
  storageType: local
  config:
    path: kusion_state.json
    path-state: kusion-state.json
```
## 可用Backend
- local

### 默认Backend

当配置文件及命令行都没有声明 Backend 配置时，默认使用 [local](#local)

### local
local类型存储state在本地文件系统上，在本地操作，不适用于多人协同

配置示例:
```
backend:
  storageType: local
  config:
    path: kusion_state.json
```
* storageType - local, 表示使用本地文件系统
* path - (可选) 配置 state 本地存储文件




