package generators

import (
	"context"
	"errors"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/core/v1/workload/network"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/proto"
	jsonutil "kusionstack.io/kusion/pkg/util/json"

	"kusionstack.io/kusion/pkg/apis/core/v1/workload"
)

type fakeModule struct{}

func (f *fakeModule) Generate(_ context.Context, _ *proto.GeneratorRequest) (*proto.GeneratorResponse, error) {
	res := v1.Resource{
		ID:   "apps.kusionstack.io/v1alpha1:PodTransitionRule:fakeNs:default-dev-foo",
		Type: "Kubernetes",
		Attributes: map[string]interface{}{
			"apiVersion": "apps.kusionstack.io/v1alpha1",
			"kind":       "PodTransitionRule",
			"metadata": map[string]interface{}{
				"creationTimestamp": interface{}(nil),
				"name":              "default-dev-foo",
				"namespace":         "fakeNs",
			},
			"spec": map[string]interface{}{
				"rules": []interface{}{map[string]interface{}{
					"availablePolicy": map[string]interface{}{
						"maxUnavailableValue": "30%",
					},
					"name": "maxUnavailable",
				}},
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app.kubernetes.io/name": "foo", "app.kubernetes.io/part-of": "default",
					},
				},
			}, "status": map[string]interface{}{},
		},
		DependsOn: []string(nil),
		Extensions: map[string]interface{}{
			"GVK": "apps.kusionstack.io/v1alpha1, Kind=PodTransitionRule",
		},
	}
	str := jsonutil.Marshal2String(res)
	b := []byte(str)
	return &proto.GeneratorResponse{
		Resources: [][]byte{b},
	}, nil
}

func mockPlugin() {
	mockey.Mock(modules.NewPlugin).To(func(key string) (*modules.Plugin, error) {
		return &modules.Plugin{Module: &fakeModule{}}, nil
	}).Build()
	mockey.Mock((*modules.Plugin).KillPluginClient).Return().Build()
}

func TestAppConfigurationGenerator_Generate_CustomNamespace(t *testing.T) {
	appName, app := buildMockApp()
	ws := buildMockWorkspace("fakeNs")

	g := &appConfigurationGenerator{
		project: "testproject",
		stack:   "test",
		appName: appName,
		app:     app,
		ws:      ws,
	}

	spec := &v1.Intent{
		Resources: []v1.Resource{},
	}

	mockPlugin()
	err := g.Generate(spec)
	assert.NoError(t, err)
	assert.NotEmpty(t, spec.Resources)

	// namespace name assertion
	for _, res := range spec.Resources {
		if res.Type != v1.Kubernetes {
			continue
		}
		actual := mapToUnstructured(res.Attributes)
		if actual.GetKind() == "Namespace" {
			assert.Equal(t, "fakeNs", actual.GetName(), "namespace name should be fakeNs")
		} else {
			ns := actual.GetNamespace()
			if ns == "" {
				// Manually get the namespace from the unstructured object.
				if ns, err = getNamespace(actual); err != nil {
					t.Fatal(err)
				}
			}
			assert.Equal(t, "fakeNs", ns, "namespace name should be fakeNs")
		}
	}
}

func TestNewAppConfigurationGeneratorFunc(t *testing.T) {
	appName, app := buildMockApp()
	ws := buildMockWorkspace("")

	t.Run("Valid app configuration generator func", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc("tesstproject", "test", appName, app, ws)()
		assert.NoError(t, err)
		assert.NotNil(t, g)
	})

	t.Run("Empty app name", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc("tesstproject", "test", "", app, ws)()
		assert.EqualError(t, err, "app name must not be empty")
		assert.Nil(t, g)
	})

	t.Run("Nil app", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc("tesstproject", "test", appName, nil, ws)()
		assert.EqualError(t, err, "can not find app configuration when generating the Intent")
		assert.Nil(t, g)
	})

	t.Run("Empty project name", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc("", "test", appName, app, ws)()
		assert.EqualError(t, err, "project name must not be empty")
		assert.Nil(t, g)
	})

	t.Run("Empty workspace", func(t *testing.T) {
		g, err := NewAppConfigurationGeneratorFunc("tesstproject", "test", appName, app, nil)()
		assert.EqualError(t, err, "workspace must not be empty")
		assert.Nil(t, g)
	})
}

