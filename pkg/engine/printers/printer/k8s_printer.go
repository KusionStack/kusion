package printer

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	loadBalancerWidth = 16

	// labelNodeRolePrefix is a label prefix for node roles
	// It's copied over to here until it's merged in core: https://github.com/kubernetes/kubernetes/pull/39112
	labelNodeRolePrefix = "node-role.kubernetes.io/"

	// nodeLabelRole specifies the role of a node
	nodeLabelRole = "kubernetes.io/role"
)

// AddHandlers adds print handlers for default Kubernetes types dealing with internal versions.
// TODO: handle errors from Handler
func AddK8sHandlers(h PrintHandler) {
	// core/v1
	h.TableHandler(printComponentStatus)
	h.TableHandler(printConfigMap)
	h.TableHandler(printEndpoints)
	h.TableHandler(printEvent)
	h.TableHandler(printNamespace)
	h.TableHandler(printNode)
	h.TableHandler(printPersistentVolumeClaim)
	h.TableHandler(printPersistentVolume)
	h.TableHandler(printPod)
	h.TableHandler(printPodTemplate)
	h.TableHandler(printReplicationController)
	h.TableHandler(printResourceQuota)
	h.TableHandler(printSecret)
	h.TableHandler(printServiceAccount)
	h.TableHandler(printService)
	// apps/v1
	h.TableHandler(printDeployment)
	h.TableHandler(printReplicaSet)
	h.TableHandler(printDaemonSet)
	h.TableHandler(printStatefulSet)
	h.TableHandler(printControllerRevision)
	// discovery.k8s.io/v1
	h.TableHandler(printEndpointSlice)
	// batch/v1
	h.TableHandler(printCronJob)
	h.TableHandler(printJob)
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

	ready := !strings.Contains(externalIP, "pending")
	return fmt.Sprintf("Type: %s, InternalIP: %s, ExternalIP: %s, Port(s): %s",
		svcType, internalIP, externalIP, svcPorts), ready
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

	ready := false
	for i := range obj.Subsets {
		ss := &obj.Subsets[i]
		if len(ss.Addresses) > 0 {
			ready = true
			break
		}
	}
	return fmt.Sprintf("Endpoints: %s", eps), ready
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

func printComponentStatus(obj *corev1.ComponentStatus) (string, bool) {
	status := "Unknown"
	message := ""
	err := ""
	for _, condition := range obj.Conditions {
		if condition.Type == corev1.ComponentHealthy {
			if condition.Status == corev1.ConditionTrue {
				status = "Healthy"
			} else {
				status = "Unhealthy"
			}
			message = condition.Message
			err = condition.Error
			break
		}
	}

	ready := status == "Healthy"
	return fmt.Sprintf("Status: %s, Message: %s, error: %s", status, message, err), ready
}

func printConfigMap(obj *corev1.ConfigMap) (string, bool) {
	data := int64(len(obj.Data) + len(obj.BinaryData))
	age := translateTimestampSince(obj.CreationTimestamp)
	return fmt.Sprintf("Data: %d, Age: %s", data, age), true
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

func printEvent(obj *corev1.Event) (string, bool) {
	firstTimestamp := translateTimestampSince(obj.FirstTimestamp)
	if obj.FirstTimestamp.IsZero() {
		firstTimestamp = translateMicroTimestampSince(obj.EventTime)
	}

	lastTimestamp := translateTimestampSince(obj.LastTimestamp)
	if obj.LastTimestamp.IsZero() {
		lastTimestamp = firstTimestamp
	}

	var target string
	if len(obj.InvolvedObject.Name) > 0 {
		target = fmt.Sprintf("%s/%s", strings.ToLower(obj.InvolvedObject.Kind), obj.InvolvedObject.Name)
	} else {
		target = strings.ToLower(obj.InvolvedObject.Kind)
	}

	return fmt.Sprintf("Last Seen: %s, Type: %s, Reason: %s, Target: %s, Message: %s",
		lastTimestamp, obj.Type, obj.Reason, target, obj.Message), true
}

// translateMicroTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateMicroTimestampSince(timestamp metav1.MicroTime) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

func printNode(obj *corev1.Node) (string, bool) {
	conditionMap := make(map[corev1.NodeConditionType]*corev1.NodeCondition)
	NodeAllConditions := []corev1.NodeConditionType{corev1.NodeReady}
	for i := range obj.Status.Conditions {
		cond := obj.Status.Conditions[i]
		conditionMap[cond.Type] = &cond
	}
	var status []string
	for _, validCondition := range NodeAllConditions {
		if condition, ok := conditionMap[validCondition]; ok {
			if condition.Status == corev1.ConditionTrue {
				status = append(status, string(condition.Type))
			} else {
				status = append(status, "Not"+string(condition.Type))
			}
		}
	}
	if len(status) == 0 {
		status = append(status, "Unknown")
	}
	if obj.Spec.Unschedulable {
		status = append(status, "SchedulingDisabled")
	}
	statusStr := strings.Join(status, ",")

	roles := strings.Join(findNodeRoles(obj), ",")
	if len(roles) == 0 {
		roles = "<none>"
	}

	ready := statusStr == string(corev1.NodeReady)
	return fmt.Sprintf("Status: %s, Roles: %s, Age: %s, Version: %s",
		statusStr, roles, translateTimestampSince(obj.CreationTimestamp), obj.Status.NodeInfo.KubeletVersion), ready
}

// findNodeRoles returns the roles of a given node.
// The roles are determined by looking for:
// * a node-role.kubernetes.io/<role>="" label
// * a kubernetes.io/role="<role>" label
func findNodeRoles(node *corev1.Node) []string {
	roles := sets.NewString()
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, labelNodeRolePrefix):
			if role := strings.TrimPrefix(k, labelNodeRolePrefix); len(role) > 0 {
				roles.Insert(role)
			}

		case k == nodeLabelRole && v != "":
			roles.Insert(v)
		}
	}
	return roles.List()
}

