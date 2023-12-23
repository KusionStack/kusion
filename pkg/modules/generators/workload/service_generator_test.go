package workload

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/container"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/network"
)

func Test_workloadServiceGenerator_Generate(t *testing.T) {
	cm := `apiVersion: v1
data:
    example.txt: some file contents
kind: ConfigMap
metadata:
    creationTimestamp: null
    name: default-dev-foo-nginx-0
    namespace: default
`
	csSvc := `apiVersion: v1
kind: Service
metadata:
    annotations:
        service-workload-type: CollaSet
        service.beta.kubernetes.io/alibaba-cloud-loadbalancer-spec: slb.s1.small
    creationTimestamp: null
    labels:
        app.kubernetes.io/name: foo
        app.kubernetes.io/part-of: default
        kusionstack.io/control: "true"
        service-workload-type: CollaSet
    name: default-dev-foo-public
    namespace: default
spec:
    ports:
        - name: default-dev-foo-public-80-tcp
          port: 80
          protocol: TCP
          targetPort: 80
    selector:
        app.kubernetes.io/name: foo
        app.kubernetes.io/part-of: default
    type: LoadBalancer
status:
    loadBalancer: {}
`

	deploySvc := `apiVersion: v1
kind: Service
metadata:
    annotations:
        service.beta.kubernetes.io/alibaba-cloud-loadbalancer-spec: slb.s1.small
    creationTimestamp: null
    labels:
        app.kubernetes.io/name: foo
        app.kubernetes.io/part-of: default
        kusionstack.io/control: "true"
        service-workload-type: Deployment
    name: default-dev-foo-public
    namespace: default
spec:
    ports:
        - name: default-dev-foo-public-80-tcp
          port: 80
          protocol: TCP
          targetPort: 80
    selector:
        app.kubernetes.io/name: foo
        app.kubernetes.io/part-of: default
    type: LoadBalancer
status:
    loadBalancer: {}
`
	cs := `apiVersion: apps.kusionstack.io/v1alpha1
kind: CollaSet
metadata:
    annotations:
        service-workload-type: CollaSet
    creationTimestamp: null
    labels:
        app.kubernetes.io/name: foo
        app.kubernetes.io/part-of: default
        service-workload-type: CollaSet
    name: default-dev-foo
    namespace: default
spec:
    replicas: 2
    scaleStrategy: {}
    selector:
        matchLabels:
            app.kubernetes.io/name: foo
            app.kubernetes.io/part-of: default
    template:
        metadata:
            annotations:
                service-workload-type: CollaSet
            creationTimestamp: null
            labels:
                app.kubernetes.io/name: foo
                app.kubernetes.io/part-of: default
                service-workload-type: CollaSet
        spec:
            containers:
                - image: nginx:v1
                  name: nginx
                  resources: {}
                  volumeMounts:
                    - mountPath: /tmp
                      name: default-dev-foo-nginx-0
            volumes:
                - configMap:
                    defaultMode: 511
                    name: default-dev-foo-nginx-0
                  name: default-dev-foo-nginx-0
    updateStrategy: {}
status: {}
`
	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
    creationTimestamp: null
    labels:
        app.kubernetes.io/name: foo
        app.kubernetes.io/part-of: default
        service-workload-type: Deployment
    name: default-dev-foo
    namespace: default
spec:
    replicas: 4
    selector:
        matchLabels:
            app.kubernetes.io/name: foo
            app.kubernetes.io/part-of: default
    strategy: {}
    template:
        metadata:
            creationTimestamp: null
            labels:
                app.kubernetes.io/name: foo
                app.kubernetes.io/part-of: default
                service-workload-type: Deployment
        spec:
            containers:
                - image: nginx:v1
                  name: nginx
                  resources: {}
                  volumeMounts:
                    - mountPath: /tmp
                      name: default-dev-foo-nginx-0
            volumes:
                - configMap:
                    defaultMode: 511
                    name: default-dev-foo-nginx-0
                  name: default-dev-foo-nginx-0
status: {}
`
	type fields struct {
		project       *apiv1.Project
		stack         *apiv1.Stack
		appName       string
		service       *workload.Service
		serviceConfig apiv1.GenericConfig
	}
	type args struct {
		spec *apiv1.Intent
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    []string
	}{
		{
			name: "CollaSet",
			fields: fields{
				project: &apiv1.Project{
					Name: "default",
					Path: "/test",
				},
				stack: &apiv1.Stack{
					Name: "dev",
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
					Ports: []network.Port{
						{
							Type:     network.CSPAliyun,
							Port:     80,
							Protocol: "TCP",
							Public:   true,
						},
					},
				},
				serviceConfig: apiv1.GenericConfig{
					"type": "CollaSet",
					"labels": map[string]any{
						"service-workload-type": "CollaSet",
					},
					"annotations": map[string]any{
						"service-workload-type": "CollaSet",
					},
				},
			},
			args: args{
				spec: &apiv1.Intent{},
			},
			wantErr: false,
			want:    []string{cm, cs, csSvc},
		},
		{
			name: "Deployment",
			fields: fields{
				project: &apiv1.Project{
					Name: "default",
					Path: "/test",
				},
				stack: &apiv1.Stack{
					Name: "dev",
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
					},
					Ports: []network.Port{
						{
							Type:     network.CSPAliyun,
							Port:     80,
							Protocol: "TCP",
							Public:   true,
						},
					},
				},
				serviceConfig: apiv1.GenericConfig{
					"replicas": 4,
					"labels": map[string]any{
						"service-workload-type": "Deployment",
					},
				},
			},
			args: args{
				spec: &apiv1.Intent{},
			},
			wantErr: false,
			want:    []string{cm, deploy, deploySvc},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := modules.GeneratorContext{
				Project: tt.fields.project,
				Stack:   tt.fields.stack,
				Application: &inputs.AppConfiguration{
					Name: tt.fields.appName,
					Workload: &workload.Workload{
						Service: tt.fields.service,
					},
				},
				Namespace: tt.fields.project.Name,
			}
			g := &workloadServiceGenerator{
				project:       tt.fields.project,
				stack:         tt.fields.stack,
				appName:       tt.fields.appName,
				service:       tt.fields.service,
				serviceConfig: tt.fields.serviceConfig,
				namespace:     tt.fields.project.Name,
				context:       ctx,
			}
			if err := g.Generate(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			for i := range tt.args.spec.Resources {
				b, err := yaml.Marshal(tt.args.spec.Resources[i].Attributes)
				require.NoError(t, err)
				require.Equal(t, tt.want[i], string(b))
			}
		})
	}
}

func TestCompleteServiceInput(t *testing.T) {
	testcases := []struct {
		name             string
		service          *workload.Service
		config           apiv1.GenericConfig
		success          bool
		completedService *workload.Service
	}{
		{
			name: "use type in workspace config",
			service: &workload.Service{
				Base: workload.Base{
					Containers: map[string]container.Container{
						"nginx": {
							Image: "nginx:v1",
						},
					},
					Replicas: 2,
					Labels: map[string]string{
						"k1": "v1",
					},
					Annotations: map[string]string{
						"k1": "v1",
					},
				},
			},
			config: apiv1.GenericConfig{
				"type": "CollaSet",
			},
			success: true,
			completedService: &workload.Service{
				Base: workload.Base{
					Containers: map[string]container.Container{
						"nginx": {
							Image: "nginx:v1",
						},
					},
					Replicas: 2,
					Labels: map[string]string{
						"k1": "v1",
					},
					Annotations: map[string]string{
						"k1": "v1",
					},
				},
				Type: "CollaSet",
			},
		},
		{
			name: "use default type",
			service: &workload.Service{
				Base: workload.Base{
					Containers: map[string]container.Container{
						"nginx": {
							Image: "nginx:v1",
						},
					},
					Replicas: 2,
					Labels: map[string]string{
						"k1": "v1",
					},
					Annotations: map[string]string{
						"k1": "v1",
					},
				},
			},
			config:  nil,
			success: true,
			completedService: &workload.Service{
				Base: workload.Base{
					Containers: map[string]container.Container{
						"nginx": {
							Image: "nginx:v1",
						},
					},
					Replicas: 2,
					Labels: map[string]string{
						"k1": "v1",
					},
					Annotations: map[string]string{
						"k1": "v1",
					},
				},
				Type: "Deployment",
			},
		},
		{
			name: "invalid field type",
			service: &workload.Service{
				Base: workload.Base{
					Containers: map[string]container.Container{
						"nginx": {
							Image: "nginx:v1",
						},
					},
					Replicas: 2,
					Labels: map[string]string{
						"k1": "v1",
					},
					Annotations: map[string]string{
						"k1": "v1",
					},
				},
			},
			config: apiv1.GenericConfig{
				"type": 1,
			},
			success:          false,
			completedService: nil,
		},
		{
			name: "unsupported type",
			service: &workload.Service{
				Base: workload.Base{
					Containers: map[string]container.Container{
						"nginx": {
							Image: "nginx:v1",
						},
					},
					Replicas: 2,
					Labels: map[string]string{
						"k1": "v1",
					},
					Annotations: map[string]string{
						"k1": "v1",
					},
				},
			},
			config: apiv1.GenericConfig{
				"type": "unsupported",
			},
			success:          false,
			completedService: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := completeServiceInput(tc.service, tc.config)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.True(t, reflect.DeepEqual(tc.service, tc.completedService))
			}
		})
	}
}
