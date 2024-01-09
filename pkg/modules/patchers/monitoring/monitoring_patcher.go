package monitoring

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"kusionstack.io/kube-api/apps/v1alpha1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/monitoring"
	"kusionstack.io/kusion/pkg/workspace"
)

type monitoringPatcher struct {
	app           *inputs.AppConfiguration
	modulesConfig map[string]apiv1.GenericConfig
}

// NewMonitoringPatcher returns a Patcher.
func NewMonitoringPatcher(app *inputs.AppConfiguration, modulesConfig map[string]apiv1.GenericConfig) (modules.Patcher, error) {
	return &monitoringPatcher{
		app:           app,
		modulesConfig: modulesConfig,
	}, nil
}

// NewMonitoringPatcherFunc returns a NewPatcherFunc.
func NewMonitoringPatcherFunc(app *inputs.AppConfiguration, modulesConfig map[string]apiv1.GenericConfig) modules.NewPatcherFunc {
	return func() (modules.Patcher, error) {
		return NewMonitoringPatcher(app, modulesConfig)
	}
}

// Patch implements Patcher interface.
func (p *monitoringPatcher) Patch(resources map[string][]*apiv1.Resource) error {
	// If AppConfiguration does not contain monitoring config, return
	if p.app.Monitoring == nil {
		return nil
	}

	// Patch workspace configurations for monitoring generator.
	if err := p.parseWorkspaceConfig(); err != nil {
		return err
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

	if p.app.Monitoring.OperatorMode {
		monitoringLabels["kusion_monitoring_appname"] = p.app.Name
	} else {
		// If Prometheus doesn't run as an operator, kusion will generate the
		// most widely-known annotation for workloads that can be consumed by
		// the out-of-the-box community version of Prometheus server
		// installation shown as below. In this case, path and port cannot be
		// omitted
		if p.app.Monitoring.Path == "" || p.app.Monitoring.Port == "" {
			return monitoring.ErrPathAndPortEmpty
		}
		monitoringAnnotations["prometheus.io/scrape"] = "true"
		monitoringAnnotations["prometheus.io/scheme"] = p.app.Monitoring.Scheme
		monitoringAnnotations["prometheus.io/path"] = p.app.Monitoring.Path
		monitoringAnnotations["prometheus.io/port"] = p.app.Monitoring.Port
	}

	if err := modules.PatchResource(resources, modules.GVKDeployment, func(obj *appsv1.Deployment) error {
		obj.Labels = modules.MergeMaps(obj.Labels, monitoringLabels)
		obj.Annotations = modules.MergeMaps(obj.Annotations, monitoringAnnotations)
		obj.Spec.Template.Labels = modules.MergeMaps(obj.Spec.Template.Labels, monitoringLabels)
		obj.Spec.Template.Annotations = modules.MergeMaps(obj.Spec.Template.Annotations, monitoringAnnotations)
		return nil
	}); err != nil {
		return err
	}

	if err := modules.PatchResource(resources, modules.GVKDeployment, func(obj *v1alpha1.CollaSet) error {
		obj.Labels = modules.MergeMaps(obj.Labels, monitoringLabels)
		obj.Annotations = modules.MergeMaps(obj.Annotations, monitoringAnnotations)
		obj.Spec.Template.Labels = modules.MergeMaps(obj.Spec.Template.Labels, monitoringLabels)
		obj.Spec.Template.Annotations = modules.MergeMaps(obj.Spec.Template.Annotations, monitoringAnnotations)
		return nil
	}); err != nil {
		return err
	}

	if err := modules.PatchResource(resources, modules.GVKService, func(obj *corev1.Service) error {
		obj.Labels = modules.MergeMaps(obj.Labels, monitoringLabels)
		obj.Annotations = modules.MergeMaps(obj.Annotations, monitoringAnnotations)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// parseWorkspaceConfig parses the config items for monitoring generator in workspace configurations.
func (p *monitoringPatcher) parseWorkspaceConfig() error {
	wsConfig, ok := p.modulesConfig[monitoring.ModuleName]
	// If AppConfiguration contains monitoring config but workspace does not,
	// respond with the error ErrEmptyModuleConfigBlock
	if p.app.Monitoring != nil && !ok {
		return workspace.ErrEmptyModuleConfigBlock
	}

	if operatorMode, ok := wsConfig[monitoring.OperatorModeKey]; ok {
		p.app.Monitoring.OperatorMode = operatorMode.(bool)
	}

	if monitorType, ok := wsConfig[monitoring.MonitorTypeKey]; ok {
		p.app.Monitoring.MonitorType = monitoring.MonitorType(monitorType.(string))
	} else {
		p.app.Monitoring.MonitorType = monitoring.DefaultMonitorType
	}

	if interval, ok := wsConfig[monitoring.IntervalKey]; ok {
		p.app.Monitoring.Interval = prometheusv1.Duration(interval.(string))
	} else {
		p.app.Monitoring.Interval = monitoring.DefaultInterval
	}

	if timeout, ok := wsConfig[monitoring.TimeoutKey]; ok {
		p.app.Monitoring.Timeout = prometheusv1.Duration(timeout.(string))
	} else {
		p.app.Monitoring.Timeout = monitoring.DefaultTimeout
	}

	if scheme, ok := wsConfig[monitoring.SchemeKey]; ok {
		p.app.Monitoring.Scheme = scheme.(string)
	} else {
		p.app.Monitoring.Scheme = monitoring.DefaultScheme
	}

	parsedTimeout, err := time.ParseDuration(string(p.app.Monitoring.Timeout))
	if err != nil {
		return err
	}
	parsedInterval, err := time.ParseDuration(string(p.app.Monitoring.Interval))
	if err != nil {
		return err
	}

	if parsedTimeout > parsedInterval {
		return monitoring.ErrTimeoutGreaterThanInterval
	}

	return nil
}
