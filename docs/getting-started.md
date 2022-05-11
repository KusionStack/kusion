# Getting Started

## Requirements

- Have [Konfig](https://github.com/KusionStack/Konfig) cloned.
- Have a [kubeconfig](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) file (default location is `~/.kube/config`).

## Install Kusion

Kusion is a binary file in nature, we offer different format for all platform, including windows, linux, mac, etc.
Kusion is responsible for KCL code which is compiled by KCLVM. These requirements are packed in [kusionup](https://github.com/KusionStack/kusionup) tool.
So, we install `kusionup` first, and then pick up a compatible version for your working environment.
You can see installation details from [here](https://kusionstack.io/docs/user_docs/getting-started/install/kusionup/).

Next, let's start kusion installing:

```shell
kusionup install
```

The command above will install a right self-adapting version based on your machine. 
After finishing installing, run following command to make sure kusion can work correctly.

```shell
kusion version
```

## Init Project

`kusion init` will initialize KCL file structure and base codes for a new project.
Before moving to next step, make sure you have cloned [Konfig](https://github.com/KusionStack/Konfig).
Now enter root directory of konfig, cause the new project imported Konfig repo.

First, run the following command:

```shell
kusion init
```

the output is similar to:

```shell
? Please choose a template:  [Use arrows to move, type to filter]
> deployment-single-stack    A minimal kusion project of single stack
```

`kusion init` provides online/offline mode, default is offline. You can append flag `--yes` to switch online mode.

Next, just press *Enter* and kusion is asking the required params, both of project and stack:

```shell
This command will walk you through creating a new kusion project.

Enter a value or leave blank to accept the (default), and press <ENTER>.
Press ^C at any time to quit.

Project Config:
? Project Name: [? for help] (my-app) 
```

Then, you can input value or just use default from template(just press on *Enter* continuously).

The output is similar to:

```shell
Project Config:
? Project Name: my-app
? ServiceName: frontend-svc
? NodePort: 30000
? ProjectName: my-app
Stack Config: dev
? Stack: dev
? Image: gcr.io/google-samples/gb-frontend:v4
? ClusterName: kubernetes-dev
Created project 'my-app'
```

## Compile Locally

`kusion compile` compiles KCL into YAML.
Now, you have a simple project of kubernetes Deployment and Service in the same Namespace.

To compile KCL, run the following command: 

```shell
kusion compile -w my-app/dev
```

and check file `my-app/dev/ci-test/stdout.golden.yaml`, see the compiled result.

## Apply Configuration

`kusion apply` applies a configuration stack of resource(s) by work directory to Kubernetes runtime.
We are starting apply the compiled result to Kubernetes.

Now, run the following command:

```shell
kusion apply -w my-app/dev
```

the output is similar to:

```shell
 SUCCESS  Compiling in stack dev...                                                                                                                                                                    

Stack: dev  ID                                   Action
 * ├─       v1:Namespace:my-app                  Create
 * ├─       v1:Service:my-app:frontend-svc       Create
 * └─       apps/v1:Deployment:my-app:my-appdev  Create

? Do you want to apply these diffs?  [Use arrows to move, type to filter]
  yes
> details
  no
```

Next, move arrows up, choose `yes` option and press *Enter*, you will see:

```shell
Start applying diffs ...
 SUCCESS  Create v1:Namespace:my-app success                                                                                                                                                           
 SUCCESS  Create v1:Service:my-app:frontend-svc success                                                                                                                                                
 SUCCESS  Create apps/v1:Deployment:my-app:my-appdev success                                                                                                                                           
Create apps/v1:Deployment:my-app:my-appdev success [3/3] █████████████ 100% | 0s

Apply complete! Resources: 3 created, 0 updated, 0 deleted.
```

Last, you can check Deployment and Service is running or not with `kubectl` tool.

## Clean Up

Kusion has created a Namespace, a Deployment and a Service. It can do creation, also undo it. 
`kusion destroy` helps you to delete a configuration stack. 
Run the following command, these resources just created will be deleted:

```shell
kusion destroy -w my-app/dev
```

The output is similar to:

```shell
 SUCCESS  Compiling in stack dev...                                                                                                                                                                    

Stack: dev  ID                                   Action
 * ├─       apps/v1:Deployment:my-app:my-appdev  Delete
 * ├─       v1:Service:my-app:frontend-svc       Delete
 * └─       v1:Namespace:my-app                  Delete

? Do you want to destroy these diffs?  [Use arrows to move, type to filter]
  yes
> details
  no
```

Just like apply command, move arrow to `yes` and press *Enter*, you will see:

```shell
Start destroying resources......
 SUCCESS  Delete apps/v1:Deployment:my-app:my-appdev success                                                                                                                                           
 SUCCESS  Delete v1:Service:my-app:frontend-svc success                                                                                                                                                
 SUCCESS  Delete v1:Namespace:my-app success                                                                                                                                                           
Delete v1:Namespace:my-app success [3/3] █████████████████████████████ 100% | 0s
```
