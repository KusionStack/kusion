package monitoring

import (
	"fmt"
	"time"

	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kusionstack.io/kusion/pkg/modules/inputs"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs/monitoring"
	"kusionstack.io/kusion/pkg/workspace"
)

type monitoringGenerator struct {
	project       *apiv1.Project
	stack         *apiv1.Stack
	appName       string
	app           *inputs.AppConfiguration
	modulesConfig map[string]apiv1.GenericConfig
	namespace     string
}

func NewMonitoringGenerator(ctx modules.GeneratorContext) (modules.Generator, error) {
	if len(ctx.Project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(ctx.Application.Name) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}
	return &monitoringGenerator{
		project:       ctx.Project,
		stack:         ctx.Stack,
		app:           ctx.Application,
		appName:       ctx.Application.Name,
		modulesConfig: ctx.ModuleInputs,
		namespace:     ctx.Namespace,
	}, nil
}

func NewMonitoringGeneratorFunc(ctx modules.GeneratorContext) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewMonitoringGenerator(ctx)
	}
}

func (g *monitoringGenerator) Generate(spec *apiv1.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(apiv1.Resources, 0)
	}
	// If AppConfiguration does not contain monitoring config, return
	if g.app.Monitoring == nil {
		return nil
	}

	// Patch workspace configurations for monitoring generator.
	if err := g.parseWorkspaceConfig(); err != nil {
		return err
	}

	if g.app.Monitoring != nil && g.app.Monitoring.OperatorMode {
		if g.app.Monitoring.MonitorType == monitoring.ServiceMonitorType {
			serviceMonitor, err := g.buildMonitorObject(g.app.Monitoring.MonitorType)
			if err != nil {
				return err
			}
			err = modules.AppendToIntent(
				apiv1.Kubernetes,
				modules.KubernetesResourceID(
					serviceMonitor.(*prometheusv1.ServiceMonitor).TypeMeta,
					serviceMonitor.(*prometheusv1.ServiceMonitor).ObjectMeta,
				),
				spec,
				serviceMonitor,
			)
			if err != nil {
				return err
			}
		} else if g.app.Monitoring.MonitorType == monitoring.PodMonitorType {
			podMonitor, err := g.buildMonitorObject(g.app.Monitoring.MonitorType)
			if err != nil {
				return err
			}
			err = modules.AppendToIntent(
				apiv1.Kubernetes,
				modules.KubernetesResourceID(
					podMonitor.(*prometheusv1.PodMonitor).TypeMeta,
					podMonitor.(*prometheusv1.PodMonitor).ObjectMeta,
				),
				spec,
				podMonitor,
			)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("MonitorType should either be service or pod %s", g.app.Monitoring.MonitorType)
		}
	}

	return nil
}

// parseWorkspaceConfig parses the config items for monitoring generator in workspace configurations.
func (g *monitoringGenerator) parseWorkspaceConfig() error {
	wsConfig, ok := g.modulesConfig[monitoring.ModuleName]
	// If AppConfiguration contains monitoring config but workspace does not,
	// respond with the error ErrEmptyModuleConfigBlock
	if g.app.Monitoring != nil && !ok {
		return workspace.ErrEmptyModuleConfigBlock
	}

	if operatorMode, ok := wsConfig[monitoring.OperatorModeKey]; ok {
		g.app.Monitoring.OperatorMode = operatorMode.(bool)
	}

	if monitorType, ok := wsConfig[monitoring.MonitorTypeKey]; ok {
		g.app.Monitoring.MonitorType = monitoring.MonitorType(monitorType.(string))
	} else {
		g.app.Monitoring.MonitorType = monitoring.DefaultMonitorType
	}

	if interval, ok := wsConfig[monitoring.IntervalKey]; ok {
		g.app.Monitoring.Interval = prometheusv1.Duration(interval.(string))
	} else {
		g.app.Monitoring.Interval = monitoring.DefaultInterval
	}

	if timeout, ok := wsConfig[monitoring.TimeoutKey]; ok {
		g.app.Monitoring.Timeout = prometheusv1.Duration(timeout.(string))
	} else {
		g.app.Monitoring.Timeout = monitoring.DefaultTimeout
	}

	if scheme, ok := wsConfig[monitoring.SchemeKey]; ok {
		g.app.Monitoring.Scheme = scheme.(string)
	} else {
		g.app.Monitoring.Scheme = monitoring.DefaultScheme
	}

	parsedTimeout, err := time.ParseDuration(string(g.app.Monitoring.Timeout))
	if err != nil {
		return err
	}
	parsedInterval, err := time.ParseDuration(string(g.app.Monitoring.Interval))
	if err != nil {
		return err
	}

	if parsedTimeout > parsedInterval {
		return monitoring.ErrTimeoutGreaterThanInterval
	}

	return nil
}

func (g *monitoringGenerator) buildMonitorObject(monitorType monitoring.MonitorType) (runtime.Object, error) {
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
	monitoringLabels := map[string]string{
		"kusion_monitoring_appname": g.appName,
	}

	if monitorType == "Service" {
		serviceEndpoint := prometheusv1.Endpoint{
			Interval:      g.app.Monitoring.Interval,
			ScrapeTimeout: g.app.Monitoring.Timeout,
			Port:          g.app.Monitoring.Port,
			Path:          g.app.Monitoring.Path,
			Scheme:        g.app.Monitoring.Scheme,
		}
		serviceEndpointList := []prometheusv1.Endpoint{serviceEndpoint}
		serviceMonitor := &prometheusv1.ServiceMonitor{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceMonitor",
				APIVersion: prometheusv1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-service-monitor", modules.UniqueAppName(g.project.Name, g.stack.Name, g.appName)),
				Namespace: g.namespace,
			},
			Spec: prometheusv1.ServiceMonitorSpec{
				Selector: metav1.LabelSelector{
					MatchLabels: monitoringLabels,
				},
				Endpoints: serviceEndpointList,
			},
		}
		return serviceMonitor, nil
	} else if monitorType == "Pod" {
		podMetricsEndpoint := prometheusv1.PodMetricsEndpoint{
			Interval:      g.app.Monitoring.Interval,
			ScrapeTimeout: g.app.Monitoring.Timeout,
			Port:          g.app.Monitoring.Port,
			Path:          g.app.Monitoring.Path,
			Scheme:        g.app.Monitoring.Scheme,
		}
		podMetricsEndpointList := []prometheusv1.PodMetricsEndpoint{podMetricsEndpoint}

		podMonitor := &prometheusv1.PodMonitor{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodMonitor",
				APIVersion: prometheusv1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-pod-monitor", modules.UniqueAppName(g.project.Name, g.stack.Name, g.appName)),
				Namespace: g.namespace,
			},
			Spec: prometheusv1.PodMonitorSpec{
				Selector: metav1.LabelSelector{
					MatchLabels: monitoringLabels,
				},
				PodMetricsEndpoints: podMetricsEndpointList,
			},
		}
		return podMonitor, nil
	}

	return nil, fmt.Errorf("MonitorType should either be service or pod %s", monitorType)
}
