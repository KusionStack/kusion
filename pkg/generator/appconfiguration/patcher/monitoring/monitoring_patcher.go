package monitoring

import (
	appsv1 "k8s.io/api/apps/v1"

	"kusionstack.io/kube-api/apps/v1alpha1"
	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	modelsapp "kusionstack.io/kusion/pkg/models/appconfiguration"
	"kusionstack.io/kusion/pkg/projectstack"
)

type monitoringPatcher struct {
	appName string
	app     *modelsapp.AppConfiguration
	project *projectstack.Project
}

// NewMonitoringPatcher returns a Patcher.
func NewMonitoringPatcher(appName string, app *modelsapp.AppConfiguration, project *projectstack.Project) (appconfiguration.Patcher, error) {
	return &monitoringPatcher{
		appName: appName,
		app:     app,
		project: project,
	}, nil
}

// NewMonitoringPatcherFunc returns a NewPatcherFunc.
func NewMonitoringPatcherFunc(appName string, app *modelsapp.AppConfiguration, project *projectstack.Project) appconfiguration.NewPatcherFunc {
	return func() (appconfiguration.Patcher, error) {
		return NewMonitoringPatcher(appName, app, project)
	}
}

// Patch implements Patcher interface.
func (p *monitoringPatcher) Patch(resources map[string][]*models.Resource) error {
	if p.app.Monitoring == nil || p.project.ProjectConfiguration.Prometheus == nil {
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

	if p.project.ProjectConfiguration.Prometheus.OperatorMode {
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

	if err := appconfiguration.PatchResource[appsv1.Deployment](resources, appconfiguration.GVKDeployment, func(obj *appsv1.Deployment) error {
		obj.Labels = appconfiguration.MergeMaps(obj.Labels, monitoringLabels)
		obj.Annotations = appconfiguration.MergeMaps(obj.Annotations, monitoringAnnotations)
		obj.Spec.Template.Labels = appconfiguration.MergeMaps(obj.Spec.Template.Labels, monitoringLabels)
		obj.Spec.Template.Annotations = appconfiguration.MergeMaps(obj.Spec.Template.Annotations, monitoringAnnotations)
		return nil
	}); err != nil {
		return err
	}

	if err := appconfiguration.PatchResource[v1alpha1.CollaSet](resources, appconfiguration.GVKDeployment, func(obj *v1alpha1.CollaSet) error {
		obj.Labels = appconfiguration.MergeMaps(obj.Labels, monitoringLabels)
		obj.Annotations = appconfiguration.MergeMaps(obj.Annotations, monitoringAnnotations)
		obj.Spec.Template.Labels = appconfiguration.MergeMaps(obj.Spec.Template.Labels, monitoringLabels)
		obj.Spec.Template.Annotations = appconfiguration.MergeMaps(obj.Spec.Template.Annotations, monitoringAnnotations)
		return nil
	}); err != nil {
		return err
	}
	return nil
}
