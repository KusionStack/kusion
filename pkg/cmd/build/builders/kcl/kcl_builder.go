package kcl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	kcl "kcl-lang.io/kcl-go"
	kclpkg "kcl-lang.io/kcl-go/pkg/kcl"
	"kcl-lang.io/kcl-go/pkg/spec/gpyrpc"
	"kcl-lang.io/kpm/pkg/api"
	"kcl-lang.io/kpm/pkg/opt"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
	"kusionstack.io/kusion/pkg/cmd/build/builders/crd"
	"kusionstack.io/kusion/pkg/cmd/build/builders/kcl/rest"
	"kusionstack.io/kusion/pkg/log"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/pkg/util/yaml"
)

type Builder struct{}

var (
	_          builders.Builder = (*Builder)(nil)
	enableRest bool
)

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

func (g *Builder) Build(o *builders.Options, _ *project.Project, stack *stack.Stack) (*intent.Intent, error) {
	compileResult, err := Run(o, stack)
	if err != nil {
		return nil, err
	}

	// convert Run result to i
	i, err := KCLResult2Intent(compileResult.Documents)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func Run(o *builders.Options, stack *stack.Stack) (*CompileResult, error) {
	optList, err := BuildOptions(o.WorkDir, o.Arguments)
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

func BuildOptions(workDir string, arguments map[string]string) ([]kcl.Option, error) {
	optList := make([]kcl.Option, 0)

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
		withOpt = kcl.WithIncludeSchemaTypePath(true)
		if withOpt.Err != nil {
			return nil, withOpt.Err
		}
		optList = append(optList, withOpt)
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

// CompileUsingCmd simply call kcl cmd
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

func KCLResult2Intent(kclResults []kcl.KCLResult) (*intent.Intent, error) {
	resources := make([]intent.Resource, len(kclResults))

	for i, result := range kclResults {
		// Marshal kcl result to bytes
		bytes, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}

		msg := string(bytes)
		if len(msg) > MaxLogLength {
			msg = msg[0:MaxLogLength]
		}

		log.Infof("convert kcl result to resource: %s", msg)

		// Parse json data as models.Resource
		var item intent.Resource
		if err = json.Unmarshal(bytes, &item); err != nil {
			return nil, err
		}
		resources[i] = item
	}

	return &intent.Intent{Resources: resources}, nil
}
