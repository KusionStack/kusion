package common

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/third_party/kubevela/kubevela/apis/condition"
	workflowv1alpha1 "kusionstack.io/kusion/third_party/kubevela/workflow/api/v1alpha1"
)

// AppStatus defines the observed state of Application
type AppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	condition.ConditionedStatus `json:",inline"`

	// The generation observed by the application controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	Phase ApplicationPhase `json:"status,omitempty"`

	// Components record the related Components created by Application Controller
	Components []corev1.ObjectReference `json:"components,omitempty"`

	// Services record the status of the application services
	Services []ApplicationComponentStatus `json:"services,omitempty"`

	// Workflow record the status of workflow
	Workflow *WorkflowStatus `json:"workflow,omitempty"`

	// LatestRevision of the application configuration it generates
	// +optional
	LatestRevision *Revision `json:"latestRevision,omitempty"`

	// AppliedResources record the resources that the  workflow step apply.
	AppliedResources []ClusterObjectReference `json:"appliedResources,omitempty"`

	// PolicyStatus records the status of policy
	// Deprecated This field is only used by EnvBinding Policy which is deprecated.
	PolicyStatus []PolicyStatus `json:"policy,omitempty"`
}

// ApplicationPhase is a label for the condition of an application at the current time
type ApplicationPhase string

// ApplicationComponent describe the component of application
type ApplicationComponent struct {
	Name string `json:"name"`
	Type string `json:"type"`
	// ExternalRevision specified the component revisionName
	ExternalRevision string `json:"externalRevision,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Properties *runtime.RawExtension `json:"properties,omitempty"`

	DependsOn []string                     `json:"dependsOn,omitempty"`
	Inputs    workflowv1alpha1.StepInputs  `json:"inputs,omitempty"`
	Outputs   workflowv1alpha1.StepOutputs `json:"outputs,omitempty"`

	// Traits define the trait of one component, the type must be array to keep the order.
	Traits []ApplicationTrait `json:"traits,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// scopes in ApplicationComponent defines the component-level scopes
	// the format is <scope-type:scope-instance-name> pairs, the key represents type of `ScopeDefinition` while the value represent the name of scope instance.
	Scopes map[string]string `json:"scopes,omitempty"`

	// ReplicaKey is not empty means the component is replicated. This field is designed so that it can't be specified in application directly.
	// So we set the json tag as "-". Instead, this will be filled when using replication policy.
	ReplicaKey string `json:"-"`
}

// ApplicationComponentStatus record the health status of App component
type ApplicationComponentStatus struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Cluster   string `json:"cluster,omitempty"`
	Env       string `json:"env,omitempty"`
	// WorkloadDefinition is the definition of a WorkloadDefinition, such as deployments/apps.v1
	WorkloadDefinition WorkloadGVK              `json:"workloadDefinition,omitempty"`
	Healthy            bool                     `json:"healthy"`
	Message            string                   `json:"message,omitempty"`
	Traits             []ApplicationTraitStatus `json:"traits,omitempty"`
	Scopes             []corev1.ObjectReference `json:"scopes,omitempty"`
}

// WorkloadGVK refer to a Workload Type
type WorkloadGVK struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

// ApplicationTrait defines the trait of application
type ApplicationTrait struct {
	Type string `json:"type"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Properties *runtime.RawExtension `json:"properties,omitempty"`
}

// ApplicationTraitStatus records the trait health status
type ApplicationTraitStatus struct {
	Type    string `json:"type"`
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}

// WorkflowStatus record the status of workflow
type WorkflowStatus struct {
	AppRevision string                            `json:"appRevision,omitempty"`
	Mode        string                            `json:"mode"`
	Phase       workflowv1alpha1.WorkflowRunPhase `json:"status,omitempty"`
	Message     string                            `json:"message,omitempty"`

	Suspend      bool   `json:"suspend"`
	SuspendState string `json:"suspendState,omitempty"`

	Terminated bool `json:"terminated"`
	Finished   bool `json:"finished"`

	ContextBackend *corev1.ObjectReference               `json:"contextBackend,omitempty"`
	Steps          []workflowv1alpha1.WorkflowStepStatus `json:"steps,omitempty"`

	StartTime metav1.Time `json:"startTime,omitempty"`
	// +nullable
	EndTime metav1.Time `json:"endTime,omitempty"`
}

// Revision has name and revision number
type Revision struct {
	Name     string `json:"name"`
	Revision int64  `json:"revision"`

	// RevisionHash record the hash value of the spec of ApplicationRevision object.
	RevisionHash string `json:"revisionHash,omitempty"`
}

// ClusterObjectReference defines the object reference with cluster.
type ClusterObjectReference struct {
	Cluster                string              `json:"cluster,omitempty"`
	Creator                ResourceCreatorRole `json:"creator,omitempty"`
	corev1.ObjectReference `json:",inline"`
}

// ResourceCreatorRole defines the resource creator.
type ResourceCreatorRole string

// PolicyStatus records the status of policy
// Deprecated
type PolicyStatus struct {
	Name string `json:"name"`
	Type string `json:"type"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Status *runtime.RawExtension `json:"status,omitempty"`
}
