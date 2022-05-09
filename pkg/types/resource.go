package types

type K8SResource string

const (
	K8sPod                     = K8SResource("Pod")
	K8sDeployment              = K8SResource("Deployment")
	K8sService                 = K8SResource("Service")
	K8sDaemonSet               = K8SResource("DaemonSet")
	K8sReplicaSet              = K8SResource("ReplicaSet")
	K8sStatefulSet             = K8SResource("StatefulSet")
	K8sHorizontalPodAutoscaler = K8SResource("HorizontalPodAutoscaler")
	K8sJob                     = K8SResource("HorizontalPodAutoscaler")
	K8sNode                    = K8SResource("Node")
	K8sNamespace               = K8SResource("Namespace")
	K8sNetworkPolicy           = K8SResource("NetworkPolicy")
	K8sRoleBinding             = K8SResource("RoleBinding")
	K8sClusterRole             = K8SResource("ClusterRole")
	K8sClusterRoleBinding      = K8SResource("ClusterRoleBinding")
	K8sServiceAccount          = K8SResource("ServiceAccount")
)
