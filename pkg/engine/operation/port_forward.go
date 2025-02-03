package operation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	yamlv2 "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/printers/convertor"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes/kubeops"
)

var (
	ErrEmptySpec               = errors.New("empty resources in spec")
	ErrEmptyService            = errors.New("empty k8s service")
	ErrNotOneSvcWithTargetPort = errors.New("only support one k8s service to forward with target port")
)

type PortForwardOperation struct {
	models.Operation
}

type PortForwardRequest struct {
	models.Request
	Spec *v1.Spec
	Port int
}

func (bpo *PortForwardOperation) PortForward(req *PortForwardRequest, stopChan chan struct{}, readyChan chan struct{}) error {
	ctx := context.Background()
	if err := validatePortForwardRequest(req); err != nil {
		return err
	}

	// Find Kubernetes Service in the resources of Spec.
	services := make(map[*v1.Resource]*corev1.Service)
	for _, res := range req.Spec.Resources {
		// Skip non-Kubernetes resources.
		if res.Type != v1.Kubernetes {
			continue
		}

		// Convert interface{} to unstructured.
		rYaml, err := yamlv2.Marshal(res.Attributes)
		if err != nil {
			return fmt.Errorf("failed to convert resource attributes to unstructured raw yaml: %v", err)
		}

		// Decode YAML manifest into unstructured.Unstructured.
		decUnstructured := k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &unstructured.Unstructured{}

		_, _, err = decUnstructured.Decode(rYaml, nil, obj)
		if err != nil {
			return fmt.Errorf("failed to decode yaml manifest into unstructured object: %v", err)
		}

		if obj.GetKind() != convertor.Service {
			continue
		}

		convertedObj := convertor.ToK8s(obj)
		services[&res] = convertedObj.(*corev1.Service)
	}

	if len(services) == 0 {
		return ErrEmptyService
	}

	filteredServices := make(map[*v1.Resource]*corev1.Service)
	for res, svc := range services {
		targetPortFound := false
		for _, port := range svc.Spec.Ports {
			if port.Port == int32(req.Port) {
				targetPortFound = true
				continue
			}
		}

		if targetPortFound {
			filteredServices[res] = svc
		}
	}
	services = filteredServices

	if len(services) != 1 {
		return ErrNotOneSvcWithTargetPort
	}

	// Port-forward the Service with client-go.
	failed := make(chan error)
	for res, svc := range services {
		namespace := svc.GetNamespace()
		serviceName := svc.GetName()

		var servicePort int
		if req.Port == 0 {
			// We will use the first port in Service if not specified.
			servicePort = int(svc.Spec.Ports[0].Port)
		} else {
			servicePort = req.Port
		}

		cfg, err := clientcmd.BuildConfigFromFlags("", kubeops.GetKubeConfig(res))
		if err != nil {
			return err
		}

		clientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return err
		}

		go func() {
			err = ForwardPort(ctx, cfg, clientset, namespace, serviceName, servicePort, servicePort, stopChan, readyChan)
			failed <- err
		}()
	}
	err := <-failed
	return err
}

func ForwardPort(
	ctx context.Context,
	restConfig *rest.Config,
	clientset *kubernetes.Clientset,
	namespace, serviceName string,
	servicePort, localPort int,
	stopChan chan struct{}, readyChan chan struct{},
) error {
	svc, err := clientset.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if svc == nil {
		return fmt.Errorf("failed to find service: %s", serviceName)
	}

	labels := []string{}
	for k, v := range svc.Spec.Selector {
		labels = append(labels, strings.Join([]string{k, v}, "="))
	}
	label := strings.Join(labels, ",")

	// Select the first pod to forward the target port.
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: label, Limit: 1,
	})
	if err != nil {
		return err
	}
	if len(pods.Items) < 1 {
		return fmt.Errorf("pods of the service '%s' not found", serviceName)
	}
	pod := pods.Items[0]

	fmt.Printf("Forwarding localhost port to targetPort of pod '%s' selected by the service '%s' (%d:%d)\n",
		pod.Name, serviceName, localPort, servicePort)

	// Build a URL for SPDY connection for port-forwarding.
	url := clientset.CoreV1().RESTClient().Post().
		Resource("pods").Namespace(pod.Namespace).Name(pod.Name).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return err
	}

	ports := []string{fmt.Sprintf("%d:%d", localPort, servicePort)}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
	out, errOut := os.Stdout, os.Stderr

	fw, err := portforward.NewOnAddresses(dialer, []string{"localhost"}, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		return err
	}

	return fw.ForwardPorts()
}

func validatePortForwardRequest(req *PortForwardRequest) error {
	if req == nil {
		return errors.New("request is nil")
	}
	if err := release.ValidateSpec(req.Spec); err != nil {
		return err
	}
	if req.Port < 0 || req.Port > 65535 {
		return fmt.Errorf("invalid port %d", req.Port)
	}
	return nil
}
