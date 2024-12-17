package run

import (
	"os"
	"path/filepath"
	"runtime"

	kcl "kcl-lang.io/kcl-go"
	kclpkg "kcl-lang.io/kcl-go/pkg/kcl"
	"kcl-lang.io/kpm/pkg/client"
	"kcl-lang.io/kpm/pkg/downloader"
	"kcl-lang.io/kpm/pkg/opt"
)

// CodeRunner compiles and runs the target DSL based configuration code
// and returns configuration data in plain format.
type CodeRunner interface {
	Run(workingDir string, arguments map[string]string) ([]byte, error)
}

// KPMRunner should implement the CodeRunner interface.
var _ CodeRunner = &KPMRunner{}

// KPMRunner implements the CodeRunner interface.
type KPMRunner struct {
	Host     string
	Username string
	Password string
}

// Run calls KPM api to compile and run KCL based configuration code.
func (r *KPMRunner) Run(workDir string, arguments map[string]string) (res []byte, err error) {
	cacheDir := filepath.Join(workDir, ".kclvm")
	defer func(path string) {
		if err != nil {
			return
		}
		err = os.RemoveAll(path)
	}(cacheDir)

	optList, err := buildKCLOptions(workDir, arguments)
	if err != nil {
		return nil, err
	}
	cli, err := client.NewKpmClient()
	if err != nil {
		return nil, err
	}
	cli.DepDownloader = downloader.NewOciDownloader(runtime.GOOS + "/" + runtime.GOARCH)

	// Login to the private oci registry.
	if r.Host != "" && r.Username != "" && r.Password != "" {
		if err = cli.LoginOci(r.Host, r.Username, r.Password); err != nil {
			return nil, err
		}
	}

	result, err := cli.RunWithOpts(
		opt.WithKclOption(*kclpkg.NewOption().Merge(optList...)),
		opt.WithNoSumCheck(false),
		opt.WithLogWriter(nil),
	)
	if err != nil {
		return nil, err
	}

	return []byte(result.GetRawYamlResult()), nil
}

// buildKCLOptions returns list of KCL options.
func buildKCLOptions(workDir string, arguments map[string]string) ([]kcl.Option, error) {
	optList := make([]kcl.Option, 3)

	// build arguments option
	for k, v := range arguments {
		argStr := k + "=" + v
		withOpt := kcl.WithOptions(argStr)
		optList = append(optList, withOpt)
	}

	// build workDir option
	withOpt := kcl.WithWorkDir(workDir)
	optList = append(optList, withOpt)

	// eliminate null values in the result
	withOpt = kcl.WithDisableNone(true)
	if withOpt.Err != nil {
		return nil, withOpt.Err
	}
	optList = append(optList, withOpt)

	// holds include schema type path
	withOpt = kcl.WithFullTypePath(true)
	if withOpt.Err != nil {
		return nil, withOpt.Err
	}
	optList = append(optList, withOpt)

	return optList, nil
}
