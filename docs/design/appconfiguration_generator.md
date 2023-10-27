# AppConfiguration Generator Proposal

## Motivation
Currently(v0.9.0), `AppConfigurationGenerator` is composed of a series of `GeneratorFunc`. Each `GeneratorFunc` creates a corresponding `Generator` for a subresource, and then the `Generate` method of each Generator is called to generate the resource and add it to the `spec`.

```
AppConfigurationGenerator
├─ GeneratorFunc 1
│   └─ Generator 1
│       └─ Generate
├─ GeneratorFunc 2
│   └─ Generator 2
│       └─ Generate
├─ GeneratorFunc 3
│   └─ Generator 3
│       └─ Generate
└─ ...

```
Each `NewXXXGeneratorFunc` call chain is as follows:
```
+------------------------+                                                
|  NewXXXGeneratorFunc   |                                                
|                        |                                                
+------------------------+                                                
            |                                                            
            v                                                            
+------------------------+                                                
|     XXXGenerator       |                                                
|                        |                                                
+------------------------+                                                
            |  (generate)                                                          
            v                                                            
+------------------------+                                                
|         spec           |                                                
|                        |                                                
+------------------------+
```
The current `AppConfigurationGenerator` lacks extensibility, and as more `GeneratorFunc` and `Generator`s are added, the code becomes less maintainable and readable. The following examples illustrate this problem.


**Example 1:**

Currently, each resource type has one `Generator`. However, many extension types not only require their own `Generator` but also modifications to other types, mainly the workload type. This results in `WorkloadGeneratorFunc` needing more and more input parameters and complex code logic.
```go
gfs := []appconfiguration.NewGeneratorFunc{
        NewNamespaceGeneratorFunc(g.project.Name),
        accessories.NewDatabaseGeneratorFunc(g.project, g.stack, g.appName, g.app.Workload, g.app.Database),
        workload.NewWorkloadGeneratorFunc(g.project, g.stack, g.appName, g.app.Workload, g.app.Monitoring, g.app.OpsRule),
        trait.NewOpsRuleGeneratorFunc(g.project, g.stack, g.appName, g.app),
        NewMonitoringGeneratorFunc(g.project, g.app.Monitoring, g.appName),
        // The OrderedResourcesGenerator should be executed after all resources are generated.
        NewOrderedResourcesGeneratorFunc(),
}
```
The `WorkloadGenerator` needs to handle tasks such as injecting monitoring-related annotations and labels, which should be the responsibility of the `MonitoringGenerator`.
```go
monitoringLabels := make(map[string]string)
monitoringAnnotations := make(map[string]string)
if g.monitoring != nil {
    if g.project.ProjectConfiguration.Prometheus != nil && g.project.ProjectConfiguration.Prometheus.OperatorMode {
        monitoringLabels["kusion_monitoring_appname"] = g.appName
    } else if g.project.ProjectConfiguration.Prometheus != nil && !g.project.ProjectConfiguration.Prometheus.OperatorMode {
        // If Prometheus doesn't run as an operator, kusion will generate the
        // most widely-known annotation for workloads that can be consumed by
        // the out-of-the-box community version of Prometheus server
        // installation shown as below:
        monitoringAnnotations["prometheus.io/scrape"] = "true"
        monitoringAnnotations["prometheus.io/scheme"] = g.monitoring.Scheme
        monitoringAnnotations["prometheus.io/path"] = g.monitoring.Path
        monitoringAnnotations["prometheus.io/port"] = g.monitoring.Port
    }
}
```
**Example 2:**

The `DataBaseGenerator` needs to inject env in the corresponding resources of the `WorkloadGenerator`. The current implementation requires the `DataBaseGenerator` to be executed before the `WorkloadGenerator`.
```go
func (g *databaseGenerator) injectSecret(secret *v1.Secret) error {
    // Inject the database information into the containers of service workload.
    if g.workload.Service != nil {
        for _, v := range g.workload.Service.Containers {
            v.Env[dbHostAddressEnv] = "secret://" + secret.Name + "/hostAddress"
            v.Env[dbUsernameEnv] = "secret://" + secret.Name + "/username"
            v.Env[dbPasswordEnv] = "secret://" + secret.Name + "/password"
        }
    }
    
    // Inject the database information into the containers of job workload.
    if g.workload.Job != nil {
        for _, v := range g.workload.Job.Containers {
            v.Env[dbHostAddressEnv] = "secret://" + secret.Name + "/hostAddress"
            v.Env[dbUsernameEnv] = "secret://" + secret.Name + "/username"
            v.Env[dbPasswordEnv] = "secret://" + secret.Name + "/password"
        }
    }
    
    return nil
}
```
## Goal
1. Refactor the code in the Generator section to improve its extensibility, making it easier and more maintainable to add `Generator`s.

## Non-Goal
1. It is not considered to split `Generator` out of this code repository and put it into a separate repository such as Generator-Catalog.

## Proposal

We introduce a new `XXXPatcherFunc`. The current `XXXGeneratorFunc` is used only to generate new resources, while the `XXXPatcherFunc` is used only for additional patching of resources.
```go
func (g *appConfigurationGenerator) Generate(spec *models.Spec) error {
	// Generator logic only generates new resources
	gfs := []appconfiguration.NewGeneratorFunc{
		NewNamespaceGeneratorFunc(g.project.Name),
		accessories.NewDatabaseGeneratorFunc(g.project, g.stack, g.appName, g.app.Workload, g.app.Database),
		...
	}
	if err := appconfiguration.CallGenerators(spec, gfs...); err != nil {
		return err
	}
	
	// Patcher logic patches generated resources
	pfs := []appconfiguration.NewPatcherFunc{
		trait.NewPatcherFunc(g.project, g.stack, g.appName, g.app),
	}
	if err := appconfiguration.CallPatchers(spec.Resources.KubernetesKinds(), pfs...); err != nil {
		return err
	}
	return nil
}
```


### Render Order
Perform all `GeneratorFunc` tasks first, followed by all `PatcherFunc` tasks.
### Patcher Logic
First, classify the resources in the spec by resource type (if it is a Kubernetes type), and perform patching on the specified resource types. For example, the opsRulePatcher needs to patch the MaxUnavailable field for resources of the deployment type.
```go
// Patch implements Patcher interface.
func (p *opsRulePatcher) Patch(resources map[string][]*models.Resource) error {
    if p.app.OpsRule == nil {
        return nil
    }
	
    for _, r := range resources["apps/v1:Deployment"] {
        // convert unstructured to typed object
        var deployment appsv1.Deployment
        if err := runtime.DefaultUnstructuredConverter.FromUnstructured(r.Attributes, &deployment); err != nil {
            return err
        }
        maxUnavailable := intstr.Parse(p.app.OpsRule.MaxUnavailable)
        deployment.Spec.Strategy.RollingUpdate.MaxUnavailable = &maxUnavailable
        // convert typed object to unstructured
        updated, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&deployment)
        if err != nil {
            return err
        }
        r.Attributes = updated
    }
    return nil
}
```

Resource type classification can be done by saving the type information of this resource in `Resource.Extensions`, or by extracting it from the ID (e.g. apps/v1:Deployment:testproject:testproject-teststack-app1).
```go
// KubernetesKinds returns a map of Kubernetes GVK to resources
func (rs Resources) KubernetesKinds() map[string][]*Resource {
    m := make(map[string][]*Resource)
    for i := range rs {
        resource := &rs[i]
        if resource.Type != Kubernetes {
            continue
        }
        gvk := resource.Extensions["groupVersionKind"].(string)
        m[gvk] = append(m[gvk], resource)
    }
    return m
}
```