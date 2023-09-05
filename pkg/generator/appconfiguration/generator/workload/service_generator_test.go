package workload

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/network"
	"kusionstack.io/kusion/pkg/projectstack"
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
	svc := `apiVersion: v1
kind: Service
metadata:
    annotations:
        service.beta.kubernetes.io/alibaba-cloud-loadbalancer-spec: slb.s1.small
    creationTimestamp: null
    labels:
        app.kubernetes.io/name: foo
        app.kubernetes.io/part-of: default
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
    creationTimestamp: null
    labels:
        app.kubernetes.io/name: foo
        app.kubernetes.io/part-of: default
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
            creationTimestamp: null
            labels:
                app.kubernetes.io/name: foo
                app.kubernetes.io/part-of: default
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
    name: default-dev-foo
    namespace: default
spec:
    replicas: 2
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
		want    []string
	}{
		{
			name: "CollaSet",
			fields: fields{
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
					Ports: []network.Port{
						{
							Port:     80,
							Protocol: "TCP",
							Public:   true,
						},
					},
				},
			},
			args: args{
				spec: &models.Spec{},
			},
			wantErr: false,
			want:    []string{cm, cs, svc},
		},
		{
			name: "Deployment",
			fields: fields{
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
					Type: "Deployment",
					Ports: []network.Port{
						{
							Port:     80,
							Protocol: "TCP",
							Public:   true,
						},
					},
				},
			},
			args: args{
				spec: &models.Spec{},
			},
			wantErr: false,
			want:    []string{cm, deploy, svc},
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
			for i := range tt.args.spec.Resources {
				b, err := yaml.Marshal(tt.args.spec.Resources[i].Attributes)
				require.NoError(t, err)
				require.Equal(t, tt.want[i], string(b))
			}
		})
	}
}
