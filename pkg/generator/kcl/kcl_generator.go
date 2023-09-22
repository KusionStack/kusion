package kcl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	kcl "kcl-lang.io/kcl-go"
	kclpkg "kcl-lang.io/kcl-go/pkg/kcl"
	"kcl-lang.io/kcl-go/pkg/spec/gpyrpc"
	"kcl-lang.io/kpm/pkg/api"
	"kcl-lang.io/kpm/pkg/opt"

	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/generator/kcl/rest"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/resources/crd"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/pkg/util/yaml"
)

type Generator struct{}

var (
	_          generator.Generator = (*Generator)(nil)
	enableRest bool
)

const IncludeSchemaTypePath = "include_schema_type_path"

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

func (g *Generator) GenerateSpec(o *generator.Options, _ *projectstack.Project, stack *projectstack.Stack) (*models.Spec, error) {
	compileResult, err := Run(o, stack)
	if err != nil {
		return nil, err
	}

	// convert Run result to spec
	spec, err := engine.KCLResult2Spec(compileResult.Documents)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func Run(o *generator.Options, stack *projectstack.Stack) (*CompileResult, error) {
	optList, err := BuildOptions(o.WorkDir, o.Settings, o.Overrides, o.Arguments, o.DisableNone, o.OverrideAST)
	if err != nil {
		return nil, err
	}
	log.Debugf("Compile filenames: %v", o.Filenames)
	log.Debugf("Compile options: %s", jsonutil.MustMarshal2PrettyString(optList))

	var result *kcl.KCLResultList
	if o.IsKclPkg {
		result, err = api.RunPkgWithOpt(
			&opt.CompileOptions{
				Option: kclpkg.NewOption().Merge(optList...),
			},
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

	crdObjs, err := readCRDs(workDir)
	if err != nil {
		return err
	}
	if len(crdObjs) != 0 {
		// Append to Documents
		for _, obj := range crdObjs {
			if doc, flag := obj.(map[string]interface{}); flag {
				resource, err := k8sResource2ResourceMap(doc)
				if err != nil {
					return err
				}
				r.Documents = append(r.Documents, resource)
			}
		}

		// Update RawYAMLResult
		items := make([]interface{}, len(r.Documents))
		for i, doc := range r.Documents {
			items[i] = doc
		}
		r.RawYAMLResult = yaml.MergeToOneYAML(items...)
	}

	return nil
}

func readCRDs(workDir string) ([]interface{}, error) {
	projectPath := path.Dir(workDir)
	crdPath := path.Join(projectPath, crd.Directory)
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

func BuildOptions(
	workDir string,
	settings, overrides []string,
	arguments map[string]string,
	disableNone, overrideAST bool,
) ([]kcl.Option, error) {
	optList := make([]kcl.Option, 0)
	// build settings option
	for _, setting := range settings {
		if workDir != "" {
			setting = filepath.Join(workDir, setting)
		}
		opt := kcl.WithSettings(setting)
		if opt.Err != nil {
			return nil, opt.Err
		}

		optList = append(optList, opt)
	}

	// build arguments option
	for k, v := range arguments {
		argStr := k + "=" + v
		opt := kcl.WithOptions(argStr)
		if opt.Err != nil {
			return nil, opt.Err
		}

		optList = append(optList, opt)
	}

	// build overrides option
	opt := kcl.WithOverrides(overrides...)
	if opt.Err != nil {
		return nil, opt.Err
	}

	optList = append(optList, opt)

	// build disable none option
	opt = kcl.WithDisableNone(disableNone)
	if opt.Err != nil {
		return nil, opt.Err
	}

	optList = append(optList, opt)

	// open PrintOverride option
	opt = kcl.WithPrintOverridesAST(overrideAST)
	if opt.Err != nil {
		return nil, opt.Err
	}

	optList = append(optList, opt)

	// build workDir option
	opt = kcl.WithWorkDir(workDir)
	if opt.Err != nil {
		return nil, opt.Err
	}
	optList = append(optList, opt)

	if arguments[IncludeSchemaTypePath] == "true" {
		opt = kcl.WithIncludeSchemaTypePath(true)
		if opt.Err != nil {
			return nil, opt.Err
		}
		optList = append(optList, opt)
	}

	return optList, nil
}

func normResult(resp *gpyrpc.ExecProgram_Result) (*CompileResult, error) {
	var result CompileResult
	if strings.TrimSpace(resp.JsonResult) == "" {
		return &result, nil
	}

	var mList []map[string]interface{}
	if err := json.Unmarshal([]byte(resp.JsonResult), &mList); err != nil {
		return nil, err
	}

	if len(mList) == 0 {
		return nil, fmt.Errorf("normResult: invalid result: %s", resp.JsonResult)
	}

	var kclResults []kcl.KCLResult
	for _, m := range mList {
		if len(m) != 0 {
			kclResults = append(kclResults, m)
		}
	}

	return &CompileResult{
		Documents: kclResults,
	}, nil
}

// Simply call kcl cmd
func CompileUsingCmd(sourceKclFiles []string, targetFile string, args map[string]string, settings []string) (string, string, error) {
	kclArgs := []string{
		genKclArgs(args, settings), "-n", "-o", targetFile,
	}
	kclArgs = append(kclArgs, sourceKclFiles...)
	cmd := exec.Command(kclAppPath, kclArgs...)
	cmd.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func genKclArgs(args map[string]string, settings []string) string {
	kclArgs := ""
	for key, value := range args {
		kclArgs += fmt.Sprintf("-D %s=%s ", key, value)
	}
	if len(settings) > 0 {
		kclArgs += fmt.Sprintf("-Y %s ", strings.Join(settings, " "))
	}
	return kclArgs
}

func Overwrite(fileName string, overrides []string) (bool, error) {
	return kcl.OverrideFile(fileName, overrides, []string{})
}

// Get kcl cli path
func GetKclPath() string {
	return kclAppPath
}

// Get kclvm cli path
func GetKclvmPath() string {
	return filepath.Join(filepath.Dir(kclAppPath), "kclvm")
}
