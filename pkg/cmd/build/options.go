package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	yamlv2 "gopkg.in/yaml.v2"
	"kcl-lang.io/kpm/pkg/api"

<<<<<<< HEAD
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
=======
	"kusionstack.io/kusion/pkg/engine/api/builders"
>>>>>>> b551565 (feat: kusion server, engine api and refactor preview logic)
	"kusionstack.io/kusion/pkg/project"
)

const (
	KclFile = "kcl.yaml"
)

type Options struct {
	IsKclPkg  bool
	Filenames []string
	Flags
}

type Flags struct {
	Output    string
	WorkDir   string
	Settings  []string
	Arguments map[string]string
	NoStyle   bool
	Backend   string
	Workspace string
}

const Stdout = "stdout"

func NewBuildOptions() *Options {
	return &Options{
		Filenames: []string{},
		Flags: Flags{
			Arguments: map[string]string{},
			Settings:  make([]string, 0),
		},
	}
}

func (o *Options) Complete(args []string) error {
	o.Filenames = args
	return o.PreSet(project.IsStack)
}

func (o *Options) Validate() error {
	var wrongFiles []string
	for _, filename := range o.Filenames {
		if filepath.Ext(filename) != ".k" {
			wrongFiles = append(wrongFiles, filename)
		}
	}
	if len(wrongFiles) != 0 {
		return fmt.Errorf("you can only compile files with suffix .k, these are wrong files: %v", wrongFiles)
	}
	return nil
}

func (o *Options) Run() error {
	// Set no style
	if o.NoStyle {
		pterm.DisableStyling()
		pterm.DisableColor()
	}

	// Parse project and stack of work directory
	proj, stack, err := project.DetectProjectAndStack(o.WorkDir)
	if err != nil {
		return err
	}

	// Get workspace from backend
	storage, err := backend.NewWorkspaceStorage(o.Backend)
	if err != nil {
		return err
	}
	ws, err := storage.Get(o.Workspace)
	if err != nil {
		return err
	}

	sp, err := IntentWithSpinner(
		&builders.Options{
			IsKclPkg:  o.IsKclPkg,
			WorkDir:   o.WorkDir,
			Filenames: o.Filenames,
			Settings:  o.Settings,
			Arguments: o.Arguments,
			NoStyle:   o.NoStyle,
		},
		proj,
		stack,
		ws,
	)
	if err != nil {
		return err
	}

	yaml, err := yamlv2.Marshal(sp)
	if err != nil {
		return err
	}
	if o.Output == Stdout || o.Output == "" {
		fmt.Print(string(yaml))
	} else {
		if o.WorkDir != "" {
			o.Output = filepath.Join(o.WorkDir, o.Output)
		}

		err = os.WriteFile(o.Output, yaml, 0o666)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) PreSet(preCheck func(cur string) bool) error {
	curDir := o.WorkDir
	if o.WorkDir == "" {
		curDir, _ = os.Getwd()
	}
	if ok := preCheck(curDir); !ok {
		if o.Output == "" {
			o.Output = Stdout
		}
		return nil
	}

	if _, err := api.GetKclPackage(o.WorkDir); err == nil {
		o.IsKclPkg = true
		return nil
	}

	if len(o.Settings) == 0 {
		o.Settings = []string{KclFile}
	}
	return nil
}
