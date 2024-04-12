package service

import "github.com/hashicorp/go-multierror"

const DefaultMaxRoutines = 10

// CommonOptions is the common options for preview and sync operation.
type CommonOptions struct {
	PullOptions `json:",inline" yaml:",inline"`
	// Envs lets you set the env when executes in the form "key=value".
	Envs []string `json:"envs,omitempty" yaml:"envs,omitempty"`
	// DisableState is the flag to disable state management.
	DisableState bool `json:"disableState,omitempty"`
	// Extensions is the extensions for the stack request.
	Extensions Extensions `json:"extensions,omitempty"`
}

// Extensions is the extensions for the stack request.
type Extensions struct {
	Kusion    KusionExtensions    `json:"kusion,omitempty"`
	Terraform TerraformExtensions `json:"terraform,omitempty"`
}

// KusionExtensions is the extensions for the Kusion stack request.
type KusionExtensions struct {
	IgnoreFields []string `json:"ignoreFields,omitempty"`
	KCLArguments []string `json:"kclArguments,omitempty" yaml:"kclArguments,omitempty"`
	Color        bool     `json:"color,omitempty" yaml:"color,omitempty"`
	SpecFile     string   `json:"specFile,omitempty" yaml:"specFile,omitempty"`
	Version      string   `json:"version,omitempty" yaml:"version,omitempty"`
}

// TerraformExtensions is the extensions for the Terraform stack request.
type TerraformExtensions struct{}

// PullOptions is the options for pull.
type PullOptions struct {
	// SourceProviderType is the type of the source provider.
	SourceProviderType string `json:"sourceProviderType,omitempty" yaml:"sourceProviderType,omitempty"`
	// The version of the stack to be pulled. If not specified,
	// use the desired of stack.
	Version         string   `json:"version,omitempty" yaml:"version,omitempty"`
	AdditionalPaths []string `json:"additionalPaths,omitempty" yaml:"additionalPaths,omitempty"`
}

// PreviewOptions is the options for preview.
type PreviewOptions struct {
	CommonOptions `json:",inline" yaml:",inline"`
	// TODO: need to implement
	DriftMode bool `json:"driftMode,omitempty" yaml:"driftMode,omitempty"`
	// OutputFormat specify the output format, one of '', 'json',
	// default is empty ('').
	OutputFormat string `json:"outputFormat,omitempty" yaml:"outputFormat,omitempty"`
}

// SyncOptions is the options for sync.
type SyncOptions struct {
	CommonOptions `json:",inline" yaml:",inline"`
}

// // setupResult represents the result of a setup operation.
// type setupResult struct {
// 	*pullResult     `json:",inline" yaml:",inline"`
// 	*getStateResult `json:",inline" yaml:",inline"`
// }

// // pullResult represents the result of a pull operation.
// type pullResult struct {
// 	SourceRoot   string `json:"sourceRoot" yaml:"sourceRoot"`
// 	StackAbsPath string `json:"stackAbsPath" yaml:"stackAbsPath"`
// }

// // getStateResult represents the result of a getState operation.
// type getStateResult struct {
// 	StorageKey   string `json:"storageKey" yaml:"storageKey"`
// 	StateAbsPath string `json:"stateAbsPath" yaml:"stateAbsPath"`
// }

// PreviewResult represents the result of a preview operation.
type PreviewResult struct {
	HasChange bool   `json:"hasChange" yaml:"hasChange"`
	Raw       string `json:"raw" yaml:"raw"`
	JSON      string `json:"json,omitempty" yaml:"json,omitempty"`
	Error     error  `json:"error,omitempty" yaml:"error,omitempty" swaggertype:"string"`
}

// NewEmptyPreviewResult creates an empty PreviewResult struct.
func NewEmptyPreviewResult() PreviewResult {
	return PreviewResult{}
}

// NewPreviewResultOnlyError creates a PreviewResult struct containing
// only an error.
func NewPreviewResultOnlyError(err error) PreviewResult {
	return PreviewResult{Error: err}
}

// SyncResult represents the result of a sync operation.
type SyncResult struct {
	Raw   string `json:"raw" yaml:"raw"`
	Error error  `json:"error,omitempty" yaml:"error,omitempty" swaggertype:"string"`
}

// NewEmptySyncResult creates an empty SyncResult struct.
func NewEmptySyncResult() SyncResult {
	return SyncResult{}
}

// NewSyncResultOnlyError creates a SyncResult struct containing
// only an error.
func NewSyncResultOnlyError(err error) SyncResult {
	return SyncResult{Error: err}
}

type (
	// Represents a group of PreviewResult structs
	PreviewResults []PreviewResult
	// Represents a group of SyncResult structs
	SyncResults []SyncResult
)

// Error returns the error of all PreviewResult structs.
func (r PreviewResults) Error() error {
	var errs *multierror.Error
	for _, v := range []PreviewResult(r) {
		errs = multierror.Append(errs, v.Error)
	}

	return errs.ErrorOrNil()
}

// Error returns the error of all SyncResult structs.
func (r SyncResults) Error() error {
	var errs *multierror.Error
	for _, v := range []SyncResult(r) {
		errs = multierror.Append(errs, v.Error)
	}

	return errs.ErrorOrNil()
}
