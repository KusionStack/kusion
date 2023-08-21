package network

import (
	"fmt"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ac "kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/network"
	"kusionstack.io/kusion/pkg/projectstack"
)

const kindK8sIngress = "Ingress"

// containerPortsGenerator is used to generate k8s service.
type routeGenerator struct {
	project        *projectstack.Project
	stack          *projectstack.Stack
	appName        string
	containerPorts []network.ContainerPort
	routes         map[string]network.Route
}

// NewRouteGenerator returns a new routeGenerator instance.
func NewRouteGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	cps []network.ContainerPort,
	routes map[string]network.Route,
) (ac.Generator, error) {
	if project.Name == "" {
		return nil, fmt.Errorf("project name must not be empty")
	}
	if stack.Name == "" {
		return nil, fmt.Errorf("stack name must not be empty")
	}
	if appName == "" {
		return nil, fmt.Errorf("app name must not be empty")
	}
	if len(cps) == 0 {
		return nil, fmt.Errorf("container posts must not be empty")
	}
	if routes == nil {
		return nil, fmt.Errorf("route must not be empty")
	}
	if err := validAccessPort(cps, routes); err != nil {
		return nil, err
	}

	return &routeGenerator{
		project:        project,
		stack:          stack,
		appName:        appName,
		containerPorts: cps,
		routes:         routes,
	}, nil
}

// NewRouteGeneratorFunc returns a new NewGeneratorFunc that returns a
// routeGenerator instance.
func NewRouteGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	cps []network.ContainerPort,
	routes map[string]network.Route,
) ac.NewGeneratorFunc {
	return func() (ac.Generator, error) {
		return NewRouteGenerator(project, stack, appName, cps, routes)
	}
}

func (g *routeGenerator) Generate(spec *models.Spec) error {
	svcName := ac.UniqueAppName(g.project.Name, g.stack.Name, g.appName)
	var ingressSpec v1.IngressSpec
	secretHosts := make(map[string][]string)

	for host, route := range g.routes {
		// add rule for per path
		var ingressRuleValue v1.HTTPIngressRuleValue

		for _, path := range route.Paths {
			pathType := v1.PathType(path.PathType)
			if pathType == "" {
				pathType = v1.PathTypeImplementationSpecific
			}
			ingressPath := v1.HTTPIngressPath{
				Path:     path.Path,
				PathType: &pathType,
				Backend: v1.IngressBackend{
					Service: &v1.IngressServiceBackend{
						Name: svcName,
						Port: v1.ServiceBackendPort{
							Number: int32(path.AccessPort),
						},
					},
				},
			}

			ingressRuleValue.Paths = append(ingressRuleValue.Paths, ingressPath)
		}
		rule := v1.IngressRule{
			Host: host,
			IngressRuleValue: v1.IngressRuleValue{
				HTTP: &ingressRuleValue,
			},
		}
		ingressSpec.Rules = append(ingressSpec.Rules, rule)

		// generate tlsSecret-hostNames map
		if route.TLSSecret != "" && host != "" {
			secretHosts[route.TLSSecret] = append(secretHosts[route.TLSSecret], host)
		}
	}

	// add tls per tlsSecret
	for secret, hosts := range secretHosts {
		tls := v1.IngressTLS{
			Hosts:      hosts,
			SecretName: secret,
		}
		ingressSpec.TLS = append(ingressSpec.TLS, tls)
	}

	// generate kusion resource
	resource := &v1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       kindK8sIngress,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ac.UniqueAppName(g.project.Name, g.stack.Name, g.appName),
			Namespace: g.project.Name,
		},
		Spec: ingressSpec,
	}
	resourceID := ac.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta)

	return ac.AppendToSpec(resourceID, resource, spec)
}

func validAccessPort(cps []network.ContainerPort, routes map[string]network.Route) error {
	accessPorts := make(map[int]struct{})
	for _, cp := range cps {
		accessPorts[cp.AccessPort] = struct{}{}
	}

	accessPortPaths := make(map[int]string)
	accessPortHosts := make(map[int]string)
	for host, route := range routes {
		for _, path := range route.Paths {
			if _, ok := accessPorts[path.AccessPort]; !ok {
				return fmt.Errorf("accessPort %d of host %s and path %s is not exposed",
					path.AccessPort, host, path.Path)
			}
			if _, ok := accessPortPaths[path.AccessPort]; !ok {
				return fmt.Errorf("host %s, path %s and host %s, path %s use the same accessPort %d",
					accessPortHosts[path.AccessPort], accessPortPaths[path.AccessPort], host, path.Path, path.AccessPort)
			}
			accessPortPaths[path.AccessPort] = path.Path
			accessPortHosts[path.AccessPort] = host
		}
	}

	return nil
}