func printPersistentVolumeClaim(obj *corev1.PersistentVolumeClaim) (string, bool) {
	phase := obj.Status.Phase
	if obj.ObjectMeta.DeletionTimestamp != nil {
		phase = "Terminating"
	}

	storage := obj.Spec.Resources.Requests[corev1.ResourceStorage]
	capacity := ""
	accessModes := ""
	volumeMode := "<unset>"
	if obj.Spec.VolumeName != "" {
		accessModes = getAccessModesAsString(obj.Status.AccessModes)
		storage = obj.Status.Capacity[corev1.ResourceStorage]
		capacity = storage.String()
	}

	if obj.Spec.VolumeMode != nil {
		volumeMode = string(*obj.Spec.VolumeMode)
	}

	storageClass := getPersistentVolumeClaimClass(obj)
	age := translateTimestampSince(obj.CreationTimestamp)
	return fmt.Sprintf("Status: %s, Volume: %s, Capacity: %s, Access Modes: %s, StorageClass: %s, Age: %s, VolumeMode: %s",
		string(phase), obj.Spec.VolumeName, capacity, accessModes, storageClass, age, volumeMode), true
}

// getAccessModesAsString returns a string representation of an array of access modes.
// modes, when present, are always in the same order: RWO,ROX,RWX,RWOP.
func getAccessModesAsString(modes []corev1.PersistentVolumeAccessMode) string {
	modes = removeDuplicateAccessModes(modes)
	modesStr := []string{}
	if containsAccessMode(modes, corev1.ReadWriteOnce) {
		modesStr = append(modesStr, "RWO")
	}
	if containsAccessMode(modes, corev1.ReadOnlyMany) {
		modesStr = append(modesStr, "ROX")
	}
	if containsAccessMode(modes, corev1.ReadWriteMany) {
		modesStr = append(modesStr, "RWX")
	}
	return strings.Join(modesStr, ",")
}

// removeDuplicateAccessModes returns an array of access modes without any duplicates
func removeDuplicateAccessModes(modes []corev1.PersistentVolumeAccessMode) []corev1.PersistentVolumeAccessMode {
	accessModes := []corev1.PersistentVolumeAccessMode{}
	for _, m := range modes {
		if !containsAccessMode(accessModes, m) {
			accessModes = append(accessModes, m)
		}
	}
	return accessModes
}

