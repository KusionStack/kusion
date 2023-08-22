package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
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
				},
			}
			expectedStack := &projectstack.Stack{}
			expectedAppName := "test"
			actualGenerator, _ := NewWorkloadGenerator(expectedProject, expectedStack, expectedAppName, tc.expectedWorkload)
			spec := &models.Spec{}
			err := actualGenerator.Generate(spec)
			assert.NoError(t, err, "Error should be nil")
			assert.NotNil(t, spec.Resources, "Resources should not be nil")
			assert.Len(t, spec.Resources, 1, "Number of resources mismatch")

			// Check the generated resource
			resource := spec.Resources[0]
			actual := mapToUnstructured(resource.Attributes)

			assert.Equal(t, expectedProject.Name, actual.GetNamespace(), "Namespace mismatch")
			assert.Equal(t, appconfiguration.UniqueAppName(expectedProject.Name, expectedStack.Name, expectedAppName), actual.GetName(), "Name mismatch")

			if tc.expectedWorkload.Header.Type == "Service" {
				assert.Equal(t, "Deployment", actual.GetKind(), "Resource kind mismatch")
				assert.Equal(t, appconfiguration.MergeMaps(appconfiguration.UniqueAppLabels(expectedProject.Name, expectedAppName), tc.expectedWorkload.Service.Labels), actual.GetLabels(), "Labels mismatch")
				assert.Equal(t, appconfiguration.MergeMaps(tc.expectedWorkload.Service.Annotations), actual.GetAnnotations(), "Annotations mismatch")
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
			Env:   make(map[string]string),
		}
		appContainers["container2"] = container.Container{
			Image: "image2",
			Env:   map[string]string{"key": "value"},
		}

		actualContainers, err := toOrderedContainers(appContainers)

		assert.NoError(t, err, "Error should be nil")
		assert.Len(t, actualContainers, 2, "Number of containers mismatch")
		assert.Equal(t, "container1", actualContainers[0].Name, "Container name mismatch")
		assert.Equal(t, "image1", actualContainers[0].Image, "Container image mismatch")
		assert.Equal(t, "container2", actualContainers[1].Name, "Container name mismatch")
		assert.Equal(t, "image2", actualContainers[1].Image, "Container image mismatch")
		assert.Len(t, actualContainers[1].Env, 1, "Number of env vars mismatch")
		assert.Equal(t, "key", actualContainers[1].Env[0].Name, "Env var name mismatch")
		assert.Equal(t, "value", actualContainers[1].Env[0].Value, "Env var value mismatch")
	})
}
