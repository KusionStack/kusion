package version

import (
	"errors"
	"fmt"
	"strings"

	"kusionstack.io/kusion/pkg/version"
)

const jsonOutput = "json"

type VersionOptions struct {
	Output string
}

func NewVersionOptions() *VersionOptions {
	return &VersionOptions{}
}

func (o *VersionOptions) Validate() error {
	if o.Output != "" && o.Output != jsonOutput {
		return errors.New("invalid output type, output must be 'json'")
	}
	return nil
}

func (o *VersionOptions) Run() {
	if strings.ToLower(o.Output) == jsonOutput {
		fmt.Println(version.JSON())
	} else {
		fmt.Println(version.String())
	}
}
