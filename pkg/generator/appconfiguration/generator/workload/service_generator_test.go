package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kusionstack.io/kube-api/apps/v1alpha1"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
	"kusionstack.io/kusion/pkg/projectstack"
)

func Test_workloadServiceGenerator_Generate(t *testing.T) {
	replica := int32(2)
	mode := int32(511)
	cs := &v1alpha1.CollaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CollaSet",
			APIVersion: "apps.kusionstack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "default-dev-foo",
			Labels: map[string]string{
				"app.kubernetes.io/name":    "foo",
				"app.kubernetes.io/part-of": "default",
			},
		},
		Spec: v1alpha1.CollaSetSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":    "foo",
					"app.kubernetes.io/part-of": "default",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":    "foo",
						"app.kubernetes.io/part-of": "default",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:v1",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "default-dev-foo-nginx-0",
									MountPath: "/tmp",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "default-dev-foo-nginx-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "default-dev-foo-nginx-0",
									},
									DefaultMode: &mode,
								},
							},
						},
					},
				},
			},
		},
	}
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cs)
	assert.NoError(t, err)

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-dev-foo-nginx-0",
			Namespace: "default",
		},
		Data: map[string]string{
			"example.txt": "some file contents",
		},
	}
	unstructuredConfigMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cm)
	assert.NoError(t, err)

	type fields struct {
		project *projectstack.Project
		stack   *projectstack.Stack
		appName string
		service *workload.Service
	}
	type args struct {
		spec *models.Spec
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "CollaSet", fields: struct {
				project *projectstack.Project
				stack   *projectstack.Stack
				appName string
				service *workload.Service
			}{
				project: &projectstack.Project{
					ProjectConfiguration: projectstack.ProjectConfiguration{
						Name: "default",
					},
					Path: "/test",
				},
				stack: &projectstack.Stack{
					StackConfiguration: projectstack.StackConfiguration{Name: "dev"},
				},
				appName: "foo",
				service: &workload.Service{
					Base: workload.Base{
						Containers: map[string]container.Container{
							"nginx": {
								Image: "nginx:v1",
								Files: map[string]container.FileSpec{
									"/tmp/example.txt": {
										Content: "some file contents",
										Mode:    "0777",
									},
								},
							},
						},
						Replicas: 2,
					},
					Type: "CollaSet",
				},
			},
			args: struct {
				spec *models.Spec
			}{
				spec: &models.Spec{},
			}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &workloadServiceGenerator{
				project: tt.fields.project,
				stack:   tt.fields.stack,
				appName: tt.fields.appName,
				service: tt.fields.service,
			}
			if err := g.Generate(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, unstructuredConfigMap, tt.args.spec.Resources[0].Attributes)
			assert.Equal(t, unstructured, tt.args.spec.Resources[1].Attributes)
		})
	}
}
