// Provide general KCL compilation method
package compile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pterm/pterm"
	kcl "kusionstack.io/kclvm-go"
	"kusionstack.io/kclvm-go/pkg/spec/gpyrpc"

	"kusionstack.io/kusion/pkg/compile/rest"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/manifest"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/projectstack"
	jsonUtil "kusionstack.io/kusion/pkg/util/json"
)

var (
	restClient *rest.Client
	enableRest bool
)

func Init() error {
	c, err := rest.New()
	if err != nil {
		return err
	}
	restClient = c
	enableRest = true
	return nil
}

func EnableRPC() bool {
	return !enableRest
}

func CompileWithSpinner(workDir string, filenames, settings, arguments, overrides []string,
	stack *projectstack.Stack,
) (*manifest.Manifest, *pterm.SpinnerPrinter, error) {
	// Spinner
	sp := pterm.DefaultSpinner.
		WithSequence("⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ ").
		WithDelay(time.Millisecond * 100)

	sp, _ = sp.Start(fmt.Sprintf("Compiling in stack %s...", stack.Name))

	// compile by kcl go sdk
	r, err := Compile(workDir, filenames, settings, arguments, overrides, true, false)
	if err != nil {
		return nil, sp, err
	}
	// Construct resource from compile result to build request
	resources, err := engine.ConvertKCLResult2Resources(r.Documents)
	if err != nil {
		return nil, sp, err
	}

	return resources, sp, nil
}

// Compile General KCL compilation method
func Compile(workDir string, filenames, settings, arguments, overrides []string, disableNone bool, overrideAST bool) (*CompileResult, error) {
	optList, err := buildOptions(workDir, settings, arguments, overrides, disableNone, overrideAST)
	if err != nil {
		return nil, err
	}

	log.Debugf("Compile filenames: %v", filenames)
	log.Debugf("Compile options: %s", jsonUtil.MustMarshal2PrettyString(optList))

	// call kcl run
	result, err := kcl.RunFiles(filenames, optList...)
	if err != nil {
		return nil, err
	}
	return NewCompileResult(result), nil
}

func buildOptions(workDir string, settings, arguments, overrides []string, disableNone, overrideAST bool) ([]kcl.Option, error) {
	var optList []kcl.Option
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
	for _, arg := range arguments {
		opt := kcl.WithOptions(arg)
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