func buildMockApp() (string, *v1.AppConfiguration) {
	return "app1", &v1.AppConfiguration{
		Workload: &workload.Workload{
			Header: workload.Header{
				Type: "Service",
			},
			Service: &workload.Service{
				Base: workload.Base{},
				Type: "Deployment",
				Ports: []network.Port{
					{
						Port:     80,
						Protocol: "TCP",
					},
				},
			},
		},
	}
}

func buildMockWorkspace(namespace string) *v1.Workspace {
	return &v1.Workspace{
		Name: "test",
		Modules: v1.ModuleConfigs{
			"kusionstack/database@v0.1": {
				Default: v1.GenericConfig{
					"type":         "aws",
					"version":      "5.7",
					"instanceType": "db.t3.micro",
				},
				ModulePatcherConfigs: v1.ModulePatcherConfigs{
					"smallClass": {
						GenericConfig: v1.GenericConfig{
							"instanceType": "db.t3.small",
						},
						ProjectSelector: []string{"foo", "bar"},
					},
				},
			},
			"port": {
				Default: v1.GenericConfig{
					"type": "aws",
				},
			},
			"namespace": {
				Default: v1.GenericConfig{
					"name": namespace,
				},
			},
		},
		Runtimes: &v1.RuntimeConfigs{
			Kubernetes: &v1.KubernetesConfig{
				KubeConfig: "/etc/kubeconfig.yaml",
			},
		},
	}
}

func mapToUnstructured(data map[string]interface{}) *unstructured.Unstructured {
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetUnstructuredContent(data)
	return unstructuredObj
}

func getNamespace(actual *unstructured.Unstructured) (string, error) {
	metadata, ok := actual.Object["metadata"].(map[interface{}]interface{})
	if !ok {
		return "", errors.New("failed to get metadata from unstructed object")
	}
	ns, ok := metadata["namespace"].(string)
	if !ok {
		return "", errors.New("failed to get namespace from metadata")
	}

	return ns, nil
}

func Test_patchWorkload(t *testing.T) {
	replica := int32(2)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-deployment",
			Labels: map[string]string{
				"oldLabel": "oldValue",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"oldLabel": "oldValue",
					},
					Annotations: map[string]string{
						"oldAnnotation": "oldValue",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "my-app",
							Image: "my-app-image",
							Env: []corev1.EnvVar{
								{
									Name:  "MY_ENV",
									Value: "my-env-value",
								},
							},
						},
					},
				},
			},
		},
	}
	// convert deploy to unstructured
	deploymentUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(deployment)
	res := &v1.Resource{
		ID:         "apps/v1:Deployment:default:default-dev-foo",
		Type:       "Kubernetes",
		Attributes: deploymentUnstructured,
	}

	t.Run("Patch labels and annotations", func(t *testing.T) {
		patcher := &v1.Patcher{
			Labels:      map[string]string{"newLabel": "newValue"},
			Annotations: map[string]string{"newAnnotation": "newValue"},
		}

		err := patchWorkload(res, patcher)
		assert.NoError(t, err)

		workloadLabels := res.Attributes["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
		podLabels := res.Attributes["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})

		// assert deployment workloadLabels
		assert.Equal(t, "newValue", workloadLabels["newLabel"])
		assert.Equal(t, "oldValue", workloadLabels["oldLabel"])
		// assert pod labels
		assert.Equal(t, "newValue", podLabels["newLabel"])
		assert.Equal(t, "oldValue", podLabels["oldLabel"])

		annotations := res.Attributes["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
		// get pod annotations
		podAnnotations := res.Attributes["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
		// assert deployment annotations
		assert.Equal(t, "newValue", annotations["newAnnotation"])
		// assert pod annotations
		assert.Equal(t, "newValue", podAnnotations["newAnnotation"])
		assert.Equal(t, "oldValue", podLabels["oldLabel"])
	})

	t.Run("Patch environment variables", func(t *testing.T) {
		patcher := &v1.Patcher{
			Environments: []corev1.EnvVar{
				{
					Name:  "NEW_ENV",
					Value: "my-new-value",
				},
			},
		}

		err = patchWorkload(res, patcher)
		assert.NoError(t, err)

		containers := res.Attributes["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})
		env := containers[0].(map[string]interface{})["env"].([]interface{})
		assert.Contains(t, env, map[string]interface{}{"name": "NEW_ENV", "value": "my-new-value"})
		assert.Contains(t, env, map[string]interface{}{"name": "MY_ENV", "value": "my-env-value"})
	})
}
