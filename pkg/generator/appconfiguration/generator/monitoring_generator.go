package generator

import (
	"fmt"

	prometheusV1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/monitoring"
	"kusionstack.io/kusion/pkg/projectstack"
)

type monitoringGenerator struct {
	project *projectstack.Project
	monitor *monitoring.Monitor
	appName string
}

func NewMonitoringGenerator(project *projectstack.Project, monitor *monitoring.Monitor, appName string) (appconfiguration.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(appName) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}
	return &monitoringGenerator{
		project: project,
		monitor: monitor,
		appName: appName,
	}, nil
}

func NewMonitoringGeneratorFunc(project *projectstack.Project, monitor *monitoring.Monitor, appName string) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewMonitoringGenerator(project, monitor, appName)
	}
}

func (g *monitoringGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	monitoringLabels := map[string]string{
		"kusion_monitoring_appname": g.appName,
	}

	if g.monitor != nil && g.monitor.OperatorMode {
		if g.monitor.MonitorType == "service" {
			serviceEndpoint := prometheusV1.Endpoint{
				Interval:      g.monitor.Interval,
				ScrapeTimeout: g.monitor.Timeout,
				Port:          g.monitor.Port,
				Path:          g.monitor.Path,
				Scheme:        g.monitor.Scheme,
			}
			serviceEndpointList := []prometheusV1.Endpoint{serviceEndpoint}
			serviceMonitor := &prometheusV1.ServiceMonitor{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceMonitor",
					APIVersion: prometheusV1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-service-monitor", g.appName), Namespace: g.project.Name},
				Spec: prometheusV1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: monitoringLabels,
					},
					Endpoints: serviceEndpointList,
				},
			}
			err := appconfiguration.AppendToSpec(
				models.Kubernetes,
				appconfiguration.KubernetesResourceID(serviceMonitor.TypeMeta, serviceMonitor.ObjectMeta),
				spec,
				serviceMonitor,
			)
			if err != nil {
				return err
			}
		} else if g.monitor != nil && g.monitor.MonitorType == "pod" {
			podMetricsEndpoint := prometheusV1.PodMetricsEndpoint{
				Interval:      g.monitor.Interval,
				ScrapeTimeout: g.monitor.Timeout,
				Port:          g.monitor.Port,
				Path:          g.monitor.Path,
				Scheme:        g.monitor.Scheme,
			}
			podMetricsEndpointList := []prometheusV1.PodMetricsEndpoint{podMetricsEndpoint}

			podMonitor := &prometheusV1.PodMonitor{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PodMonitor",
					APIVersion: prometheusV1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-pod-monitor", g.appName), Namespace: g.project.Name},
				Spec: prometheusV1.PodMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: monitoringLabels,
					},
					PodMetricsEndpoints: podMetricsEndpointList,
				},
			}

			err := appconfiguration.AppendToSpec(
				models.Kubernetes,
				appconfiguration.KubernetesResourceID(podMonitor.TypeMeta, podMonitor.ObjectMeta),
				spec,
				podMonitor,
			)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("MonitorType should either be service or pod %s", g.monitor.MonitorType)
		}
	}

	return nil
}
