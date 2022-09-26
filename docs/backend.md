# backend
kusion state backends定义state存储位置，默认情况下，kusion使用local类型存储state在本地磁盘上，对于团队协作项目，state可存储在远程服务上，允许多人使用

## backend 配置
### 配置文件

kusion 通过 project.yaml 中 backend 配置储存，例如
```yaml
backend:
  storageType: local
  config:
    path: kusion_state.json
```
* storageType - 声明储存类型
* config - 声明对应存储类型所需参数
### 命令行配置
```sh
kusion apply --backend-type local --backend-config path=kusion-state.json
```
### 合并配置
当配置文件中 config 和命令行中 --backend-config 同时配置时，整个配置合并配置文件和命令行配置，例如
```yaml
backend:
  storageType: local
  config:
    path: kusion_state.json
```

```sh
kusion apply --backend-config path-state=kusion-state.json
```
合并后 backend config 为
```yaml
backend:
  storageType: local
  config:
    path: kusion_state.json
    path-state: kusion-state.json
```

## 可用Backend
- local
- oss
- s3
- db

### 默认Backend

当配置文件及命令行都没有声明 Backend 配置时，默认使用 [local](#local)

### local
local类型存储state在本地文件系统上，在本地操作，不适用于多人协同

配置示例:
```yaml
backend:
  storageType: local
  config:
    path: kusion_state.json
```
* storageType - local, 表示使用本地文件系统
* path - (可选) 配置 state 本地存储文件

### oss

oss 类型存储 state 在阿里云 OSS 上

配置示例:
```yaml
backend:
  storageType: oss
  config:
    endpoint: oss-cn-beijing.aliyuncs.com
    bucket: kusion-oss
```
```sh
kusion apply -C accessKeyID=********** -C accessKeySecret=************
```

* storageType - oss, 表示使用阿里云 OSS 对象存储
* endpoint - (必选) 阿里云 OSS 访问地址
* bucket - (必选) 阿里云 bucket 名称
* accessKeyID - (必选) 阿里云 accessKeyID
* accessKeySecret - (必选) 阿里云 accessKeySecret

### s3 

s3 类型存储 state 在 AWS S3 对象存储

```yaml
backend:
  storageType: s3
  config:
    endpoint: http://localhost:9000
    bucket: kusion-s3
    region: us-east-1
```
```sh
kusion apply -C accessKeyID=********* -C accessKeySecret=*********
```

* storageType - s3, 表示使用 AWS S3 对象存储
* endpoint - (必选) AWS S3 访问地址
* bucket - (必选) S3 bucket 名称
* accessKeyID - (必选) AWS accessKeyID
* accessKeySecret - (必选) AWS accessKeySecret

### db

db 类型存储 state 在 数据库中

```yaml
backend:
  storageType: db
  config:
    dbHost: 127.0.0.1
    dbName: kusiondemo
    dbPort: 3306
```

```sh
kusion apply -C dbUser=********* -C dbPassword=**********
```

* storageType - db, 表示使用数据库存储
* dbHost - (必选) 数据库访问地址
* dbName - (必选) 数据库名称
* dbPort - (必选) 数据库端口
* dbUser - (必选) 数据库用户
* dbPassword - (必选) 数据库访问密码
