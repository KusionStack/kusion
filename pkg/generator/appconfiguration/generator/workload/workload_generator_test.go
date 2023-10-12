package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/network"
	"kusionstack.io/kusion/pkg/projectstack"
)

func TestNewWorkloadGenerator(t *testing.T) {
	t.Run("NewWorkloadGenerator should return a valid generator", func(t *testing.T) {
		expectedProject := &projectstack.Project{
			ProjectConfiguration: projectstack.ProjectConfiguration{
				Name: "test",
			},
		}
		expectedStack := &projectstack.Stack{}
		expectedWorkload := &workload.Workload{}
		expectedAppName := "test"

		actualGenerator, err := NewWorkloadGenerator(expectedProject, expectedStack, expectedAppName, expectedWorkload)

		assert.NoError(t, err, "Error should be nil")
		assert.NotNil(t, actualGenerator, "Generator should not be nil")
		assert.Equal(t, expectedProject, actualGenerator.(*workloadGenerator).project, "Project mismatch")
		assert.Equal(t, expectedStack, actualGenerator.(*workloadGenerator).stack, "Stack mismatch")
		assert.Equal(t, expectedAppName, actualGenerator.(*workloadGenerator).appName, "AppName mismatch")
		assert.Equal(t, expectedWorkload, actualGenerator.(*workloadGenerator).workload, "Workload mismatch")
	})
}

func TestNewWorkloadGeneratorFunc(t *testing.T) {
	t.Run("NewWorkloadGeneratorFunc should return a valid generator function", func(t *testing.T) {
		expectedProject := &projectstack.Project{
			ProjectConfiguration: projectstack.ProjectConfiguration{
				Name: "test",
			},
		}
		expectedStack := &projectstack.Stack{}
		expectedWorkload := &workload.Workload{}
		expectedAppName := "test"

		generatorFunc := NewWorkloadGeneratorFunc(expectedProject, expectedStack, expectedAppName, expectedWorkload)
		actualGenerator, err := generatorFunc()

		assert.NoError(t, err, "Error should be nil")
		assert.NotNil(t, actualGenerator, "Generator should not be nil")
		assert.Equal(t, expectedProject, actualGenerator.(*workloadGenerator).project, "Project mismatch")
		assert.Equal(t, expectedStack, actualGenerator.(*workloadGenerator).stack, "Stack mismatch")
		assert.Equal(t, expectedAppName, actualGenerator.(*workloadGenerator).appName, "AppName mismatch")
		assert.Equal(t, expectedWorkload, actualGenerator.(*workloadGenerator).workload, "Workload mismatch")
	})
}

