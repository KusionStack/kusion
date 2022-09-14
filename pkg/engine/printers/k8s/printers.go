package k8s

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"kusionstack.io/kusion/pkg/engine/printers"
)

const (
	loadBalancerWidth = 16
)

// AddHandlers adds print handlers for default Kubernetes types dealing with internal versions.
// TODO: handle errors from Handler
func AddHandlers(h printers.PrintHandler) {
	// core/v1
	h.TableHandler(printNamespace)
	h.TableHandler(printService)
	h.TableHandler(printEndpoints)
	// apps/v1
	h.TableHandler(printDeployment)
	// discovery.k8s.io/v1
	h.TableHandler(printEndpointSlice)
}

func printNamespace(obj *corev1.Namespace) (string, bool) {
	return fmt.Sprintf("Phase: %s", obj.Status.Phase),
		obj.Status.Phase == corev1.NamespaceActive
}

func printService(obj *corev1.Service) (string, bool) {
	svcType := obj.Spec.Type
	internalIP := "<none>"
	if len(obj.Spec.ClusterIPs) > 0 {
		internalIP = obj.Spec.ClusterIPs[0]
	}

	externalIP := getServiceExternalIP(obj, false)
	svcPorts := makePortString(obj.Spec.Ports)
	if len(svcPorts) == 0 {
		svcPorts = "<none>"
	}

	return fmt.Sprintf("Type: %s, InternalIP: %s, ExternalIP: %s, Port(s): %s", svcType, internalIP, externalIP, svcPorts),
		!strings.Contains(externalIP, "pending")
}

func getServiceExternalIP(svc *corev1.Service, wide bool) string {
	switch svc.Spec.Type {
	case corev1.ServiceTypeClusterIP:
		if len(svc.Spec.ExternalIPs) > 0 {
			return strings.Join(svc.Spec.ExternalIPs, ",")
		}
		return "<none>"
	case corev1.ServiceTypeNodePort:
		if len(svc.Spec.ExternalIPs) > 0 {
			return strings.Join(svc.Spec.ExternalIPs, ",")
		}
		return "<none>"
	case corev1.ServiceTypeLoadBalancer:
		lbIps := loadBalancerStatusStringer(svc.Status.LoadBalancer, wide)
		if len(svc.Spec.ExternalIPs) > 0 {
			results := []string{}
			if len(lbIps) > 0 {
				results = append(results, strings.Split(lbIps, ",")...)
			}
			results = append(results, svc.Spec.ExternalIPs...)
			return strings.Join(results, ",")
		}
		if len(lbIps) > 0 {
			return lbIps
		}
		return "<pending>"
	case corev1.ServiceTypeExternalName:
		return svc.Spec.ExternalName
	}
	return "<unknown>"
}

func makePortString(ports []corev1.ServicePort) string {
	pieces := make([]string, len(ports))
	for ix := range ports {
		port := &ports[ix]
		pieces[ix] = fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		if port.NodePort > 0 {
			pieces[ix] = fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol)
		}
	}
	return strings.Join(pieces, ",")
}

// loadBalancerStatusStringer behaves mostly like a string interface and converts the given status to a string.
// `wide` indicates whether the returned value:meant for --o=wide output. If not, it's clipped to 16 bytes.
func loadBalancerStatusStringer(s corev1.LoadBalancerStatus, wide bool) string {
	ingress := s.Ingress
	result := sets.NewString()
	for i := range ingress {
		if ingress[i].IP != "" {
			result.Insert(ingress[i].IP)
		} else if ingress[i].Hostname != "" {
			result.Insert(ingress[i].Hostname)
		}
	}

	r := strings.Join(result.List(), ",")
	if !wide && len(r) > loadBalancerWidth {
		r = r[0:(loadBalancerWidth-3)] + "..."
	}
	return r
}

func printEndpoints(obj *corev1.Endpoints) (string, bool) {
	eps := formatEndpoints(obj, nil)

	flag := false
	for i := range obj.Subsets {
		ss := &obj.Subsets[i]
		if len(ss.Addresses) > 0 {
			flag = true
			break
		}
	}
	return fmt.Sprintf("Endpoints: %s", eps), flag
}

// Set ports=nil for all ports.
func formatEndpoints(endpoints *corev1.Endpoints, ports sets.String) string {
	if len(endpoints.Subsets) == 0 {
		return "<none>"
	}
	list := []string{}
	max := 3
	more := false
	count := 0
	for i := range endpoints.Subsets {
		ss := &endpoints.Subsets[i]
		if len(ss.Ports) == 0 {
			// It's possible to have headless services with no ports.
			for i := range ss.Addresses {
				if len(list) == max {
					more = true
				}
				if !more {
					list = append(list, ss.Addresses[i].IP)
				}
				count++
			}
		} else {
			// "Normal" services with ports defined.
			for i := range ss.Ports {
				port := &ss.Ports[i]
				if ports == nil || ports.Has(port.Name) {
					for i := range ss.Addresses {
						if len(list) == max {
							more = true
						}
						addr := &ss.Addresses[i]
						if !more {
							hostPort := net.JoinHostPort(addr.IP, strconv.Itoa(int(port.Port)))
							list = append(list, hostPort)
						}
						count++
					}
				}
			}
		}
	}
	ret := strings.Join(list, ",")
	if more {
		return fmt.Sprintf("%s + %d more...", ret, count-max)
	}
	return ret
}

func printDeployment(obj *appsv1.Deployment) (string, bool) {
	desiredReplicas := *(obj.Spec.Replicas)
	updatedReplicas := obj.Status.UpdatedReplicas
	readyReplicas := obj.Status.ReadyReplicas
	availableReplicas := obj.Status.AvailableReplicas
	return fmt.Sprintf("readyReplicas: %d, desiredReplicas: %d, updatedReplicas: %d, availableReplicas: %d",
		readyReplicas, desiredReplicas, updatedReplicas, availableReplicas), desiredReplicas == availableReplicas
}

func printEndpointSlice(obj *discoveryv1.EndpointSlice) (string, bool) {
	addressType := string(obj.AddressType)
	ports := formatDiscoveryPorts(obj.Ports)
	eps := formatDiscoveryEndpoints(obj.Endpoints)

	return fmt.Sprintf("AddressType: %s, Ports: %s, Endpoints: %s", addressType, ports, eps),
		len(obj.Endpoints) > 0
}

func formatDiscoveryPorts(ports []discoveryv1.EndpointPort) string {
	list := []string{}
	max := 3
	more := false
	count := 0
	for _, port := range ports {
		if len(list) < max {
			portNum := "*"
			if port.Port != nil {
				portNum = strconv.Itoa(int(*port.Port))
			} else if port.Name != nil {
				portNum = *port.Name
			}
			list = append(list, portNum)
		} else if len(list) == max {
			more = true
		}
		count++
	}
	return listWithMoreString(list, more, count, max)
}

func formatDiscoveryEndpoints(endpoints []discoveryv1.Endpoint) string {
	list := []string{}
	max := 3
	more := false
	count := 0
	for _, endpoint := range endpoints {
		for _, address := range endpoint.Addresses {
			if len(list) < max {
				list = append(list, address)
			} else if len(list) == max {
				more = true
			}
			count++
		}
	}
	return listWithMoreString(list, more, count, max)
}

func listWithMoreString(list []string, more bool, count, max int) string {
	ret := strings.Join(list, ",")
	if more {
		return fmt.Sprintf("%s + %d more...", ret, count-max)
	}
	if ret == "" {
		ret = "<unset>"
	}
	return ret
}
