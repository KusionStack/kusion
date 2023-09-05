package workload

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kube-api/apps/v1alpha1"

	"k8s.io/apimachinery/pkg/util/intstr"
	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/generator/appconfiguration/generator/workload/network"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/monitoring"
	"kusionstack.io/kusion/pkg/models/appconfiguration/trait"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/projectstack"
)

// workloadServiceGenerator is a struct for generating service workload resources.
type workloadServiceGenerator struct {
	project    *projectstack.Project
	stack      *projectstack.Stack
	appName    string
	service    *workload.Service
	monitoring *monitoring.Monitor
	opsRule    *trait.OpsRule
}

// NewWorkloadServiceGenerator returns a new workloadServiceGenerator instance.
func NewWorkloadServiceGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	service *workload.Service,
	monitoring *monitoring.Monitor,
	opsRule *trait.OpsRule,
) (appconfiguration.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(appName) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}

	if service == nil {
		return nil, fmt.Errorf("service workload must not be nil")
	}

	return &workloadServiceGenerator{
		project:    project,
		stack:      stack,
		appName:    appName,
		service:    service,
		monitoring: monitoring,
		opsRule:    opsRule,
	}, nil
}

// NewWorkloadServiceGeneratorFunc returns a new NewGeneratorFunc that returns a workloadServiceGenerator instance.
func NewWorkloadServiceGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	service *workload.Service,
	monitoring *monitoring.Monitor,
	opsRule *trait.OpsRule,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewWorkloadServiceGenerator(project, stack, appName, service, monitoring, opsRule)
	}
}

// Generate generates a service workload resource to the given spec.
func (g *workloadServiceGenerator) Generate(spec *models.Spec) error {
	service := g.service
	if service == nil {
		return nil
	}

	// Create an empty resource slice if it doesn't exist yet.
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	uniqueAppName := appconfiguration.UniqueAppName(g.project.Name, g.stack.Name, g.appName)

	// Create a slice of containers based on the app's
	// containers along with related volumes and configMaps.
	containers, volumes, configMaps, err := toOrderedContainers(service.Containers, uniqueAppName)
	if err != nil {
		return err
	}

	// Create ConfigMap objects based on the app's configuration.
	for _, cm := range configMaps {
		cmObj := cm
		cmObj.Namespace = g.project.Name
		if err = appconfiguration.AppendToSpec(
			models.Kubernetes,
			appconfiguration.KubernetesResourceID(cmObj.TypeMeta, cmObj.ObjectMeta),
			spec,
			&cmObj,
		); err != nil {
			return err
		}
	}

	// If Prometheus runs as an operator, it relies on Custom Resources to
	// manage the scrape configs. CRs (ServiceMonitors and PodMonitors) rely on
	// corresponding resources (Services and Pods) to have labels that can be
	// used as part of the label selector for the CR to determine which
	// service/pods to scrape from.
	// Here we choose the label name kusion_monitoring_appname for two reasons:
	// 1. Unlike the label validation in Kubernetes, the label name accepted by
	// Prometheus cannot contain non-alphanumeric characters except underscore:
	// https://github.com/prometheus/common/blob/main/model/labels.go#L94
	// 2. The name should be unique enough that is only created by Kusion and
	// used to identify a certain application
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

	labels := appconfiguration.MergeMaps(appconfiguration.UniqueAppLabels(g.project.Name, g.appName), g.service.Labels, monitoringLabels)
	annotations := appconfiguration.MergeMaps(g.service.Annotations, monitoringAnnotations)
	selector := appconfiguration.UniqueAppLabels(g.project.Name, g.appName)

	// Create a K8s workload object based on the app's configuration.
	// common parts
	objectMeta := metav1.ObjectMeta{
		Labels:      labels,
		Annotations: annotations,
		Name:        uniqueAppName,
		Namespace:   g.project.Name,
	}
	podTemplateSpec := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			Containers: containers,
			Volumes:    volumes,
		},
	}

	var resource any
	typeMeta := metav1.TypeMeta{}

	switch service.Type {
	case workload.TypeDeploy:
		typeMeta = metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       workload.TypeDeploy,
		}
		spec := appsv1.DeploymentSpec{
			Replicas: appconfiguration.GenericPtr(int32(service.Replicas)),
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: podTemplateSpec,
		}
		if g.opsRule != nil && g.opsRule.MaxUnavailable != "" {
			maxUnavailable := intstr.Parse(g.opsRule.MaxUnavailable)
			spec.Strategy = appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &maxUnavailable,
				},
			}
		}
		resource = &appsv1.Deployment{
			TypeMeta:   typeMeta,
			ObjectMeta: objectMeta,
			Spec:       spec,
		}
	case workload.TypeCollaset:
		typeMeta = metav1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.String(),
			Kind:       workload.TypeCollaset,
		}
		resource = &v1alpha1.CollaSet{
			TypeMeta:   typeMeta,
			ObjectMeta: objectMeta,
			Spec: v1alpha1.CollaSetSpec{
				Replicas: appconfiguration.GenericPtr(int32(service.Replicas)),
				Selector: &metav1.LabelSelector{MatchLabels: selector},
				Template: podTemplateSpec,
			},
		}
	}

	// Add the Deployment resource to the spec.
	if err = appconfiguration.AppendToSpec(models.Kubernetes, appconfiguration.KubernetesResourceID(typeMeta, objectMeta), spec, resource); err != nil {
		return err
	}

	// generate K8s Service from ports config.
	portsGeneratorFunc := network.NewPortsGeneratorFunc(g.appName, g.project.Name, g.stack.Name, selector, labels, annotations, g.service.Ports)
	if err = appconfiguration.CallGenerators(spec, portsGeneratorFunc); err != nil {
		return err
	}

	return nil
}