func TestWorkloadGenerator_Generate(t *testing.T) {
	testCases := []struct {
		name             string
		expectedWorkload *workload.Workload
	}{
		{
			name: "Generate should generate the expected service",
			expectedWorkload: &workload.Workload{
				Header: workload.Header{
					Type: "Service",
				},
				Service: &workload.Service{
					Base: workload.Base{},
					Type: "Deployment",
					Ports: []network.Port{
						{
							Type:     network.CSPAliyun,
							Port:     80,
							Protocol: "TCP",
							Public:   true,
						},
					},
				},
			},
		},
		{
			name: "Generate should generate the expected job",
			expectedWorkload: &workload.Workload{
				Header: workload.Header{
					Type: "Job",
				},
				Job: &workload.Job{
					Base:     workload.Base{},
					Schedule: "* * * * *",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedProject := &projectstack.Project{
				ProjectConfiguration: projectstack.ProjectConfiguration{
					Name: "test",
					Prometheus: &projectstack.PrometheusConfig{
						OperatorMode: false,
						MonitorType:  "Pod",
					},
				},
			}
			expectedStack := &projectstack.Stack{
				StackConfiguration: projectstack.StackConfiguration{
					Name: "teststack",
				},
			}
			expectedAppName := "test"

			actualGenerator, _ := NewWorkloadGenerator(expectedProject, expectedStack, expectedAppName, tc.expectedWorkload)
			spec := &models.Spec{}
			err := actualGenerator.Generate(spec)
			assert.NoError(t, err, "Error should be nil")
			assert.NotNil(t, spec.Resources, "Resources should not be nil")

			// Check the generated resource
			resource := spec.Resources[0]
			actual := mapToUnstructured(resource.Attributes)

			assert.Equal(t, expectedProject.Name, actual.GetNamespace(), "Namespace mismatch")
			assert.Equal(t, appconfiguration.UniqueAppName(expectedProject.Name, expectedStack.Name, expectedAppName), actual.GetName(), "Name mismatch")

			if tc.expectedWorkload.Header.Type == "Service" {
				assert.Equal(t, "Deployment", actual.GetKind(), "Resource kind mismatch")
				assert.Equal(t, appconfiguration.MergeMaps(appconfiguration.UniqueAppLabels(expectedProject.Name, expectedAppName), tc.expectedWorkload.Service.Labels), actual.GetLabels(), "Labels mismatch")
			} else if tc.expectedWorkload.Header.Type == "Job" {
				assert.Equal(t, "CronJob", actual.GetKind(), "Resource kind mismatch")
				assert.Equal(t, appconfiguration.MergeMaps(appconfiguration.UniqueAppLabels(expectedProject.Name, expectedAppName), tc.expectedWorkload.Job.Labels), actual.GetLabels(), "Labels mismatch")
				assert.Equal(t, appconfiguration.MergeMaps(tc.expectedWorkload.Job.Annotations), actual.GetAnnotations(), "Annotations mismatch")
			}
		})
	}
}

func TestToOrderedContainers(t *testing.T) {
	t.Run("toOrderedContainers should convert app containers to ordered containers", func(t *testing.T) {
		appContainers := make(map[string]container.Container)
		appContainers["container1"] = container.Container{
			Image: "image1",
			Env:   make(yaml.MapSlice, 0),
		}
		appContainers["container2"] = container.Container{
			Image: "image2",
			Env: yaml.MapSlice{
				{
					Key:   "key",
					Value: "value",
				},
			},
		}
		appContainers["container3"] = container.Container{
			Image: "image3",
			Files: map[string]container.FileSpec{
				"/tmp/example1/file.txt": {
					Content: "some file contents",
					Mode:    "0777",
				},
				"/tmp/example2/file.txt": {
					Content: "some file contents",
					Mode:    "0644",
				},
			},
		}

		actualContainers, actualVolumes, actualConfigMaps, err := toOrderedContainers(appContainers, "mock-app-name")
		wantedConfigMapData := map[string]string{"file.txt": "some file contents"}

		assert.NoError(t, err, "Error should be nil")
		assert.Len(t, actualContainers, 3, "Number of containers mismatch")
		assert.Equal(t, "container1", actualContainers[0].Name, "Container name mismatch")
		assert.Equal(t, "image1", actualContainers[0].Image, "Container image mismatch")
		assert.Equal(t, "container2", actualContainers[1].Name, "Container name mismatch")
		assert.Equal(t, "image2", actualContainers[1].Image, "Container image mismatch")
		assert.Len(t, actualContainers[1].Env, 1, "Number of env vars mismatch")
		assert.Equal(t, "key", actualContainers[1].Env[0].Name, "Env var name mismatch")
		assert.Equal(t, "value", actualContainers[1].Env[0].Value, "Env var value mismatch")
		assert.Equal(t, "container3", actualContainers[2].Name, "Container name mismatch")
		assert.Equal(t, "image3", actualContainers[2].Image, "Container image mismatch")
		assert.Equal(t, "mock-app-name-container3-0", actualContainers[2].VolumeMounts[0].Name, "Container volumeMount name mismatch")
		assert.Equal(t, "/tmp/example1", actualContainers[2].VolumeMounts[0].MountPath, "Container volumeMount path mismatch")
		assert.Equal(t, "/tmp/example2", actualContainers[2].VolumeMounts[1].MountPath, "Container volumeMount path mismatch")
		assert.Equal(t, int32(511), *actualVolumes[0].ConfigMap.DefaultMode, "Volume mode mismatch")
		assert.Equal(t, int32(420), *actualVolumes[1].ConfigMap.DefaultMode, "Volume mode mismatch")
		assert.Equal(t, wantedConfigMapData, actualConfigMaps[0].Data, "ConfigMap data mismatch")
		assert.Equal(t, wantedConfigMapData, actualConfigMaps[1].Data, "ConfigMap data mismatch")
	})
	t.Run("toOrderedContainers should convert app containers with probe to ordered containers", func(t *testing.T) {
		appContainers := map[string]container.Container{
			"nginx": {
				Image: "nginx:v1",
				Resources: map[string]string{
					"cpu":    "2-4",
					"memory": "4Gi-8Gi",
				},
				LivenessProbe: &container.Probe{
					ProbeHandler: &container.ProbeHandler{
						TypeWrapper: container.TypeWrapper{
							Type: "Exec",
						},
						ExecAction: &container.ExecAction{
							Command: []string{"/bin/sh", "-c", "echo live"},
						},
					},
				},
				ReadinessProbe: &container.Probe{
					ProbeHandler: &container.ProbeHandler{
						TypeWrapper: container.TypeWrapper{
							Type: "Http",
						},
						HTTPGetAction: &container.HTTPGetAction{
							URL: "http://localhost:8080/readiness",
							Headers: map[string]string{
								"header": "value",
							},
						},
					},
					InitialDelaySeconds: 10,
				},
				StartupProbe: &container.Probe{
					ProbeHandler: &container.ProbeHandler{
						TypeWrapper: container.TypeWrapper{
							Type: "Tcp",
						},
						TCPSocketAction: &container.TCPSocketAction{
							URL: "10.0.0.1:8888",
						},
					},
				},
			},
		}

		actualContainers, _, _, err := toOrderedContainers(appContainers, "mock-app-name")

		assert.NoError(t, err, "Error should be nil")
		assert.Len(t, actualContainers, 1, "Number of containers mismatch")
		assert.Equal(t, "nginx", actualContainers[0].Name, "Container name mismatch")
		assert.Equal(t, "nginx:v1", actualContainers[0].Image, "Container image mismatch")
		assert.Len(t, actualContainers[0].Resources.Requests, 2, "Number of resource requests mismatch")

		// Assert request resources
		rQuantity := actualContainers[0].Resources.Requests["cpu"]
		assert.Equal(t, "2", (&rQuantity).String(), "CPU request mismatch")
		rQuantity = actualContainers[0].Resources.Requests["memory"]
		assert.Equal(t, "4Gi", (&rQuantity).String(), "CPU request mismatch")

		// Assert limit resources
		rQuantity = actualContainers[0].Resources.Limits["cpu"]
		assert.Equal(t, "4", (&rQuantity).String(), "CPU request mismatch")
		rQuantity = actualContainers[0].Resources.Limits["memory"]
		assert.Equal(t, "8Gi", (&rQuantity).String(), "CPU request mismatch")

		assert.NotNil(t, actualContainers[0].ReadinessProbe, "ReadinessProbe should not be nil")
		assert.NotNil(t, actualContainers[0].ReadinessProbe.HTTPGet, "ReadinessProbe.HTTPGet action should not be nil")
		assert.Equal(t, "HTTP", string(actualContainers[0].ReadinessProbe.HTTPGet.Scheme), "HTTPGet.Scheme mismatch")
		assert.Equal(t, "/readiness", actualContainers[0].ReadinessProbe.HTTPGet.Path, "HTTPGet.Path mismatch")
		assert.Equal(t, "8080", actualContainers[0].ReadinessProbe.HTTPGet.Port.String(), "HTTPGet.Port mismatch")
		assert.Equal(t, "", actualContainers[0].ReadinessProbe.HTTPGet.Host, "HTTPGet.Host mismatch")
		assert.Equal(t, 1, len(actualContainers[0].ReadinessProbe.HTTPGet.HTTPHeaders), "HTTPGet.HTTPHeaders length mismatch")

		assert.NotNil(t, actualContainers[0].LivenessProbe, "LivenessProbe should not be nil")
		assert.NotNil(t, actualContainers[0].LivenessProbe.Exec, "LivenessProbe.Exec action should not be nil")
		assert.Len(t, actualContainers[0].LivenessProbe.Exec.Command, 3, "LivenessProbe.Exec commands length mismatch")
		assert.Equal(t, []string{"/bin/sh", "-c", "echo live"}, actualContainers[0].LivenessProbe.Exec.Command, "LivenessProbe.Exec commands mismatch")

		assert.NotNil(t, actualContainers[0].StartupProbe, "StartupProbe should not be nil")
		assert.NotNil(t, actualContainers[0].StartupProbe.TCPSocket, "StartupProbe.TCPSocket action should not be nil")
		assert.Equal(t, "10.0.0.1", actualContainers[0].StartupProbe.TCPSocket.Host, "TCPSocket.Host mismatch")
		assert.Equal(t, "8888", actualContainers[0].StartupProbe.TCPSocket.Port.String(), "TCPSocket.Port mismatch")
	})
	t.Run("toOrderedContainers should convert app containers with lifecycle to ordered containers", func(t *testing.T) {
		appContainers := map[string]container.Container{
			"nginx": {
				Image: "nginx:v1",
				Lifecycle: &container.Lifecycle{
					PreStop: &container.LifecycleHandler{
						TypeWrapper: container.TypeWrapper{
							Type: "Exec",
						},
						ExecAction: &container.ExecAction{
							Command: []string{"/bin/sh", "-c", "echo live"},
						},
					},
					PostStart: &container.LifecycleHandler{
						TypeWrapper: container.TypeWrapper{
							Type: "Http",
						},
						HTTPGetAction: &container.HTTPGetAction{
							URL: "http://localhost:8080/readiness",
							Headers: map[string]string{
								"header": "value",
							},
						},
					},
				},
			},
		}

		actualContainers, _, _, err := toOrderedContainers(appContainers, "mock-app-name")

		assert.NoError(t, err, "Error should be nil")
		assert.Len(t, actualContainers, 1, "Number of containers mismatch")
		assert.Equal(t, "nginx", actualContainers[0].Name, "Container name mismatch")
		assert.Equal(t, "nginx:v1", actualContainers[0].Image, "Container image mismatch")

		assert.NotNil(t, actualContainers[0].Lifecycle, "Lifecycle should not be nil")
		assert.NotNil(t, actualContainers[0].Lifecycle.PreStop, "Lifecycle.PreStop should not be nil")
		assert.NotNil(t, actualContainers[0].Lifecycle.PreStop.Exec, "PreStop.Exec action should not be nil")
		assert.Len(t, actualContainers[0].Lifecycle.PreStop.Exec.Command, 3, "PreStop.Exec commands length mismatch")
		assert.Equal(t, []string{"/bin/sh", "-c", "echo live"}, actualContainers[0].Lifecycle.PreStop.Exec.Command, "PreStop.Exec commands mismatch")
		assert.NotNil(t, actualContainers[0].Lifecycle.PostStart, "Lifecycle.PostStart should not be nil")
		assert.Equal(t, "HTTP", string(actualContainers[0].Lifecycle.PostStart.HTTPGet.Scheme), "PostStart.HTTPGet.Scheme mismatch")
		assert.Equal(t, "/readiness", actualContainers[0].Lifecycle.PostStart.HTTPGet.Path, "PostStart.HTTPGet.Path mismatch")
		assert.Equal(t, "8080", actualContainers[0].Lifecycle.PostStart.HTTPGet.Port.String(), "PostStart.HTTPGet.Port mismatch")
		assert.Equal(t, "", actualContainers[0].Lifecycle.PostStart.HTTPGet.Host, "PostStart.HTTPGet.Host mismatch")
		assert.Equal(t, 1, len(actualContainers[0].Lifecycle.PostStart.HTTPGet.HTTPHeaders), "PostStart.HTTPGet.HTTPHeaders length mismatch")
	})
}
