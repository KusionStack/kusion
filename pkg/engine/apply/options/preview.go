package options

import "kusionstack.io/kusion/pkg/util/terminal"

type PreviewOptions interface {
	Meta

	GetDetail() bool
	GetAll() bool
	GetNoStyle() bool
	GetOutput() string
	GetSpecFile() string
	GetIgnoreFields() []string
	GetValues() []string
	GetUI() *terminal.UI
}