func containsAccessMode(modes []corev1.PersistentVolumeAccessMode, mode corev1.PersistentVolumeAccessMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

// getPersistentVolumeClaimClass returns StorageClassName. If no storage class was
// requested, it returns "".
func getPersistentVolumeClaimClass(claim *corev1.PersistentVolumeClaim) string {
	// Use beta annotation first
	if class, found := claim.Annotations[corev1.BetaStorageClassAnnotation]; found {
		return class
	}

	if claim.Spec.StorageClassName != nil {
		return *claim.Spec.StorageClassName
	}

	return ""
}

func printPersistentVolume(obj *corev1.PersistentVolume) (string, bool) {
	claimRefUID := ""
	if obj.Spec.ClaimRef != nil {
		claimRefUID += obj.Spec.ClaimRef.Namespace
		claimRefUID += "/"
		claimRefUID += obj.Spec.ClaimRef.Name
	}

	modesStr := getAccessModesAsString(obj.Spec.AccessModes)
	reclaimPolicyStr := string(obj.Spec.PersistentVolumeReclaimPolicy)

	aQty := obj.Spec.Capacity[corev1.ResourceStorage]
	aSize := aQty.String()

	phase := obj.Status.Phase
	if obj.ObjectMeta.DeletionTimestamp != nil {
		phase = "Terminating"
	}
	volumeMode := "<unset>"
	if obj.Spec.VolumeMode != nil {
		volumeMode = string(*obj.Spec.VolumeMode)
	}

	storageClass := getPersistentVolumeClass(obj)
	age := translateTimestampSince(obj.CreationTimestamp)

	return fmt.Sprintf("Capacity: %s, Access Modes: %s, Reclaim Policy: %s, Status: %s, Claim: %s, StorageClass: %s, Reason: %s, Age: %s, VolumeMode: %s",
		aSize, modesStr, reclaimPolicyStr, string(phase), claimRefUID, storageClass, obj.Status.Reason, age, volumeMode), true
}

// getPersistentVolumeClass returns StorageClassName.
func getPersistentVolumeClass(volume *corev1.PersistentVolume) string {
	// Use beta annotation first
	if class, found := volume.Annotations[corev1.BetaStorageClassAnnotation]; found {
		return class
	}

	return volume.Spec.StorageClassName
}

var (
	podSuccessConditions = []metav1.TableRowCondition{
		{
			Type:    metav1.RowCompleted,
			Status:  metav1.ConditionTrue,
			Reason:  string(corev1.PodSucceeded),
			Message: "The pod has completed successfully.",
		},
	}
	podFailedConditions = []metav1.TableRowCondition{
		{
			Type:    metav1.RowCompleted,
			Status:  metav1.ConditionTrue,
			Reason:  string(corev1.PodFailed),
			Message: "The pod failed.",
		},
	}
)

func printPod(pod *corev1.Pod) (string, bool) {
	restarts := 0
	totalContainers := len(pod.Spec.Containers)
	readyContainers := 0
	lastRestartDate := metav1.NewTime(time.Time{})

	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: pod},
	}

	switch pod.Status.Phase {
	case corev1.PodSucceeded:
		row.Conditions = podSuccessConditions
	case corev1.PodFailed:
		row.Conditions = podFailedConditions
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		if container.LastTerminationState.Terminated != nil {
			terminatedDate := container.LastTerminationState.Terminated.FinishedAt
			if lastRestartDate.Before(&terminatedDate) {
				lastRestartDate = terminatedDate
			}
		}
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.LastTerminationState.Terminated != nil {
				terminatedDate := container.LastTerminationState.Terminated.FinishedAt
				if lastRestartDate.Before(&terminatedDate) {
					lastRestartDate = terminatedDate
				}
			}
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if reason == "Completed" && hasRunning {
			if hasPodReadyCondition(pod.Status.Conditions) {
				reason = "Running"
			} else {
				reason = "NotReady"
			}
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		reason = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}

	restartsStr := strconv.Itoa(restarts)
	if !lastRestartDate.IsZero() {
		restartsStr = fmt.Sprintf("%d (%s ago)", restarts, translateTimestampSince(lastRestartDate))
	}

	readyStr := fmt.Sprintf("%d/%d", readyContainers, totalContainers)
	ready := readyContainers == totalContainers
	age := translateTimestampSince(pod.CreationTimestamp)

	return fmt.Sprintf("Ready: %s, Status: %s, Restart: %s, Age: %s", readyStr, reason, restartsStr, age), ready
}

