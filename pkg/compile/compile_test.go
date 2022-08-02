//go:build !arm64
// +build !arm64

// Provide general KCL compilation method
package compile

import (
	"os/exec"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	kcl "kusionstack.io/kclvm-go"
	k2 "kusionstack.io/kclvm-go/pkg/kcl"
	"kusionstack.io/kclvm-go/pkg/spec/gpyrpc"

	"kusionstack.io/kusion/pkg/compile/rest"
)

func TestInit(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockNew(nil)
		defer monkey.UnpatchAll()
		err := Init()
		assert.Nil(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		mockNew(assert.AnError)
		defer monkey.UnpatchAll()
		err := Init()
		assert.NotNil(t, err)
	})
}

func TestCompile(t *testing.T) {
	defer monkey.UnpatchAll()
	type args struct {
		workDir     string
		filenames   []string
		settings    []string
		arguments   []string
		overrides   []string
		disableNone bool
		overrideAST bool
	}
	testArgs := args{
		filenames:   []string{},
		settings:    []string{"testdata/kcl.yaml"},
		arguments:   []string{"image=nginx:latest"},
		disableNone: true,
	}
	tests := []struct {
		name    string
		args    args
		want    *CompileResult
		wantErr bool
		prefunc func()
	}{
		{
			name:    "success",
			args:    testArgs,
			want:    NewCompileResult(&kcl.KCLResultList{}),
			wantErr: false,
			prefunc: func() { mockRunFiles(nil) },
		},
		{
			name:    "failed",
			args:    testArgs,
			want:    nil,
			wantErr: true,
			prefunc: func() { mockRunFiles(assert.AnError) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prefunc != nil {
				tt.prefunc()
			}
			got, err := Compile(tt.args.workDir, tt.args.filenames, tt.args.settings, tt.args.arguments, tt.args.overrides, tt.args.disableNone, tt.args.overrideAST)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Compile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnableRPC(t *testing.T) {
	t.Run("t1", func(t *testing.T) {
		result := EnableRPC()
		assert.False(t, result)
	})
}

func Test_normResult(t *testing.T) {
	type args struct {
		resp *gpyrpc.ExecProgram_Result
	}
	tests := []struct {
		name    string
		args    args
		want    *CompileResult
		wantErr bool
	}{
		{
			name: "empty json",
			args: args{
				resp: &gpyrpc.ExecProgram_Result{},
			},
			want:    &CompileResult{},
			wantErr: false,
		},
		{
			name: "unmarshal error",
			args: args{
				resp: &gpyrpc.ExecProgram_Result{
					JsonResult: `{"a": b}`,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unmarshal len 0",
			args: args{
				resp: &gpyrpc.ExecProgram_Result{
					JsonResult: `[]`,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				resp: &gpyrpc.ExecProgram_Result{
					JsonResult: `[{"a": "b"}]`,
				},
			},
			want: &CompileResult{
				Documents: []k2.KCLResult{map[string]interface{}{"a": "b"}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normResult(tt.args.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("normResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompileUsingCmd(t *testing.T) {
	monkey.PatchInstanceMethod(reflect.TypeOf(new(exec.Cmd)), "Run", func(_ *exec.Cmd) error {
		return nil
	})
	defer monkey.UnpatchAll()
	_, _, err := CompileUsingCmd([]string{}, "", map[string]string{"a": "b"}, []string{"kcl.yaml"})
	assert.Nil(t, err)
}

func TestOverwrite(t *testing.T) {
	monkey.Patch(kcl.OverrideFile, func(filename string, _, _ []string) (bool, error) {
		return false, nil
	})
	defer monkey.UnpatchAll()
	_, err := Overwrite("", []string{})
	assert.Nil(t, err)
}

func mockNew(mockErr error) {
	monkey.Patch(rest.New, func() (*rest.Client, error) {
		return nil, mockErr
	})
}

func mockRunFiles(mockErr error) {
	monkey.Patch(kcl.RunFiles, func(paths []string, opts ...kcl.Option) (*kcl.KCLResultList, error) {
		return &kcl.KCLResultList{}, mockErr
	})
}

// func mockRestCompile(client *rest.Client) {
// 	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Compile", func(_ *rest.Client, req *gpyrpc.ExecProgram_Args) (*rest.Result, error) {
// 		return new(rest.Result), nil
// 	})
// }

// func mockBuildOptions() {
// 	monkey.Patch(buildOptions, func(workDir string, settings, arguments, overrides []string, disableNone, overrideAST bool) ([]kcl.Option, error) {
// 		return []kcl.Option{}, nil
// 	})
// }

// func mockParseArgs() {
// 	monkey.Patch(k2.ParseArgs, func(pathList []string, opts ...kcl.Option) (kcl.Option, error) {
// 		return kcl.Option{}, nil
// 	})
// }

// func mockNormResult() {
// 	monkey.Patch(normResult, func(resp *gpyrpc.ExecProgram_Result) (*CompileResult, error) {
// 		return &CompileResult{}, nil
// 	})
// }

func Test_appendCRDs(t *testing.T) {
	t.Run("append one CRD", func(t *testing.T) {
		cs := &CompileResult{}
		err := appendCRDs("./testdata/crd", cs)
		assert.Nil(t, err)
		assert.NotNil(t, cs.Documents)
		assert.NotEmpty(t, cs.RawYAMLResult)
	})

	t.Run("no CRD to append", func(t *testing.T) {
		cs := &CompileResult{}
		err := appendCRDs("./testdata", cs)
		assert.Nil(t, err)
		assert.Nil(t, cs.Documents)
		assert.Empty(t, cs.RawYAMLResult)
	})
}

func Test_readCRDsIfExists(t *testing.T) {
	t.Run("read CRDs", func(t *testing.T) {
		crds, err := readCRDs("./testdata/crd")
		assert.Nil(t, err)
		assert.NotNil(t, crds)
	})
	t.Run("no CRDs", func(t *testing.T) {
		crds, err := readCRDs("./testdata")
		assert.Nil(t, err)
		assert.Nil(t, crds)
	})
}
