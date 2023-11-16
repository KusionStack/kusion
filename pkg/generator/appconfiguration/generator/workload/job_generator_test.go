package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/projectstack"
)

func TestNewJobGenerator(t *testing.T) {
	expectedProject := &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name: "test",
		},
	}
	expectedStack := &projectstack.Stack{}
	expectedAppName := "test"
	expectedJob := &workload.Job{}
	actual, err := NewJobGenerator(expectedProject, expectedStack, expectedAppName, expectedJob)

	assert.NoError(t, err, "Error should be nil")
	assert.NotNil(t, actual, "Generator should not be nil")
	assert.Equal(t, expectedProject, actual.(*jobGenerator).project, "Project mismatch")
	assert.Equal(t, expectedStack, actual.(*jobGenerator).stack, "Stack mismatch")
	assert.Equal(t, expectedAppName, actual.(*jobGenerator).appName, "AppName mismatch")
	assert.Equal(t, expectedJob, actual.(*jobGenerator).job, "Job mismatch")
}

func TestNewJobGeneratorFunc(t *testing.T) {
	expectedProject := &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name: "test",
		},
	}
	expectedStack := &projectstack.Stack{}
	expectedAppName := "test"
	expectedJob := &workload.Job{}
	generatorFunc := NewJobGeneratorFunc(expectedProject, expectedStack, expectedAppName, expectedJob)
	actualGenerator, err := generatorFunc()

	assert.NoError(t, err, "Error should be nil")
	assert.NotNil(t, actualGenerator, "Generator should not be nil")
	assert.Equal(t, expectedProject, actualGenerator.(*jobGenerator).project, "Project mismatch")
	assert.Equal(t, expectedStack, actualGenerator.(*jobGenerator).stack, "Stack mismatch")
	assert.Equal(t, expectedAppName, actualGenerator.(*jobGenerator).appName, "AppName mismatch")
	assert.Equal(t, expectedJob, actualGenerator.(*jobGenerator).job, "Job mismatch")
}

func TestJobGenerator_Generate(t *testing.T) {
	testCases := []struct {
		name            string
		expectedProject *projectstack.Project
		expectedStack   *projectstack.Stack
		expectedAppName string
		expectedJob     *workload.Job
	}{
		{
			name: "test generate",
			expectedProject: &projectstack.Project{
				ProjectConfiguration: projectstack.ProjectConfiguration{
					Name: "test",
				},
			},
			expectedStack:   &projectstack.Stack{},
			expectedAppName: "test",
			expectedJob:     &workload.Job{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generator, _ := NewJobGenerator(tc.expectedProject, tc.expectedStack, tc.expectedAppName, tc.expectedJob)
			spec := &models.Intent{}
			err := generator.Generate(spec)

			assert.NoError(t, err, "Error should be nil")
			assert.NotNil(t, spec.Resources, "Resources should not be nil")
			assert.Len(t, spec.Resources, 1, "Number of resources mismatch")

			// Check the generated resource
			resource := spec.Resources[0]
			actual := mapToUnstructured(resource.Attributes)

			assert.Equal(t, "Job", actual.GetKind(), "Kind mismatch")
			assert.Equal(t, tc.expectedProject.Name, actual.GetNamespace(), "Namespace mismatch")
			assert.Equal(t, appconfiguration.UniqueAppName(tc.expectedProject.Name, tc.expectedStack.Name, tc.expectedAppName), actual.GetName(), "Name mismatch")
			assert.Equal(t, appconfiguration.MergeMaps(appconfiguration.UniqueAppLabels(tc.expectedProject.Name, tc.expectedAppName), tc.expectedJob.Labels), actual.GetLabels(), "Labels mismatch")
			assert.Equal(t, appconfiguration.MergeMaps(tc.expectedJob.Annotations), actual.GetAnnotations(), "Annotations mismatch")
		})
	}
}

func mapToUnstructured(data map[string]interface{}) *unstructured.Unstructured {
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetUnstructuredContent(data)
	return unstructuredObj
}
