package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func newGeneratorContextWithJob(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	job *workload.Job,
	jobConfig apiv1.GenericConfig,
) modules.GeneratorContext {
	application := &inputs.AppConfiguration{
		Name: appName,
		Workload: &workload.Workload{
			Job: job,
		},
	}
	moduleInputs := map[string]apiv1.GenericConfig{
		workload.ModuleJob: jobConfig,
	}
	return modules.GeneratorContext{
		Project:      project,
		Stack:        stack,
		Application:  application,
		Namespace:    project.Name,
		ModuleInputs: moduleInputs,
	}
}

func TestNewJobGenerator(t *testing.T) {
	expectedProject := &apiv1.Project{
		Name: "test",
	}
	expectedStack := &apiv1.Stack{}
	expectedAppName := "test"
	expectedJob := &workload.Job{}
	expectedJobConfig := apiv1.GenericConfig{
		"labels": map[string]any{
			"workload-type": "Job",
		},
		"annotations": map[string]any{
			"workload-type": "Job",
		},
	}
	ctx := newGeneratorContextWithJob(expectedProject, expectedStack, expectedAppName, expectedJob, expectedJobConfig)
	actual, err := NewJobGenerator(ctx)

	assert.NoError(t, err, "Error should be nil")
	assert.NotNil(t, actual, "Generator should not be nil")
	assert.Equal(t, expectedProject, actual.(*jobGenerator).project, "Project mismatch")
	assert.Equal(t, expectedStack, actual.(*jobGenerator).stack, "Stack mismatch")
	assert.Equal(t, expectedAppName, actual.(*jobGenerator).appName, "AppName mismatch")
	assert.Equal(t, expectedJob, actual.(*jobGenerator).job, "Job mismatch")
	assert.Equal(t, expectedJobConfig, actual.(*jobGenerator).jobConfig, "JobConfig mismatch")
}

func TestNewJobGeneratorFunc(t *testing.T) {
	expectedProject := &apiv1.Project{
		Name: "test",
	}
	expectedStack := &apiv1.Stack{}
	expectedAppName := "test"
	expectedJob := &workload.Job{}
	expectedJobConfig := apiv1.GenericConfig{
		"labels": map[string]any{
			"workload-type": "Job",
		},
		"annotations": map[string]any{
			"workload-type": "Job",
		},
	}
	ctx := newGeneratorContextWithJob(expectedProject, expectedStack, expectedAppName, expectedJob, expectedJobConfig)
	generatorFunc := NewJobGeneratorFunc(ctx)
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
		expectedProject   *apiv1.Project
		expectedStack     *apiv1.Stack
		expectedAppName   string
		expectedJob       *workload.Job
		expectedJobConfig apiv1.GenericConfig
	}{
		{
			name: "test generate",
			expectedProject: &apiv1.Project{
				Name: "test",
			},
			expectedStack:   &apiv1.Stack{},
			expectedAppName: "test",
			expectedJob:     &workload.Job{},
			expectedJobConfig: apiv1.GenericConfig{
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
			ctx := newGeneratorContextWithJob(tc.expectedProject, tc.expectedStack, tc.expectedAppName, tc.expectedJob, tc.expectedJobConfig)
			generator, _ := NewJobGenerator(ctx)
			spec := &apiv1.Intent{}
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
