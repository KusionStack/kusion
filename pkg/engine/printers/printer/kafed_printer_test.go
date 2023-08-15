package printer

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kube-api/apps/v1alpha1"
)

func Test_printCollaSet(t *testing.T) {
	replica := int32(2)

	csReady := &v1alpha1.CollaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CollaSet",
			APIVersion: "apps.kusionstack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
		Spec: v1alpha1.CollaSetSpec{
			Replicas: &replica,
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
		Status: v1alpha1.CollaSetStatus{
			UpdatedAvailableReplicas: 2,
		},
	}

	csUnReady := &v1alpha1.CollaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CollaSet",
			APIVersion: "apps.kusionstack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
		Spec: v1alpha1.CollaSetSpec{
			Replicas: &replica,
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
		Status: v1alpha1.CollaSetStatus{
			UpdatedAvailableReplicas: 1,
		},
	}

	type args struct {
		obj *v1alpha1.CollaSet
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "CollaSet Ready",
			args: args{obj: csReady},
			want: true,
		},
		{
			name: "CollaSet UnReady",
			args: args{obj: csUnReady},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := printCollaSet(tt.args.obj)
			if got != tt.want {
				t.Errorf("printCollaSet() got = %v, want %v", got, tt.want)
			}
		})
	}
}
