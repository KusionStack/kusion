package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/modules"
)

func TestNewJobGenerator(t *testing.T) {
	expectedProject := "test"
	expectedStack := "dev"
	expectedAppName := "test"
	expectedJob := &v1.Job{}
	expectedJobConfig := v1.GenericConfig{
		"labels": v1.GenericConfig{
			"Workload-type": "Job",
		},
		"annotations": v1.GenericConfig{
			"Workload-type": "Job",
		},
	}
	actual, err := NewJobGenerator(&Generator{
		Project:   expectedProject,
		Stack:     expectedStack,
		App:       expectedAppName,
		Namespace: expectedAppName,
		Workload: &v1.Workload{
			Job: expectedJob,
		},
		PlatformConfigs: map[string]v1.GenericConfig{
			v1.ModuleJob: expectedJobConfig,
		},
	})

	assert.NoError(t, err, "Error should be nil")
	assert.NotNil(t, actual, "Generator should not be nil")
	assert.Equal(t, expectedProject, actual.(*jobGenerator).project, "Project mismatch")
	assert.Equal(t, expectedStack, actual.(*jobGenerator).stack, "Stack mismatch")
	assert.Equal(t, expectedAppName, actual.(*jobGenerator).appName, "AppName mismatch")
	assert.Equal(t, expectedJob, actual.(*jobGenerator).job, "Job mismatch")
	assert.Equal(t, expectedJobConfig, actual.(*jobGenerator).jobConfig, "JobConfig mismatch")
}

func TestNewJobGeneratorFunc(t *testing.T) {
	expectedProject := "test"
	expectedStack := "dev"
	expectedAppName := "test"
	expectedJob := &v1.Job{}
	expectedJobConfig := v1.GenericConfig{
		"labels": v1.GenericConfig{
			"workload-type": "Job",
		},
		"annotations": v1.GenericConfig{
			"workload-type": "Job",
		},
	}
	generatorFunc := NewJobGeneratorFunc(&Generator{
		Project:   expectedProject,
		Stack:     expectedStack,
		App:       expectedAppName,
		Namespace: expectedAppName,
		Workload: &v1.Workload{
			Job: expectedJob,
		},
		PlatformConfigs: map[string]v1.GenericConfig{
			v1.ModuleJob: expectedJobConfig,
		},
	})
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
		expectedProject   string
		expectedStack     string
		expectedAppName   string
		expectedJob       *v1.Job
		expectedJobConfig v1.GenericConfig
	}{
		{
			name:            "test generate",
			expectedProject: "test",
			expectedStack:   "dev",
			expectedAppName: "test",
			expectedJob:     &v1.Job{},
			expectedJobConfig: v1.GenericConfig{
				"labels": v1.GenericConfig{
					"workload-type": "Job",
				},
				"annotations": v1.GenericConfig{
					"workload-type": "Job",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generator, _ := NewJobGenerator(&Generator{
				Project:   tc.expectedProject,
				Stack:     tc.expectedStack,
				App:       tc.expectedAppName,
				Namespace: tc.expectedAppName,
				Workload: &v1.Workload{
					Job: tc.expectedJob,
				},
				PlatformConfigs: map[string]v1.GenericConfig{
					v1.ModuleJob: tc.expectedJobConfig,
				},
			})
			spec := &v1.Spec{}
			err := generator.Generate(spec)

			assert.NoError(t, err, "Error should be nil")
			assert.NotNil(t, spec.Resources, "Resources should not be nil")
			assert.Len(t, spec.Resources, 1, "Number of resources mismatch")

			// Check the generated resource
			resource := spec.Resources[0]
			actual := mapToUnstructured(resource.Attributes)

			assert.Equal(t, "Job", actual.GetKind(), "Kind mismatch")
			assert.Equal(t, tc.expectedProject, actual.GetNamespace(), "Namespace mismatch")
			assert.Equal(t, modules.UniqueAppName(tc.expectedProject, tc.expectedStack, tc.expectedAppName), actual.GetName(), "Name mismatch")
			assert.Equal(t, modules.MergeMaps(modules.UniqueAppLabels(tc.expectedProject, tc.expectedAppName), tc.expectedJob.Labels), actual.GetLabels(), "Labels mismatch")
			assert.Equal(t, modules.MergeMaps(tc.expectedJob.Annotations), actual.GetAnnotations(), "Annotations mismatch")
		})
	}
}

func mapToUnstructured(data map[string]interface{}) *unstructured.Unstructured {
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetUnstructuredContent(data)
	return unstructuredObj
}
