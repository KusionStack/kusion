package workload

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/apis/core/v1/workload/container"
	"kusionstack.io/kusion/pkg/apis/core/v1/workload/network"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/core/v1/workload"
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
	r2 := new(int32)
	*r2 = 2

	type fields struct {
		project       string
		stack         string
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
				project: "default",
				stack:   "dev",
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
						Replicas: r2,
					},
					Ports: []network.Port{
						{
							Port:     80,
							Protocol: "TCP",
						},
					},
				},
				serviceConfig: apiv1.GenericConfig{
					"type": "CollaSet",
					"labels": apiv1.GenericConfig{
						"service-workload-type": "CollaSet",
					},
					"annotations": apiv1.GenericConfig{
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
				project: "default",
				stack:   "dev",
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
							Port:     80,
							Protocol: "TCP",
						},
					},
				},
				serviceConfig: apiv1.GenericConfig{
					"replicas": 4,
					"labels": apiv1.GenericConfig{
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
			g := &ServiceGenerator{
				Project:   tt.fields.project,
				Stack:     tt.fields.stack,
				App:       tt.fields.appName,
				Service:   tt.fields.service,
				Config:    tt.fields.serviceConfig,
				Namespace: tt.fields.project,
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
	r2 := int32(2)

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
					Replicas: &r2,
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
					Replicas: &r2,
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
					Replicas: &r2,
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
					Replicas: &r2,
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
					Replicas: &r2,
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
					Replicas: &r2,
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
