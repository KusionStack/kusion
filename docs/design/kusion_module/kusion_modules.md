# Kusion Module Design Doc

## Definition

Kusion module is a reusable building block of KusionStack designed by platform engineers. Here are some explanations to make the definition more clear:

1. It represents an independent unit that provides a specific capability to the application with clear business semantics.
2. It consists of one or multiple infrastructure resources (K8s/Terraform resources), but it is not merely a collection of unrelated resources. For instance, a database, monitoring capabilities, and network access are typical Kusion Modules.
3. Modules should not have dependencies or be nested within each other.
4. AppConfig is not a Module.

For more details, please visit our [official website](https://www.kusionstack.io/docs/kusion/concepts/kusion-module).

![module](../collaboration/kusion-module.png)

## Goals

1. Design a flexible Kusion module mechanism to manage and use modules
2. Provide a user-friendly framework to enable users to develop customized modules

## Modules in AppConfiguration

```python
import models.schema.v1 as ac
import models.schema.v1.workload as wl
import models.schema.v1.workload.container as c
import models.schema.v1.workload.container.probe as p
import models.schema.v1.monitoring as m
import models.schema.v1.database as d

# Note: AppConfiguration per se is not a Kusion Module
  helloWorld: ac.AppConfiguration {
      # Built-in module
      workload: wl.Service {
          containers: {
              "main": c.Container {
                  image: "ghcr.io/kusion-stack/samples/helloworld:latest"
                  # Configure a HTTP readiness probe
                  readinessProbe: p.Probe {
                      probeHandler: p.Http {
                          url: "http://localhost:80"
                      }
                  }
              }
          }
      }

  # A collection of accessories that will be attached to the workload
    accessories: {
        # Built-in module, key represents the module source
        "kusionstack/mysql@v0.1.0" : d.MySQL {
            type: "cloud"
            version: "8.0"
        }
        # Built-in module, key represents the module source
        "kusionstack/prometheus@v0.1.0" : m.Prometheus {
            path: "/metrics"
        }
        # Customized module, key represents the module source
        "foo/customize@v0.1.0": customizedModule {
                ...
        }
    }

      # pipeline modules
      pipeline: {
          # Step is a module
          "step" : Step {
              use: "exec"
              args: ["--test-all"]
          }
      }

      # Dependent app list
      dependency: {
          dependentApps: ["init-kusion"]
      }
}
```

## Structure

An app dev-orient schema, a generator and a license file are three components required for a legal Kusion module. We strongly recommend adding a readme file and examples in the module package for completeness. An example module package is shown as follows.

```shell
$ tree example-module/
.
├── schema.k
├── kusion-module-name_v0.1.0 # binary
├── kcl.mod
├── README.md
├── LICENSE
├── examples/
│   ├── main.k
│   ├── workspace.yaml
``` 

## Lifecycle

### Execution lifecycle

#### Download and unzip

A complete set of modules of one stack consists of two parts: modules in the AppConfig and modules in the workspace. In most scenes, the two parts are the same, but modules in the workspace can be bigger than those in the AppConfig as **some modules do not contain schemas**.

We need to set KPM download path as `$KUSION_HOME/modules`, since it will always download schema-related modules in a certain path to guarantee it works correctly and then Kusion will download extra modules defined in the workspace.

**Note:** In the first version of Kusion module, we assume modules in the AppConfig and workspace are the same.

#### Build Intent

All KCL codes written by app devs will be compiled by KPM and output an intermediate YAML. Kusion combines this YAML and corresponding workspace configurations as inputs of Kusion module generators and invokes these generators to get the final Intent.

Considering workload is required for every application and other modules depend on it, Kusion will execute the `workload` module at first to generate the workload resource. For modules that need to modify attributes in the `workload` such as environments, labels and annotations, we provide a `patch`` interface to fulfill this demand.

##### Generator Interface

Kusion invokes all module generators described through gRPC with [go-plugin](https://github.com/hashicorp/go-plugin) and provides a framework to deserialize and validate input and output values to guarantee correctness. Interfaces are defined below.

```protobuf
// GeneratorRequest represents a request to generate something based on the project details
message GeneratorRequest {
  // Project represents the project name
  string project = 1;
  // Stack represents the stack name
  string stack = 2;
  // App represents the application name, which is typically the same as the namespace of Kubernetes resources
  string app = 3;
  // Workload represents the v1.Workload defined in the AppConfiguration
  bytes workload = 4;
  // DevModuleConfig is the developer's inputs of this module
  bytes dev_module_config = 5;
  // PlatformModuleConfig is the platform engineer's inputs of this module
  bytes platform_module_config = 6;
  // RuntimeConfig is the runtime configurations defined in the workspace
  bytes runtime_config = 7;
}

// GeneratorResponse represents the generate result of the generator.
message GeneratorResponse {
  // Resources is a v1.Resource array, which represents the generated resources by this module.
  repeated bytes resources = 1;
}

service Module {
  rpc Generate(GeneratorRequest) returns (GeneratorResponse);
}
```
```go
type Intent.Resource struct {
	...
	// Add a new field to represent patchers
	Patchers []Patcher
}

// Kusion will patch these fields into the workload corresponding fields
type Patcher struct{
	Environments map[string]string `json:"environments" yaml:"environments"`
	Labels map[string]string `json:"labels" yaml:"labels"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
	...
} 
```

### Clean up

Close all connections with one module generator once it has been executed.

## Develop lifecycle

### Set up a developing environment

Developing a Kusion module includes defining a KCL schema and developing a Go project. We will provide a scaffold framework repository to help developers set up the developing environment easily.

Download this repository by `git clone` and rename this project with your module name. The scaffold contains code templates and all files needed for developing a Kusion module. The structure looks like this:

```shell
$ tree example-module/
.
├── schema.k
├── generator.go
├── go.mod
├── go.sum
├── kcl.mod
├── README.md
├── LICENSE
```

### Developing

1. Communicate with app developers and identify the schema parameters.
2. Identify the left module input parameters initialized in the workspace
3. Define the app dev-orient schema
4. Develop the generator by implementing gRPC interfaces

### Local validation

We will provide a new command `kusion module build` to help developers build a module from the root directory of this project. Once this new module is built, you can move it to `$KUSION_HOME/modules` and validate this module with Kusion CLI commands.

### Publish

Publish the Kusion module to a registry with the command `kusion module publish -r [registry path]`

## Relationship

![rel](relationship.jpg)

## An open question -- How to manage on-prem infrastructures

According to the definition of the Kusion module, it is responsible for generating the Spec and passing it to the Kusion engine to make it active. For cloud resources and Kubernetes, we currently leverage Terraform providers and Kubernetes operators to manage these resources effectively. But for platform engineers who want to manage their on-premises infrastructures with Kusion, what are they supposed to do? Here are two methods.

1. Obviously, platform engineers can develop a Kubernetes operator or a Terraform provider, along with an associated Kusion module, and then publish it to a provider or Helm registry. However, this workflow is too fragmented and they have to maintain two separate logics with completely different workflows.
2. Extending the functionality of Kusion modules to include the logic for operating actual infrastructures. This would unify the development experience by providing a complete building block, including definition and execution.

The second method raises the question of whether Kusion module should be compatible with existing Terraform providers or Kubernetes operators. If compatibility is desired, we could develop an adapter or a shim to convert Terraform providers into Kusion modules. We have seen some projects have done this before, but such an adapter would be very complex and challenging to catch up with the upstream Terraform provider framework.

We are still considering this question. Any suggestions or ideas are welcome, please feel free to open an issue or a discussion in our repository.