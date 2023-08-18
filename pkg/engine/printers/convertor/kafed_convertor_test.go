//go:build ignore
// +build ignore

// fixme
// ignore for test coverage temporary due to the strange coveralls test coverage computing rules
package convertor

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"kusionstack.io/kube-api/apps/v1alpha1"
)

func TestToKafed(t *testing.T) {
	csyaml := `apiVersion: apps.kusionstack.io/v1alpha1
kind: CollaSet
metadata:
  name: foo
  namespace: default
spec:
  selector:
    matchLabels:
      app: foo
  template:
    metadata:
      labels: 
        app: foo
    spec:
      containers:
        - name: foo
          image: nginx:v1 
`

	cs := &v1alpha1.CollaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CollaSet",
			APIVersion: "apps.kusionstack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
		Spec: v1alpha1.CollaSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "foo",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "foo",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "foo",
							Image: "nginx:v1",
						},
					},
				},
			},
		},
	}

	csmap := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(csyaml), csmap)
	assert.NoError(t, err, "unmarshall yaml to map error")

	type args struct {
		o *unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want runtime.Object
	}{
		{
			name: "CollaSet",
			args: args{
				o: &unstructured.Unstructured{Object: csmap},
			},
			want: cs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToKafed(tt.args.o); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToKafed() got = %v, want %v", got, tt.want)
			}
		})
	}
}
