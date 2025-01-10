package preview

import (
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"kusionstack.io/kusion/pkg/cmd/meta"
	"kusionstack.io/kusion/pkg/util/terminal"
)

// PreviewOptions defines flags and other configuration parameters for the `preview` command.
type PreviewOptions struct {
	*meta.MetaOptions

	Detail       bool
	All          bool
	NoStyle      bool
	Output       string
	SpecFile     string
	IgnoreFields []string
	Values       []string

	UI *terminal.UI

	genericiooptions.IOStreams
}

func (o *PreviewOptions) GetDetail() bool {
	return o.Detail
}

func (o *PreviewOptions) GetAll() bool {
	return o.All
}

func (o *PreviewOptions) GetNoStyle() bool {
	return o.NoStyle
}

func (o *PreviewOptions) GetOutput() string {
	return o.Output
}

func (o *PreviewOptions) GetSpecFile() string {
	return o.SpecFile
}

func (o *PreviewOptions) GetIgnoreFields() []string {
	return o.IgnoreFields
}

func (o *PreviewOptions) GetValues() []string {
	return o.Values
}

func (o *PreviewOptions) GetUI() *terminal.UI {
	return o.UI
}
