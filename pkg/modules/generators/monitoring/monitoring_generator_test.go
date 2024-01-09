package monitoring

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/monitoring"
)

type Fields struct {
	project *apiv1.Project
	stack   *apiv1.Stack
	app     *inputs.AppConfiguration
	ws      map[string]apiv1.GenericConfig
}

type Args struct {
	spec *apiv1.Intent
}

type TestCase struct {
	name    string
	fields  Fields
	args    Args
	want    *apiv1.Intent
	wantErr bool
}

func BuildMonitoringTestCase(
	testName, projectName, stackName, appName string,
	interval, timeout, path, port, scheme, monitorType string,
	operatorMode, wantErr bool,
) *TestCase {
	var endpointType string
	var monitorKind monitoring.MonitorType
	if monitorType == "Service" {
		monitorKind = "ServiceMonitor"
		endpointType = "endpoints"
	} else if monitorType == "Pod" {
		monitorKind = "PodMonitor"
		endpointType = "podMetricsEndpoints"
	}
	expectedResources := make([]apiv1.Resource, 0)
	uniqueName := modules.UniqueAppName(projectName, stackName, appName)
	if operatorMode {
		expectedResources = []apiv1.Resource{
			{
				ID:   fmt.Sprintf("monitoring.coreos.com/v1:%s:%s:%s-%s-monitor", monitorKind, projectName, uniqueName, strings.ToLower(monitorType)),
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "monitoring.coreos.com/v1",
					"kind":       string(monitorKind),
					"metadata": map[string]interface{}{
						"creationTimestamp": nil,
						"name":              fmt.Sprintf("%s-%s-monitor", uniqueName, strings.ToLower(monitorType)),
						"namespace":         projectName,
					},
					"spec": map[string]interface{}{
						endpointType: []interface{}{
							map[string]interface{}{
								"bearerTokenSecret": map[string]interface{}{
									"key": "",
								},
								"interval":      interval,
								"scrapeTimeout": timeout,
								"path":          path,
								"port":          port,
								"scheme":        scheme,
							},
						},
						"namespaceSelector": make(map[string]interface{}),
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								"kusion_monitoring_appname": appName,
							},
						},
					},
				},
				DependsOn: nil,
				Extensions: map[string]interface{}{
					"GVK": fmt.Sprintf("monitoring.coreos.com/v1, Kind=%s", string(monitorKind)),
				},
			},
		}
	}
	testCase := &TestCase{
		name: testName,
		fields: Fields{
			project: &apiv1.Project{
				Name: projectName,
			},
			stack: &apiv1.Stack{
				Name: stackName,
			},
			app: &inputs.AppConfiguration{
				Name: appName,
				Monitoring: &monitoring.Monitor{
					Path: path,
					Port: port,
				},
			},
			ws: map[string]apiv1.GenericConfig{
				"monitoring": {
					"operatorMode": operatorMode,
					"monitorType":  monitorType,
					"scheme":       scheme,
					"interval":     interval,
					"timeout":      timeout,
				},
			},
		},
		args: Args{
			spec: &apiv1.Intent{},
		},
		want: &apiv1.Intent{
			Resources: expectedResources,
		},
		wantErr: wantErr,
	}
	return testCase
}

func TestMonitoringGenerator_Generate(t *testing.T) {
	tests := []TestCase{
		*BuildMonitoringTestCase("ServiceMonitorTest", "test-project", "test-stack", "test-app", "15s", "5s", "/metrics", "web", "http", "Service", true, false),
		*BuildMonitoringTestCase("PodMonitorTest", "test-project", "test-stack", "test-app", "15s", "5s", "/metrics", "web", "http", "Pod", true, false),
		*BuildMonitoringTestCase("ServiceAnnotationTest", "test-project", "test-stack", "test-app", "30s", "15s", "/metrics", "8080", "http", "Service", false, false),
		*BuildMonitoringTestCase("PodAnnotationTest", "test-project", "test-stack", "test-app", "30s", "15s", "/metrics", "8080", "http", "Pod", false, false),
		*BuildMonitoringTestCase("InvalidDurationTest", "test-project", "test-stack", "test-app", "15s", "5ssss", "/metrics", "8080", "http", "Pod", false, true),
		*BuildMonitoringTestCase("InvalidTimeoutTest", "test-project", "test-stack", "test-app", "15s", "30s", "/metrics", "8080", "http", "Pod", false, true),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &monitoringGenerator{
				project:       tt.fields.project,
				stack:         tt.fields.stack,
				appName:       tt.fields.app.Name,
				app:           tt.fields.app,
				modulesConfig: tt.fields.ws,
				namespace:     tt.fields.project.Name,
			}
			if err := g.Generate(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				require.Equal(t, tt.want, tt.args.spec)
			}
		})
	}
}
