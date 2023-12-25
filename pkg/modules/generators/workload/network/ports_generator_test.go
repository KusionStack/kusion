package network

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/network"
)

func TestValidatePorts(t *testing.T) {
	type args struct {
		ports []network.Port
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid_ports",
			args: struct {
				ports []network.Port
			}{
				ports: []network.Port{
					{
						Port:     80,
						Protocol: "TCP",
					},
					{
						Port:     80,
						Protocol: "UDP",
					},
					{
						Port:       80,
						TargetPort: 8080,
						Protocol:   "TCP",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid_ports",
			args: struct {
				ports []network.Port
			}{
				ports: []network.Port{
					{
						Port:     80,
						Protocol: "TCP",
					},
					{
						Port:       9090,
						TargetPort: 8080,
						Protocol:   "UDP",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePorts(tt.args.ports); (err != nil) != tt.wantErr {
				t.Errorf("validatePorts() error = %x, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePortConfig(t *testing.T) {
	testcases := []struct {
		name       string
		portConfig apiv1.GenericConfig
		success    bool
	}{
		{
			name: "valid port config",
			portConfig: apiv1.GenericConfig{
				"type": "alicloud",
			},
			success: true,
		},
		{
			name:       "empty config",
			portConfig: nil,
			success:    true,
		},
		{
			name: "invalid port config unsupported type",
			portConfig: apiv1.GenericConfig{
				"type": "unsupported",
			},
			success: false,
		},
		{
			name: "invalid port config unsupported item",
			portConfig: apiv1.GenericConfig{
				"unsupported": "unsupported",
			},
			success: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePortConfig(tc.portConfig)
			assert.Equal(t, tc.success, err == nil)
		})
	}
}

func TestCompletePort(t *testing.T) {
	testcases := []struct {
		name          string
		port          *network.Port
		portConfig    apiv1.GenericConfig
		success       bool
		completedPort *network.Port
	}{
		{
			name: "complete target port",
			port: &network.Port{
				Port:     80,
				Protocol: "TCP",
			},
			portConfig: nil,
			success:    true,
			completedPort: &network.Port{
				Port:       80,
				TargetPort: 80,
				Protocol:   "TCP",
			},
		},
		{
			name: "complete type",
			port: &network.Port{
				Port:     80,
				Protocol: "TCP",
				Public:   true,
			},
			portConfig: apiv1.GenericConfig{
				"type": "alicloud",
			},
			success: true,
			completedPort: &network.Port{
				Type:       "alicloud",
				Port:       80,
				TargetPort: 80,
				Protocol:   "TCP",
				Public:     true,
			},
		},
		{
			name: "complete failed empty port config",
			port: &network.Port{
				Port:     80,
				Protocol: "TCP",
				Public:   true,
			},
			portConfig:    nil,
			success:       false,
			completedPort: nil,
		},
		{
			name: "complete failed type not exist",
			port: &network.Port{
				Port:     80,
				Protocol: "TCP",
				Public:   true,
			},
			portConfig:    apiv1.GenericConfig{},
			success:       false,
			completedPort: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := completePort(tc.port, tc.portConfig)
			assert.Equal(t, tc.success, err == nil)
			if tc.success {
				assert.True(t, reflect.DeepEqual(tc.completedPort, tc.port))
			}
		})
	}
}

func TestPortsGenerator_Generate(t *testing.T) {
	type fields struct {
		portsGenerator
	}
	type args struct {
		spec *apiv1.Intent
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ports_generate",
			fields: struct {
				portsGenerator
			}{
				portsGenerator{
					appName:     "testApp",
					projectName: "testProject",
					stackName:   "testStack",
					selector: map[string]string{
						"test-s-key": "test-s-value",
					},
					labels: map[string]string{
						"test-l-key": "test-l-value",
					},
					annotations: map[string]string{
						"test-a-key": "test-a-value",
					},
					ports: []network.Port{
						{
							Port:       80,
							TargetPort: 80,
							Protocol:   "TCP",
							Public:     true,
						},
						{
							Port:       9090,
							TargetPort: 8080,
							Protocol:   "UDP",
							Public:     false,
						},
					},
					portConfig: apiv1.GenericConfig{
						"type": "alicloud",
					},
				},
			},
			args: struct {
				spec *apiv1.Intent
			}{
				spec: &apiv1.Intent{},
			},
			wantErr: false,
		},
	}

	privateSvc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       k8sKindService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testProject-testStack-testApp-private",
			Namespace: "testProject",
			Labels: map[string]string{
				"test-l-key": "test-l-value",
			},
			Annotations: map[string]string{
				"test-a-key": "test-a-value",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "testProject-testStack-testApp-private-9090-udp",
					Port:       9090,
					TargetPort: intstr.FromInt(8080),
					Protocol:   v1.ProtocolUDP,
				},
			},
			Selector: map[string]string{
				"test-s-key": "test-s-value",
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}
	unstructuredPrivateSvc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(privateSvc)
	assert.NoError(t, err)

	publicSvc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       k8sKindService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testProject-testStack-testApp-public",
			Namespace: "testProject",
			Labels: map[string]string{
				"test-l-key":  "test-l-value",
				kusionControl: "true",
			},
			Annotations: map[string]string{
				"test-a-key": "test-a-value",
				aliyunLBSpec: aliyunSLBS1Small,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "testProject-testStack-testApp-public-80-tcp",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"test-s-key": "test-s-value",
			},
			Type: v1.ServiceTypeLoadBalancer,
		},
	}
	unstructuredPublicSvc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(publicSvc)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &tt.fields.portsGenerator
			_ = g.complete()
			if err = g.Generate(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, unstructuredPrivateSvc, tt.args.spec.Resources[0].Attributes)
			assert.Equal(t, unstructuredPublicSvc, tt.args.spec.Resources[1].Attributes)
		})
	}
}
