package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// WorkflowMode describes the mode of workflow
type WorkflowMode string

// WorkflowRunPhase is a label for the condition of a WorkflowRun at the current time
type WorkflowRunPhase string

// WorkflowExecuteMode defines the mode of workflow execution
type WorkflowExecuteMode struct {
	// Steps is the mode of workflow steps execution
	Steps WorkflowMode `json:"steps,omitempty"`
	// SubSteps is the mode of workflow sub steps execution
	SubSteps WorkflowMode `json:"subSteps,omitempty"`
}

// WorkflowStep defines how to execute a workflow step.
type WorkflowStep struct {
	WorkflowStepBase `json:",inline"`
	SubSteps         []WorkflowStepBase `json:"subSteps,omitempty"`
}

// WorkflowStepBase defines the workflow step base
type WorkflowStepBase struct {
	// Name is the unique name of the workflow step.
	Name string `json:"name,omitempty"`
	// Type is the type of the workflow step.
	Type string `json:"type"`
	// Meta is the meta data of the workflow step.
	Meta *WorkflowStepMeta `json:"meta,omitempty"`
	// If is the if condition of the step
	If string `json:"if,omitempty"`
	// Timeout is the timeout of the step
	Timeout string `json:"timeout,omitempty"`
	// DependsOn is the dependency of the step
	DependsOn []string `json:"dependsOn,omitempty"`
	// Inputs is the inputs of the step
	Inputs StepInputs `json:"inputs,omitempty"`
	// Outputs is the outputs of the step
	Outputs StepOutputs `json:"outputs,omitempty"`

	// Properties is the properties of the step
	// +kubebuilder:pruning:PreserveUnknownFields
	Properties *runtime.RawExtension `json:"properties,omitempty"`
}

// WorkflowStepMeta contains the metadata of a workflow step
type WorkflowStepMeta struct {
	Alias string `json:"alias,omitempty"`
}

// StepOutputs defines output variable of WorkflowStep
type StepOutputs []outputItem

// StepInputs defines variable input of WorkflowStep
type StepInputs []inputItem

type inputItem struct {
	ParameterKey string `json:"parameterKey"`
	From         string `json:"from"`
}

type outputItem struct {
	ValueFrom string `json:"valueFrom"`
	Name      string `json:"name"`
}

// WorkflowStepPhase describes the phase of a workflow step.
type WorkflowStepPhase string

// StepStatus record the base status of workflow step, which could be workflow step or subStep
type StepStatus struct {
	ID    string            `json:"id"`
	Name  string            `json:"name,omitempty"`
	Type  string            `json:"type,omitempty"`
	Phase WorkflowStepPhase `json:"phase,omitempty"`
	// A human readable message indicating details about why the workflowStep is in this state.
	Message string `json:"message,omitempty"`
	// A brief CamelCase message indicating details about why the workflowStep is in this state.
	Reason string `json:"reason,omitempty"`
	// FirstExecuteTime is the first time this step execution.
	FirstExecuteTime metav1.Time `json:"firstExecuteTime,omitempty"`
	// LastExecuteTime is the last time this step execution.
	LastExecuteTime metav1.Time `json:"lastExecuteTime,omitempty"`
}

// WorkflowStepStatus record the status of a workflow step, include step status and subStep status
type WorkflowStepStatus struct {
	StepStatus     `json:",inline"`
	SubStepsStatus []StepStatus `json:"subSteps,omitempty"`
}
