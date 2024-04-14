// Copyright 2024 KusionStack Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generate

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/cmd/generate/generator"
	"kusionstack.io/kusion/pkg/cmd/generate/run"
	"kusionstack.io/kusion/pkg/cmd/meta"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
	"kusionstack.io/kusion/pkg/util/pretty"
)

var (
	generateLong = i18n.T(`
		This command generates Spec resources with given values, then write the resulting Spec resources to specific output file or stdout.

		The nearest parent folder containing a stack.yaml file is loaded from the project in the current directory.`)

	generateExample = i18n.T(`
		# Generate and write Spec resources to specific output file
		kusion generate -o /tmp/spec.yaml

		# Generate spec with custom workspace
		kusion generate -o /tmp/spec.yaml --workspace dev`)
)

// GenerateFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// GenerateOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type GenerateFlags struct {
	MetaFlags *meta.MetaFlags

	Output string
	Values []string

	genericiooptions.IOStreams
}

// GenerateOptions defines flags and other configuration parameters for the `generate` command.
type GenerateOptions struct {
	*meta.MetaOptions

	Output string
	Values []string

	genericiooptions.IOStreams
}

// NewGenerateFlags returns a default GenerateFlags
func NewGenerateFlags(streams genericiooptions.IOStreams) *GenerateFlags {
	return &GenerateFlags{
		MetaFlags: meta.NewMetaFlags(),
		IOStreams: streams,
	}
}

// NewCmdGenerate creates the `generate` command.
func NewCmdGenerate(ioStreams genericiooptions.IOStreams) *cobra.Command {
	flags := NewGenerateFlags(ioStreams)

	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate and print the resulting Spec resources of target Stack",
		Long:    generateLong,
		Example: generateExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			o, err := flags.ToOptions()
			defer cmdutil.RecoverErr(&err)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run())
			return
		},
	}

	flags.AddFlags(cmd)

	return cmd
}

// AddFlags registers flags for a cli.
func (flags *GenerateFlags) AddFlags(cmd *cobra.Command) {
	// bind flag structs
	flags.MetaFlags.AddFlags(cmd)

	cmd.Flags().StringVarP(&flags.Output, "output", "o", flags.Output, i18n.T("File to write generated Spec resources to"))
	cmd.Flags().StringArrayVar(&flags.Values, "set", []string{}, i18n.T("Set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)"))
}

// ToOptions converts from CLI inputs to runtime inputs.
func (flags *GenerateFlags) ToOptions() (*GenerateOptions, error) {
	// Convert meta options
	metaOptions, err := flags.MetaFlags.ToOptions()
	if err != nil {
		return nil, err
	}

	o := &GenerateOptions{
		MetaOptions: metaOptions,
		Output:      flags.Output,
		Values:      flags.Values,

		IOStreams: flags.IOStreams,
	}

	return o, nil
}

// Validate verifies if GenerateOptions are valid and without conflicts.
func (o *GenerateOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "Unexpected args: %v", args)
	}

	for _, value := range o.Values {
		if parts := strings.SplitN(value, "=", 2); len(parts) != 2 {
			return cmdutil.UsageErrorf(cmd, "value %s is invalid format", value)
		}
	}

	return nil
}

// Run executes the `generate` command.
func (o *GenerateOptions) Run() error {
	// build parameters
	parameters := o.buildParameters()

	// call default generator to generate Spec
	spec, err := GenerateSpecWithSpinner(o.RefProject, o.RefStack, o.RefWorkspace, parameters, true)
	if err != nil {
		return err
	}

	// write Spec to output file or a writer
	return write(spec, o.Output, o.Out)
}

// buildParameters builds parameters with given values.
func (o *GenerateOptions) buildParameters() map[string]string {
	parameters := make(map[string]string)

	for _, value := range o.Values {
		parts := strings.SplitN(value, "=", 2)
		parameters[parts[0]] = parts[1]
	}

	return parameters
}

// GenerateSpecWithSpinner calls generator to generate versioned Spec.
// Add a method wrapper for testing purposes.
func GenerateSpecWithSpinner(
	project *v1.Project,
	stack *v1.Stack,
	workspace *v1.Workspace,
	parameters map[string]string,
	noStyle bool,
) (*v1.Spec, error) {
	// Construct generator instance
	defaultGenerator := &generator.DefaultGenerator{
		Project:   project,
		Stack:     stack,
		Workspace: workspace,
		Runner:    &run.KPMRunner{},
	}

	var sp *pterm.SpinnerPrinter
	if noStyle {
		fmt.Printf("Generating Spec in the Stack %s...\n", stack.Name)
	} else {
		sp = &pretty.SpinnerT
		sp, _ = sp.Start(fmt.Sprintf("Generating Spec in the Stack %s...", stack.Name))
	}

	// style means color and prompt here. Currently, sp will be nil only when o.NoStyle is true
	style := !noStyle && sp != nil

	versionedSpec, err := defaultGenerator.Generate(stack.Path, parameters)
	if err != nil {
		if style {
			sp.Fail()
			return nil, err
		} else {
			return nil, err
		}
	}

	// success
	if style {
		sp.Success()
	} else {
		fmt.Println()
	}

	return versionedSpec, nil
}

func SpecFromFile(filePath string) (*v1.Spec, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// TODO: here we use decoder in yaml.v3 to parse resources because it converts
	// map into map[string]interface{} by default which is inconsistent with yaml.v2.
	// The use of yaml.v2 and yaml.v3 should be unified in the future.
	decoder := yamlv3.NewDecoder(bytes.NewBuffer(b))
	decoder.KnownFields(true)
	i := &v1.Spec{}
	if err = decoder.Decode(i); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse the intent file, please check if the file content is valid")
	}
	return i, nil
}

// write writes Spec resources to a file or a writer.
func write(spec *v1.Spec, output string, out io.Writer) error {
	specStr, err := yamlv3.Marshal(spec)
	if err != nil {
		return err
	}

	switch {
	case output == "":
		_, err := fmt.Fprintln(out, specStr)
		return err
	default:
		return dumpToFile(string(specStr), output)
	}
}

func dumpToFile(specStr string, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("opening file for writing Spec: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(specStr + "\n")
	return err
}