func hasPodReadyCondition(conditions []corev1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func printPodTemplate(obj *corev1.PodTemplate) (string, bool) {
	names, images := layoutContainerCells(obj.Template.Spec.Containers)
	return fmt.Sprintf("Containers: %s, Images: %s, Pod Labels: %s",
		names, images, labels.FormatLabels(obj.Template.Labels)), true
}

// Lay out all the containers on one line if you use wide output.
func layoutContainerCells(containers []corev1.Container) (names string, images string) {
	var namesBuffer bytes.Buffer
	var imagesBuffer bytes.Buffer

	for i, container := range containers {
		namesBuffer.WriteString(container.Name)
		imagesBuffer.WriteString(container.Image)
		if i != len(containers)-1 {
			namesBuffer.WriteString(",")
			imagesBuffer.WriteString(",")
		}
	}
	return namesBuffer.String(), imagesBuffer.String()
}

func printReplicationController(obj *corev1.ReplicationController) (string, bool) {
	desiredReplicas := int64(*obj.Spec.Replicas)
	currentReplicas := int64(obj.Status.Replicas)
	readyReplicas := int64(obj.Status.ReadyReplicas)
	age := translateTimestampSince(obj.CreationTimestamp)
	ready := desiredReplicas == readyReplicas
	return fmt.Sprintf("Desired: %d, Current: %d, Ready: %d, Age: %s",
		desiredReplicas, currentReplicas, readyReplicas, age), ready
}

func printResourceQuota(resourceQuota *corev1.ResourceQuota) (string, bool) {
	resources := make([]string, 0, len(resourceQuota.Status.Hard))
	for resource := range resourceQuota.Status.Hard {
		resources = append(resources, string(resource))
	}
	sort.Strings(resources)

	requestColumn := bytes.NewBuffer([]byte{})
	limitColumn := bytes.NewBuffer([]byte{})
	for i := range resources {
		w := requestColumn
		resource := corev1.ResourceName(resources[i])
		usedQuantity := resourceQuota.Status.Used[resource]
		hardQuantity := resourceQuota.Status.Hard[resource]

		// use limitColumn writer if a resource name prefixed with "limits" is found
		if pieces := strings.Split(resource.String(), "."); len(pieces) > 1 && pieces[0] == "limits" {
			w = limitColumn
		}

		fmt.Fprintf(w, "%s: %s/%s, ", resource, usedQuantity.String(), hardQuantity.String())
	}

	age := translateTimestampSince(resourceQuota.CreationTimestamp)
	request := strings.TrimSuffix(requestColumn.String(), ", ")
	limit := strings.TrimSuffix(limitColumn.String(), ", ")
	return fmt.Sprintf("Age: %s, Request: %s, Limit: %s", age, request, limit), true
}

func printSecret(obj *corev1.Secret) (string, bool) {
	age := translateTimestampSince(obj.CreationTimestamp)
	return fmt.Sprintf("Type: %s, Data: %d, Age: %s", string(obj.Type), int64(len(obj.Data)), age), true
}

func printServiceAccount(obj *corev1.ServiceAccount) (string, bool) {
	age := translateTimestampSince(obj.CreationTimestamp)
	return fmt.Sprintf("Secrets: %d, Age: %s", int64(len(obj.Secrets)), age), true
}

func printDeployment(obj *appsv1.Deployment) (string, bool) {
	desiredReplicas := *(obj.Spec.Replicas)
	updatedReplicas := obj.Status.UpdatedReplicas
	readyReplicas := obj.Status.ReadyReplicas
	availableReplicas := obj.Status.AvailableReplicas
	return fmt.Sprintf("Ready: %d/%d, Up-to-date: %d, Available: %d",
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

func printReplicaSet(obj *appsv1.ReplicaSet) (string, bool) {
	desiredReplicas := *(obj.Spec.Replicas)
	currentReplicas := obj.Status.Replicas
	readyReplicas := obj.Status.ReadyReplicas
	return fmt.Sprintf("Desired: %d, Current: %d, Ready: %d",
		desiredReplicas, currentReplicas, readyReplicas), desiredReplicas == readyReplicas
}

func printDaemonSet(obj *appsv1.DaemonSet) (string, bool) {
	desiredScheduled := obj.Status.DesiredNumberScheduled
	currentScheduled := obj.Status.CurrentNumberScheduled
	numberReady := obj.Status.NumberReady
	numberUpdated := obj.Status.UpdatedNumberScheduled
	numberAvailable := obj.Status.NumberAvailable

	return fmt.Sprintf("Desired: %d, Current: %d, Ready: %d, Up-to-date: %d, Available: %d",
			desiredScheduled, currentScheduled, numberReady, numberUpdated, numberAvailable),
		desiredScheduled == numberReady
}

func printStatefulSet(obj *appsv1.StatefulSet) (string, bool) {
	desiredReplicas := *(obj.Spec.Replicas)
	readyReplicas := obj.Status.ReadyReplicas
	createTime := translateTimestampSince(obj.CreationTimestamp)
	return fmt.Sprintf("Ready: %d/%d, Age: %s", desiredReplicas, readyReplicas, createTime),
		desiredReplicas == readyReplicas
}

func printControllerRevision(obj *appsv1.ControllerRevision) (string, bool) {
	controllerRef := metav1.GetControllerOf(obj)
	noneController := "<none>"
	controllerName := noneController
	if controllerRef != nil {
		gv, _ := schema.ParseGroupVersion(controllerRef.APIVersion)
		gvk := gv.WithKind(controllerRef.Kind)
		controllerName = formatResourceName(gvk.GroupKind(), controllerRef.Name)
	}
	revision := obj.Revision
	return fmt.Sprintf("Controller: %s, Revision: %d", controllerName, revision), controllerName != noneController
}

func formatResourceName(kind schema.GroupKind, name string) string {
	if kind.Empty() {
		return name
	}

	return strings.ToLower(kind.String()) + "/" + name
}

func printCronJob(obj *batchv1.CronJob) (string, bool) {
	lastScheduleTime := "<none>"
	if obj.Status.LastScheduleTime != nil {
		lastScheduleTime = translateTimestampSince(*obj.Status.LastScheduleTime)
	}

	return fmt.Sprintf("Schedule: %s, Suspend: %s, Active: %d, Last Schedule: %s",
			obj.Spec.Schedule, printBoolPtr(obj.Spec.Suspend), len(obj.Status.Active), lastScheduleTime),
		lastScheduleTime != "<none>"
}

func printBoolPtr(value *bool) string {
	if value != nil {
		return printBool(*value)
	}

	return "<unset>"
}

func printBool(value bool) string {
	if value {
		return "True"
	}

	return "False"
}

func printJob(obj *batchv1.Job) (string, bool) {
	var completions string
	if obj.Spec.Completions != nil {
		completions = fmt.Sprintf("%d/%d", obj.Status.Succeeded, *obj.Spec.Completions)
	} else {
		parallelism := int32(0)
		if obj.Spec.Parallelism != nil {
			parallelism = *obj.Spec.Parallelism
		}
		if parallelism > 1 {
			completions = fmt.Sprintf("%d/1 of %d", obj.Status.Succeeded, parallelism)
		} else {
			completions = fmt.Sprintf("%d/1", obj.Status.Succeeded)
		}
	}
	var jobDuration string
	switch {
	case obj.Status.StartTime == nil:
	case obj.Status.CompletionTime == nil:
		jobDuration = duration.HumanDuration(time.Since(obj.Status.StartTime.Time))
	default:
		jobDuration = duration.HumanDuration(obj.Status.CompletionTime.Sub(obj.Status.StartTime.Time))
	}

	return fmt.Sprintf("Completions: %s, Duration: %s, Age: %s",
			completions, jobDuration, translateTimestampSince(obj.CreationTimestamp)),
		obj.Status.Succeeded == *obj.Spec.Completions
}
