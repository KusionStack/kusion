# Prometheus Integration

## Table of Content
1. [Purpose](#purpose)
2. [Background](#background)
3. [Glossary](#glossary)
4. [Constraints](#constraints)
5. [Design](#design)
6. [TLS Scraping](#tls-scraping)
7. [References](#references)

## Purpose
This document captures the intended design for KusionStack's AppConfiguration model to support Prometheus-related configuration when describing an application.

Section 1-4 is intended for KusionStack users and model developers who aren't familiar with Prometheus. If you are familiar with Prometheus, you can skip the first 4 sections and head directly to [Design](#design).

## Background
### AppConfiguration
AppConfiguration model provides an interface to describe all attributes that are tied to an application.

More details on AppConfiguration model can be found in the [AppConfiguration design doc](https://github.com/KusionStack/kusion/blob/main/docs/app_configuration_model.md).

### Prometheus
Prometheus is an open-source systems monitoring and alerting toolkit originally built at SoundCloud and has now become the de facto standard for cloud-native monitoring solutions. Prometheus collects and stores its metrics as time series data, i.e. metrics information is stored with the timestamp at which it was recorded, alongside optional key-value pairs called labels<sup>[1]</sup>.

## Glossary

### Kusion concepts
For kusion related concepts such as frontend models, backend models and backbone models, please see [here](https://github.com/KusionStack/kusion/blob/main/docs/istio.md#kusion-concepts).

### Prometheus concepts
Prometheus server: The actual server that runs the Prometheus binary.

Prometheus `scrape_config`: The configuration that Prometheus honors when scraping metrics. This is defined in the Prometheus configuration (usually named `Prometheus.yml`).The full documentation can be found [here](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config).

Static scraping: As part of the scraping configuration, users can define the list of static endpoints Prometheus scrapes from in the `static_configs` section in the `scrape_config`. More can be find [here](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#static_config).
```
static_configs:
- targets: ['localhost:9090']
```

Service discovery: As opposed to static scraping, Prometheus also supports automated service discovery mechanisms that can be defined in the `xxx_sd_configs` section in the `scrape_config`, where `xxx` can be a list of options found [here](https://prometheus.io/docs/prometheus/latest/configuration/configuration). The Kubernetes-related scraping configurations are defined under `kubernetes_sd_config`, which is the most relevant configuration being discussed in this documentation.

Prometheus Operator: The Prometheus Operator provides Kubernetes native deployment and management of Prometheus and related monitoring components<sup>[2]</sup>. It leverages the [operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) to provide a way to manage Prometheus installation and configurations via CRDs.

## Constraints
KusionStack is application-centric by design and aims to provide a homogeneous experience in application delivery for all kinds of applications.

As of today, the design in this document only captures the scraping configurations Prometheus for a given application that fits into the AppConfiguration model. For example, the steps to install and configure Prometheus itself is not within consideration of this document.

### Prometheus installation
Prometheus can be installed in a Kubernetes cluster in a variety of ways. On a high level, there are a few patterns to follow:  
1. Installing the Prometheus server (along with its necessary Kubernetes resources such as `ServiceAccount`,`ClusterRole`,etc) and managing the configuration directly.
2. Installing Prometheus binary and run it in the agent mode<sup>[3]</sup> while connecting to a remote Prometheus server for aggregation. A number of cloud vendors provide monitoring solutions that resemble this approach. 
3. Using the Prometheus operator and managing the Prometheus deployment and monitoring configuration via Kubernetes CRs.

From kusion's perspective, these can be categorized into Prometheus operator vs. non-operator. Both case should be supported, but certain constraints may exist.

#### Prometheus operator
In the operator installation of Prometheus, the application scraping configuration in Kubernetes is managed via CRs, specifically:
- `ServiceMonitor` in monitoring.coreos.com/v1, for scraping services
- `PodMonitor` in monitoring.coreos.com/v1, for scraping workloads
- `Probe` in monitoring.coreos.com/v1, for scraping static targets and ingresses

These 3 CRDs above will be supported by kusion when describing an application.

The Prometheus operator acts on a few other CRDs that define the installation and configuration of Prometheus components, some of which are optional:
- `Prometheus` in monitoring.coreos.com/v1, defines a Prometheus deployment.
- `PrometheusAgent` in monitoring.coreos.com/v1alpha1, defines a Prometheus agent deployment.
- `ScrapeConfig` in monitoring.coreos.com/v1alpha1, for namespaced scraping configuration
- `PrometheusRule` in monitoring.coreos.com/v1, defines the recording and alerting rules for a Prometheus instance.
- `AlertManager` in monitoring.coreos.com/v1, defines an AlertManager cluster.
- `ThanosRuler` in monitoring.coreos.com/v1, defines a ThanosRuler deployment.

These 6 CRDs above are not within the scope of this documentation.

#### Prometheus server/agent
In the non-operator installation of Prometheus, the scraping configuration in Kubernetes is directly managed in `kubernetes_sd_configs`.

If you are managing Prometheus this way, please make sure you have the ability to directly modify this configuration. It exists in the Prometheus configuration, in the `scrape_config` section, and depending on your actual Prometheus setup, the values might be set in the `Prometheus.yml` file, a ConfigMap, or command-line arguments. More information can be found in the [Prometheus configuration documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config).

The `kubernetes_sd_config` allows customized filtering based on different criteria. Specifically, it allows Prometheus to retrieve the scraping targets(services, endpoints or workloads) by the presence of a set of annotations on the said Kubernetes resources.

#### Community most widely-used annotations
These annotations can technically be anything as long as the service discovery configuration in `scrape_config` section in the Prometheus configuration reflects the same. The most widely-used annotations are:
```
annotations:
    prometheus.io/scrape: "true"
    prometheus.io.scheme: "https"
    prometheus.io/path: "/metrics"
    prometheus.io/port: "9191"
```

The corresponding `kubernetes_sd_configs` looks like (in `relabel_configs` section):
```
...
scrape_configs:
    - job_name: 'kubernetes-service-endpoints'

    scrape_interval: 1s
    scrape_timeout: 1s

    kubernetes_sd_configs:
    - role: endpoints

    relabel_configs:
    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: replace
        target_label: __scheme__
        regex: (https?)
    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
    - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
...
```

The `source_labels` fields determine the actual key of the annotation Prometheus uses when assembling the scraping target via relabeling. `__scheme__`, `__metrics_path__`, `__address__` together points to the actual URL where Prometheus scrapes from:
```
__scheme__://__address____metrics_path__
```

However, annotation-based approach has certain limitations. A few examples of things that are impossible with this approach<sup>[4]</sup>:

- any targets with multiple ports
- any targets that don't all have the exact same authentication, http and TLS configuration
- selection is an all or nothing, not based on the well known label selection paradigm in Kubernetes

[The advise from the Prometheus team](https://github.com/prometheus-operator/prometheus-operator/issues/1547#issuecomment-401092041) is to use the `ServiceMonitor` or `PodMonitor` CRs via the Prometheus operator to manage scrape configs going forward<sup>[5]</sup>.

While acknowledging the Prometheus operator is the longer term approach, kusion should still be able to support setting an application level Prometheus scraping config via the AppConfiguration model, by making the assumption that the Prometheus installation uses the community most-adopted annotations, namely these four:
```
annotations:
    prometheus.io/scrape: "true"
    prometheus.io.scheme: "https"
    prometheus.io/path: "/metrics"
    prometheus.io/port: "9191"
```
The AppConfiguration model should automatically add the above annotations to the workloads/services/ingresses deployed via kusion when the Prometheus related configuration is present (with a set of default values such as `http`, `/metrics` and `8080`).

If the Prometheus installation is configured to recognize any other annotations on the Kubernetes objects, you can still customize the scraping config by adding those customized annotations on your workloads or services. For example:
```
annotations:
    customized.prometheus/scrape: "true"
    customized.prometheus/scheme: "https"
    customized.prometheus/path: "/metrics"
    customized.prometheus/port: "9191"
```

## Design

### Pre-requisites
- Prometheus is installed and properly configured with prometheus.yml in the cluster. This includes either the server, agent OR the operator installation.

### Goals
- Users are able to declare Prometheus-related monitoring attributes in AppConfiguration frontend model
- The backend model is able to convert the user-defined attributes to relevant configurations while rendering the Kubernetes manifests

### Supported use case
If the Prometheus operator is used to manage the monitoring configuration in the cluster (this should be indicated as a flag in the frontend model), kusion should be able to render and apply the `ServiceMonitor`, `PodMonitor` and `Probe` CRDs.

If non-operator Prometheus (server/agent) is used, kusion should automatically inject the community-standard `prometheus.io/xxxx` annotations into the workload, making the assumption that the Prometheus server/agent is configured to look for that.

If Istio is also installed in the cluster and metrics merging is enabled, it should look for the community-standard annotations on the workloads and expose the merged metrics at `/stats/prometheus` on Istio-standard port `15020`. This is Istio default behavior and does not require user participation besides making sure the annotations are in place.

### Frontend models

The frontend models should expose monitoring related attributes. Below is a sample design for `Prometheus.Scraping` (not finalized):

#### Scraping

##### Attributes
- `scheme` - (string, optional) The protocol scheme used for scraping. Possible values are `http`, `https`. Default to `http`.
- `path` - (string, optional) The path to scrape metrics from. Default to `/metrics`.
- `port` - (k8s.io/apimachinery/pkg/util/intstr.IntOrString, optional) The port to scrape metrics from. This can be a named port or a port number. Default to `8080`.
- `interval` - (protobuf.Duration, optional) The interval to scrape metrics. Only applicable if `operatorMode` is on. Default to `15s`.
- `timeout` - (protobuf.Duration, optional) The time until a scrape request times out. Only applicable if `operatorMode` is on. Default to `5s`.
- `operatorMode` - (bool, optional) Whether or not to apply scrape configs using the Prometheus operator. Requires the Prometheus operator to be present in the cluster. Default to `false`.
- `monitorType` - (string, optional) The type of resources to scrape from. Possible values are `service`, `pod`, `ingress`. Default to `pod`.

### Backend models

The backend models should provide two paths for rendering the Prometheus-related resources, based on the input of the `operatorMode` flag:
- If `operatorMode` is set to `true`, the backend model should produce one or more rendered CR(s), namely `ServiceMonitor`, `PodMonitor` or `Probe`.
- Alternative, if `operatorMode` is set to `false`, the backend model should automatically inject a set of annotations into the application pods:
```
annotations:
    prometheus.io/scrape: "xxxx"
    prometheus.io.scheme: "xxxx"
    prometheus.io/path: "/xxxx"
    prometheus.io/port: "xxxx"
```

### The backbone model

To support the Prometheus-related CRDs watched by the Prometheus operator, the backbone models need to be generated from Prometheus-operator CRDs using the `kcl-openapi` tool. More info can be found [here](https://github.com/KusionStack/kusion/blob/main/docs/istio.md#generating-backbone-models-from-istio-crds).

## TLS scraping
By default, Prometheus scrapes metrics over HTTP in cleartext. If scraping over TLS is required, or if the metrics endpoint is protected behind authentication, the TLS settings and/or authentication details must be present in the scraping configuration (or in the case of Prometheus operator, in the `ServiceMonitor` or `PodMonitor` definition). Please see the `tlsSettings`, `basic_auth`, `oauth2` and `authorization` in [scraping config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config) and [relevant CRD API Reference](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md) for more details.

### Prometheus with Istio
When Istio is installed in the cluster that Prometheus is scraping metrics from, there could be some complications.

By default, Istio-proxy configures itself in the `permissive` mode where while it expects mTLS traffic coming into the application pod, it also accepts clear-text traffic such as Prometheus scraping requests.

However, if Istio-proxy is configured to use the `strict` mTLS mode, the proxy stops accepting non-mTLS traffic, causing issues for Prometheus scraping.

#### Metrics merging and mTLS
To solve the problem, Istio have since added the capability known as [metrics merging](https://docs.google.com/document/d/1TTeN4MFmh4aUYYciR4oDBTtJsxl5-T5Tu3m3mGEdSo8/view). Metrics merging leverages the istio-agent process to scrape both the application metrics and Istio proxy metrics, and combine them together in one place for Prometheus to scrape from _without_ needing mTLS.

The Istio agent process takes note of the aforementioned `prometheus.io/xxxx` annotations and passes them as input to the Istio agent process while invoking the Istio injection mutating webhook so that Istio agent knows where to find the application metrics to combine with.

If you have the Istio mTLS mode set to `strict`, the `prometheus.io/xxxx` annotations must be present to allow Istio to locate application metrics.

Istio metrics merging is [by default on since Istio 1.7.0](https://github.com/istio/istio/pull/23433).
More on metrics merging can be found [here](https://istio.io/latest/docs/ops/integrations/prometheus/#option-1-metrics-merging).

Metrics merging have certain limitations, it does not serve the following needs<sup>[6]</sup>:
- The scraping traffic needs to be TLS-encrypted (Metrics merging serves merged metrics over cleartext scrape by Istio-agent)
- The application exposes metrics with the same name as Istio metrics
- The Prometheus server is NOT configured to look for the community-standard `prometheus.io/xxxx` annotations

If strict mTLS scraping is needed (meaning the scraping traffic needs to be encrypted over mTLS), custom TLS settings need to be present in the Prometheus scraping configuration. More info can be found [here](https://istio.io/latest/docs/ops/integrations/prometheus/#option-2-customized-scraping-configurations).

## Examples
```
monitoring: Prometheus.Scraping{
    interval        = 5s
    timeout         = 3s
    path            = "/actuator/metrics"
    port            = 9080
    scheme          = "http"
    operatorMode    = true
    monitorType     = "pod"
}
```

The above example should be rendered into the following `PodMonitor` CR:
```
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: sample-app-pod-monitor
  namespace: sample-app-ns
spec:
  selector:
    matchLabels:
      app: sample-app
  podMetricsEndpoints:
  - port: 9080
    path: "/actuator/metrics"
    scheme: "http"
    interval: 5s
    scrapeTimeout: 3s
```

Alternatively, if `operatorMode` is set to `false`(or omitted), kusion will automatically inject the following annotations into the application workload:
```
apiVersion: apps/v1
kind: Deployment
metadata:
  ...
spec:
  ...
  template:
    metadata:
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io.scheme: "http"
        prometheus.io/path: "/actuator/metrics"
        prometheus.io/port: "9080"
...
```

The above annotations should be recognized by Prometheus provided that Prometheus configs have the `__meta_kubernetes_service_annotation_prometheus_io_xxxx` meta labels defined in the `relabel_configs`, which should be on by default if you are using the extremely popular [Prometheus community helm chart](https://github.com/prometheus-community/helm-charts/blob/main/charts/prometheus/values.yaml#L828-L842) while setting up Prometheus.

## References
1. Prometheus: https://prometheus.io/docs/introduction/overview/
2. Prometheus Operator: https://github.com/prometheus-operator/prometheus-operator
3. Prometheus Agent mode: https://prometheus.io/blog/2021/11/16/agent/
4. Annotation-based approach limitations: https://github.com/prometheus-operator/prometheus-operator/issues/1547#issuecomment-401092041
5. Prometheus team advise: https://github.com/prometheus-operator/prometheus-operator/issues/1547#issuecomment-446691500
6. Istio integration with Prometheus: https://istio.io/latest/docs/ops/integrations/prometheus/