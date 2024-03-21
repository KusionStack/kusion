package kcl

import (
	"os"
	"path/filepath"

	kcl "kcl-lang.io/kcl-go"
	kclpkg "kcl-lang.io/kcl-go/pkg/kcl"
	"kcl-lang.io/kpm/pkg/api"
	"kcl-lang.io/kpm/pkg/opt"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
	"kusionstack.io/kusion/pkg/cmd/build/builders/crd"
	"kusionstack.io/kusion/pkg/cmd/build/builders/kcl/rest"
	"kusionstack.io/kusion/pkg/log"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
)

type Builder struct{}

var enableRest bool

const (
	MaxLogLength          = 3751
	IncludeSchemaTypePath = "include_schema_type_path"
)

func Init() error {
	_, err := rest.New()
	if err != nil {
		return err
	}
	enableRest = true
	return nil
}

func EnableRPC() bool {
	return !enableRest
}

func Run(o *builders.Options, stack *v1.Stack) (*CompileResult, error) {
	optList, err := BuildKCLOptions(o)
	if err != nil {
		return nil, err
	}
	log.Debugf("Compile filenames: %v", o.Filenames)
	log.Debugf("Compile options: %s", jsonutil.MustMarshal2PrettyString(optList))

	var result *kcl.KCLResultList
	if o.KclPkg != nil {
		result, err = api.RunWithOpts(
			opt.WithKclOption(*kclpkg.NewOption().Merge(optList...)),
			opt.WithNoSumCheck(true),
			opt.WithLogWriter(nil),
		)
	} else {
		// call kcl run
		log.Debug("The current directory is not a KCL Package, use kcl run instead")
		result, err = kcl.RunFiles(o.Filenames, optList...)
	}
	if err != nil {
		return nil, err
	}
	compileResult := NewCompileResult(result)

	// Append crd description to compiled result,
	// workDir may omit empty if run in stack dir
	err = appendCRDs(stack.Path, compileResult)
	if err != nil {
		return nil, err
	}
	return compileResult, err
}

func appendCRDs(workDir string, r *CompileResult) error {
	if r == nil {
		return nil
	}

	_, err := readCRDs(workDir)
	if err != nil {
		return err
	}

	return nil
}

func readCRDs(workDir string) ([]interface{}, error) {
	projectPath := filepath.Dir(workDir)
	crdPath := filepath.Join(projectPath, crd.Directory)
	_, err := os.Stat(crdPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	visitor := crd.NewVisitor(crdPath)
	return visitor.Visit()
}

func BuildKCLOptions(o *builders.Options) ([]kcl.Option, error) {
	optList := make([]kcl.Option, 0)
	settings := o.Settings
	workDir := o.WorkDir
	arguments := o.Arguments

	// build settings option
	for _, setting := range settings {
		if workDir != "" {
			setting = filepath.Join(workDir, setting)
		}
		settingOptions := kcl.WithSettings(setting)
		if settingOptions.Err != nil {
			return nil, settingOptions.Err
		}
		optList = append(optList, settingOptions)
	}

	// build arguments option
	for k, v := range arguments {
		argStr := k + "=" + v
		withOpt := kcl.WithOptions(argStr)
		if withOpt.Err != nil {
			return nil, withOpt.Err
		}

		optList = append(optList, withOpt)
	}

	// build workDir option
	withOpt := kcl.WithWorkDir(workDir)
	if withOpt.Err != nil {
		return nil, withOpt.Err
	}
	optList = append(optList, withOpt)

	// eliminate null values in the result
	withOpt = kcl.WithDisableNone(true)
	if withOpt.Err != nil {
		return nil, withOpt.Err
	}
	optList = append(optList, withOpt)

	if arguments[IncludeSchemaTypePath] == "true" {
		withOpt = kcl.WithFullTypePath(true)
		if withOpt.Err != nil {
			return nil, withOpt.Err
		}
		optList = append(optList, withOpt)
	}

	return optList, nil
}

func Overwrite(fileName string, overrides []string) (bool, error) {
	return kcl.OverrideFile(fileName, overrides, []string{})
}
