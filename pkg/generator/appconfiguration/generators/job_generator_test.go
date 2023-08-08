package generators

import (
	"testing"

	"github.com/stretchr/testify/require"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component/container"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component/workload"
)

func Test_jobGenerator_Generate(t *testing.T) {
	type fields struct {
		projectName string
		compName    string
		comp        *component.Component
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		in      *models.Spec
		want    *models.Spec
	}{
		{
			name: "test1",
			fields: fields{
				projectName: "proj1",
				compName:    "comp1",
				comp: &component.Component{
					Job: &workload.Job{
						Containers: map[string]container.Container{
							"container1": {
								Image: "nginx:v1",
							},
						},
					},
				},
			},
			in: &models.Spec{},
			want: &models.Spec{
				Resources: models.Resources{
					{
						ID:   "batch/v1:Job:proj1:proj1-comp1",
						Type: "Kubernetes",
						Attributes: map[string]interface{}{
							"apiVersion": "batch/v1",
							"kind":       "Job",
							"metadata": map[string]interface{}{
								"creationTimestamp": nil,
								"name":              "proj1-comp1",
								"namespace":         "proj1",
							},
							"spec": map[string]interface{}{
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
										"labels": map[string]interface{}{
											"app.kubernetes.io/component": "comp1",
											"app.kubernetes.io/name":      "proj1",
										},
									},
									"spec": map[string]interface{}{
										"containers": []interface{}{map[string]interface{}{
											"image":     "nginx:v1",
											"name":      "container1",
											"resources": map[string]interface{}{},
										}},
									},
								},
							},
							"status": map[string]interface{}{},
						},
						DependsOn:  nil,
						Extensions: nil,
					},
				},
			},
		},
		{
			name: "test2",
			fields: fields{
				projectName: "proj2",
				compName:    "comp2",
				comp: &component.Component{
					Job: &workload.Job{
						Containers: map[string]container.Container{
							"container1": {
								Image: "nginx:v1",
							},
						},
						Schedule: "* * * * *",
					},
				},
			},
			in: &models.Spec{},
			want: &models.Spec{Resources: models.Resources{
				{
					ID:   "batch/v1:CronJob:proj2:proj2-comp2",
					Type: "Kubernetes",
					Attributes: map[string]interface{}{
						"apiVersion": "batch/v1",
						"kind":       "CronJob",
						"metadata": map[string]interface{}{
							"creationTimestamp": nil,
							"name":              "proj2-comp2",
							"namespace":         "proj2",
						},
						"spec": map[string]interface{}{
							"jobTemplate": map[string]interface{}{
								"metadata": map[string]interface{}{
									"creationTimestamp": nil,
								},
								"spec": map[string]interface{}{
									"template": map[string]interface{}{
										"metadata": map[string]interface{}{
											"creationTimestamp": nil,
											"labels": map[string]interface{}{
												"app.kubernetes.io/component": "comp2",
												"app.kubernetes.io/name":      "proj2",
											},
										},
										"spec": map[string]interface{}{
											"containers": []interface{}{map[string]interface{}{
												"image":     "nginx:v1",
												"name":      "container1",
												"resources": map[string]interface{}{},
											}},
										},
									},
								},
							},
							"schedule": "* * * * *",
						},
						"status": map[string]interface{}{},
					},
					DependsOn:  nil,
					Extensions: nil,
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &jobGenerator{
				projectName: tt.fields.projectName,
				compName:    tt.fields.compName,
				comp:        tt.fields.comp,
			}
			err := g.Generate(tt.in)
			if err != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, tt.in)
			}
		})
	}
}
