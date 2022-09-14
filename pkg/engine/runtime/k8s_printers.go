package runtime

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	loadBalancerWidth = 16
)

func printNamespace(obj *corev1.Namespace) (string, bool) {
	return fmt.Sprintf("Phase: %s", obj.Status.Phase), true
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

	return fmt.Sprintf("Type: %s, InternalIP: %s, ExternalIP: %s, Port(s): %s", svcType, internalIP, externalIP, svcPorts), true
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

func printDeployment(obj *appsv1.Deployment) (string, bool) {
	desiredReplicas := *(obj.Spec.Replicas)
	updatedReplicas := obj.Status.UpdatedReplicas
	readyReplicas := obj.Status.ReadyReplicas
	availableReplicas := obj.Status.AvailableReplicas
	stop := desiredReplicas == availableReplicas
	return fmt.Sprintf("readyReplicas: %d, desiredReplicas: %d, updatedReplicas: %d, availableReplicas: %d",
		readyReplicas, desiredReplicas, updatedReplicas, availableReplicas), stop
}
