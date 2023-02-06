package version

import (
	"fmt"

	"kusionstack.io/kusion/pkg/version"
)

type VersionOptions struct {
	ExportJSON bool
	ExportYAML bool
	Short      bool
}

func NewVersionOptions() *VersionOptions {
	return &VersionOptions{}
}

func (o *VersionOptions) Complete() {
	if !(o.ExportYAML || o.ExportJSON || o.Short) {
		o.ExportYAML = true
	}
}

func (o *VersionOptions) Validate() error {
	if (o.ExportJSON && o.ExportYAML) || (o.ExportJSON && o.Short) || (o.ExportYAML && o.Short) {
		return fmt.Errorf("invalid options")
	}

	return nil
}

func (o *VersionOptions) Run() {
	switch {
	case o.ExportJSON:
		fmt.Println(version.JSON())
	case o.ExportYAML:
		fmt.Println(version.YAML())
	case o.Short:
		fmt.Println(version.ShortString())
	default:
		fmt.Println(version.String())
	}
}
