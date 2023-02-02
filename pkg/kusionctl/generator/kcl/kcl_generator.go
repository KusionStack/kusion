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

	"kusionstack.io/kclvm-go"
	"kusionstack.io/kclvm-go/pkg/spec/gpyrpc"

	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/kusionctl/generator"
	"kusionstack.io/kusion/pkg/kusionctl/generator/kcl/rest"
	"kusionstack.io/kusion/pkg/log"
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

func (g *Generator) GenerateSpec(o *generator.Options, stack *projectstack.Stack) (*models.Spec, error) {
	optList, err := buildOptions(o.WorkDir, o.Settings, o.Arguments, o.Overrides, o.DisableNone, o.OverrideAST)
	if err != nil {
		return nil, err
	}

	log.Debugf("Compile filenames: %v", o.Filenames)
	log.Debugf("Compile options: %s", jsonutil.MustMarshal2PrettyString(optList))

	// call kcl run
	result, err := kclvm.RunFiles(o.Filenames, optList...)
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

	// convert compile result to spec
	spec, err := engine.ResourcesYAML2Spec(compileResult.Documents)
	if err != nil {
		return nil, err
	}
	return spec, nil
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

func buildOptions(workDir string, settings, arguments, overrides []string, disableNone, overrideAST bool) ([]kclvm.Option, error) {
	optList := []kclvm.Option{}
	// build settings option
	for _, setting := range settings {
		if workDir != "" {
			setting = filepath.Join(workDir, setting)
		}
		opt := kclvm.WithSettings(setting)
		if opt.Err != nil {
			return nil, opt.Err
		}

		optList = append(optList, opt)
	}

	// build arguments option
	for _, arg := range arguments {
		opt := kclvm.WithOptions(arg)
		if opt.Err != nil {
			return nil, opt.Err
		}

		optList = append(optList, opt)
	}

	// build overrides option
	opt := kclvm.WithOverrides(overrides...)
	if opt.Err != nil {
		return nil, opt.Err
	}

	optList = append(optList, opt)

	// build disable none option
	opt = kclvm.WithDisableNone(disableNone)
	if opt.Err != nil {
		return nil, opt.Err
	}

	optList = append(optList, opt)

	// open PrintOverride option
	opt = kclvm.WithPrintOverridesAST(overrideAST)
	if opt.Err != nil {
		return nil, opt.Err
	}

	optList = append(optList, opt)

	// build workDir option
	opt = kclvm.WithWorkDir(workDir)
	if opt.Err != nil {
		return nil, opt.Err
	}
	optList = append(optList, opt)

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

	var kclResults []kclvm.KCLResult
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
	return kclvm.OverrideFile(fileName, overrides, []string{})
}

// Get kcl cli path
func GetKclPath() string {
	return kclAppPath
}

// Get kclvm cli path
func GetKclvmPath() string {
	return filepath.Join(filepath.Dir(kclAppPath), "kclvm")
}
