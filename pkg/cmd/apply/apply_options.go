package apply

import (
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"kusionstack.io/kusion/pkg/cmd/preview"
)

// ApplyOptions defines flags and other configuration parameters for the `apply` command.
type ApplyOptions struct {
	*preview.PreviewOptions

	Yes         bool
	DryRun      bool
	Watch       bool
	Timeout     int
	PortForward int

	genericiooptions.IOStreams
}

func (o *ApplyOptions) GetYes() bool {
	return o.Yes
}

func (o *ApplyOptions) GetDryRun() bool {
	return o.DryRun
}

func (o *ApplyOptions) GetWatch() bool {
	return o.Watch
}

func (o *ApplyOptions) GetTimeout() int {
	return o.Timeout
}

func (o *ApplyOptions) GetPortForward() int {
	return o.PortForward
}
