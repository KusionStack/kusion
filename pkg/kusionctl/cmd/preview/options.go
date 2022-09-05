package preview

import (
	applycmd "kusionstack.io/kusion/pkg/kusionctl/cmd/apply"
	compilecmd "kusionstack.io/kusion/pkg/kusionctl/cmd/compile"
)

type PreviewOptions struct {
	compilecmd.CompileOptions
	Yes          bool
	Detail       bool
	NoStyle      bool
	IgnoreFields []string
}

func NewPreviewOptions() *PreviewOptions {
	return &PreviewOptions{
		CompileOptions: compilecmd.CompileOptions{
			Filenames: []string{},
			Arguments: []string{},
			Settings:  []string{},
			Overrides: []string{},
		},
	}
}

func (o *PreviewOptions) Complete(args []string) {
	o.CompileOptions.Complete(args)
}

func (o *PreviewOptions) Validate() error {
	return o.CompileOptions.Validate()
}

func (o *PreviewOptions) Run() error {
	applyOptions := applycmd.ApplyOptions{
		CompileOptions: o.CompileOptions,
		Yes:            o.Yes,
		Detail:         o.Detail,
		NoStyle:        o.NoStyle,
		OnlyPreview:    true,
	}

	return applyOptions.Run()
}
