package monitoring

import (
	appsv1 "k8s.io/api/apps/v1"

	"kusionstack.io/kube-api/apps/v1alpha1"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/modules"
	modelsapp "kusionstack.io/kusion/pkg/modules/inputs"
)

type monitoringPatcher struct {
	appName string
	app     *modelsapp.AppConfiguration
	project *project.Project
}

// NewMonitoringPatcher returns a Patcher.
func NewMonitoringPatcher(appName string, app *modelsapp.AppConfiguration, project *project.Project) (modules.Patcher, error) {
	return &monitoringPatcher{
		appName: appName,
		app:     app,
		project: project,
	}, nil
}

// NewMonitoringPatcherFunc returns a NewPatcherFunc.
func NewMonitoringPatcherFunc(appName string, app *modelsapp.AppConfiguration, project *project.Project) modules.NewPatcherFunc {
	return func() (modules.Patcher, error) {
		return NewMonitoringPatcher(appName, app, project)
	}
}

// Patch implements Patcher interface.
func (p *monitoringPatcher) Patch(resources map[string][]*intent.Resource) error {
	if p.app.Monitoring == nil || p.project.Configuration.Prometheus == nil {
		return nil
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

	if p.project.Configuration.Prometheus.OperatorMode {
		monitoringLabels["kusion_monitoring_appname"] = p.appName
	} else {
		// If Prometheus doesn't run as an operator, kusion will generate the
		// most widely-known annotation for workloads that can be consumed by
		// the out-of-the-box community version of Prometheus server
		// installation shown as below:
		monitoringAnnotations["prometheus.io/scrape"] = "true"
		monitoringAnnotations["prometheus.io/scheme"] = p.app.Monitoring.Scheme
		monitoringAnnotations["prometheus.io/path"] = p.app.Monitoring.Path
		monitoringAnnotations["prometheus.io/port"] = p.app.Monitoring.Port
	}

	if err := modules.PatchResource[appsv1.Deployment](resources, modules.GVKDeployment, func(obj *appsv1.Deployment) error {
		obj.Labels = modules.MergeMaps(obj.Labels, monitoringLabels)
		obj.Annotations = modules.MergeMaps(obj.Annotations, monitoringAnnotations)
		obj.Spec.Template.Labels = modules.MergeMaps(obj.Spec.Template.Labels, monitoringLabels)
		obj.Spec.Template.Annotations = modules.MergeMaps(obj.Spec.Template.Annotations, monitoringAnnotations)
		return nil
	}); err != nil {
		return err
	}

	if err := modules.PatchResource[v1alpha1.CollaSet](resources, modules.GVKDeployment, func(obj *v1alpha1.CollaSet) error {
		obj.Labels = modules.MergeMaps(obj.Labels, monitoringLabels)
		obj.Annotations = modules.MergeMaps(obj.Annotations, monitoringAnnotations)
		obj.Spec.Template.Labels = modules.MergeMaps(obj.Spec.Template.Labels, monitoringLabels)
		obj.Spec.Template.Annotations = modules.MergeMaps(obj.Spec.Template.Annotations, monitoringAnnotations)
		return nil
	}); err != nil {
		return err
	}
	return nil
}
