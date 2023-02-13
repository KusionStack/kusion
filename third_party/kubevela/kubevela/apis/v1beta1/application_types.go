package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/third_party/kubevela/kubevela/apis/common"
	workflowv1alpha1 "kusionstack.io/kusion/third_party/kubevela/workflow/api/v1alpha1"
)

// Application is the Schema for the applications API
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec  `json:"spec,omitempty"`
	Status common.AppStatus `json:"status,omitempty"`
}

// ApplicationSpec is the spec of Application
type ApplicationSpec struct {
	Components []common.ApplicationComponent `json:"components"`

	// Policies defines the global policies for all components in the app, e.g. security, metrics, gitops,
	// multi-cluster placement rules, etc.
	// Policies are applied after components are rendered and before workflow steps are executed.
	Policies []AppPolicy `json:"policies,omitempty"`

	// Workflow defines how to customize the control logic.
	// If workflow is specified, Vela won't apply any resource, but provide rendered output in AppRevision.
	// Workflow steps are executed in array order, and each step:
	// - will have a context in annotation.
	// - should mark "finish" phase in status.conditions.
	Workflow *Workflow `json:"workflow,omitempty"`

	// TODO(wonderflow): we should have application level scopes supported here
}

// AppPolicy defines a global policy for all components in the app.
type AppPolicy struct {
	// Name is the unique name of the policy.
	Name string `json:"name"`

	Type string `json:"type"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Properties *runtime.RawExtension `json:"properties,omitempty"`
}

// Workflow defines workflow steps and other attributes
type Workflow struct {
	Ref   string                                `json:"ref,omitempty"`
	Mode  *workflowv1alpha1.WorkflowExecuteMode `json:"mode,omitempty"`
	Steps []workflowv1alpha1.WorkflowStep       `json:"steps,omitempty"`
}
