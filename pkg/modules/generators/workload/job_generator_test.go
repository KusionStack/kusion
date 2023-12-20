package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	workspaceapi "kusionstack.io/kusion/pkg/apis/workspace"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func TestNewJobGenerator(t *testing.T) {
	expectedProject := &project.Project{
		Configuration: project.Configuration{
			Name: "test",
		},
	}
	expectedStack := &stack.Stack{}
	expectedAppName := "test"
	expectedJob := &workload.Job{}
	expectedJobConfig := workspaceapi.GenericConfig{
		"labels": map[string]any{
			"workload-type": "Job",
		},
		"annotations": map[string]any{
			"workload-type": "Job",
		},
	}
	actual, err := NewJobGenerator(expectedProject, expectedStack, expectedAppName, expectedJob, expectedJobConfig)

	assert.NoError(t, err, "Error should be nil")
	assert.NotNil(t, actual, "Generator should not be nil")
	assert.Equal(t, expectedProject, actual.(*jobGenerator).project, "Project mismatch")
	assert.Equal(t, expectedStack, actual.(*jobGenerator).stack, "Stack mismatch")
	assert.Equal(t, expectedAppName, actual.(*jobGenerator).appName, "AppName mismatch")
	assert.Equal(t, expectedJob, actual.(*jobGenerator).job, "Job mismatch")
	assert.Equal(t, expectedJobConfig, actual.(*jobGenerator).jobConfig, "JobConfig mismatch")
}

func TestNewJobGeneratorFunc(t *testing.T) {
	expectedProject := &project.Project{
		Configuration: project.Configuration{
			Name: "test",
		},
	}
	expectedStack := &stack.Stack{}
	expectedAppName := "test"
	expectedJob := &workload.Job{}
	expectedJobConfig := workspaceapi.GenericConfig{
		"labels": map[string]any{
			"workload-type": "Job",
		},
		"annotations": map[string]any{
			"workload-type": "Job",
		},
	}
	generatorFunc := NewJobGeneratorFunc(expectedProject, expectedStack, expectedAppName, expectedJob, expectedJobConfig)
	actualGenerator, err := generatorFunc()

	assert.NoError(t, err, "Error should be nil")
	assert.NotNil(t, actualGenerator, "Generator should not be nil")
	assert.Equal(t, expectedProject, actualGenerator.(*jobGenerator).project, "Project mismatch")
	assert.Equal(t, expectedStack, actualGenerator.(*jobGenerator).stack, "Stack mismatch")
	assert.Equal(t, expectedAppName, actualGenerator.(*jobGenerator).appName, "AppName mismatch")
	assert.Equal(t, expectedJob, actualGenerator.(*jobGenerator).job, "Job mismatch")
	assert.Equal(t, expectedJobConfig, actualGenerator.(*jobGenerator).jobConfig, "JobConfig mismatch")
}

func TestJobGenerator_Generate(t *testing.T) {
	testCases := []struct {
		name              string
		expectedProject   *project.Project
		expectedStack     *stack.Stack
		expectedAppName   string
		expectedJob       *workload.Job
		expectedJobConfig workspaceapi.GenericConfig
	}{
		{
			name: "test generate",
			expectedProject: &project.Project{
				Configuration: project.Configuration{
					Name: "test",
				},
			},
			expectedStack:   &stack.Stack{},
			expectedAppName: "test",
			expectedJob:     &workload.Job{},
			expectedJobConfig: workspaceapi.GenericConfig{
				"labels": map[string]any{
					"workload-type": "Job",
				},
				"annotations": map[string]any{
					"workload-type": "Job",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generator, _ := NewJobGenerator(tc.expectedProject, tc.expectedStack, tc.expectedAppName, tc.expectedJob, tc.expectedJobConfig)
			spec := &intent.Intent{}
			err := generator.Generate(spec)

			assert.NoError(t, err, "Error should be nil")
			assert.NotNil(t, spec.Resources, "Resources should not be nil")
			assert.Len(t, spec.Resources, 1, "Number of resources mismatch")

			// Check the generated resource
			resource := spec.Resources[0]
			actual := mapToUnstructured(resource.Attributes)

			assert.Equal(t, "Job", actual.GetKind(), "Kind mismatch")
			assert.Equal(t, tc.expectedProject.Name, actual.GetNamespace(), "Namespace mismatch")
			assert.Equal(t, modules.UniqueAppName(tc.expectedProject.Name, tc.expectedStack.Name, tc.expectedAppName), actual.GetName(), "Name mismatch")
			assert.Equal(t, modules.MergeMaps(modules.UniqueAppLabels(tc.expectedProject.Name, tc.expectedAppName), tc.expectedJob.Labels), actual.GetLabels(), "Labels mismatch")
			assert.Equal(t, modules.MergeMaps(tc.expectedJob.Annotations), actual.GetAnnotations(), "Annotations mismatch")
		})
	}
}

func mapToUnstructured(data map[string]interface{}) *unstructured.Unstructured {
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetUnstructuredContent(data)
	return unstructuredObj
}
