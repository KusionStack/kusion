package kcl

import (
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	kcl "kcl-lang.io/kcl-go"
	"kcl-lang.io/kcl-go/pkg/spec/gpyrpc"

	"kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/cmd/build/builders"
	"kusionstack.io/kusion/pkg/cmd/build/builders/kcl/rest"
)

func TestInit(t *testing.T) {
	mockey.PatchConvey("success", t, func() {
		mockNew(nil)
		err := Init()
		assert.Nil(t, err)
	})
	mockey.PatchConvey("failed", t, func() {
		mockNew(assert.AnError)
		err := Init()
		assert.NotNil(t, err)
	})
}

func TestGenerateIntent(t *testing.T) {
	fakeStack := &v1.Stack{
		Name: "fake-stack",
		Path: filepath.Join(".", "testdata"),
	}

	type args struct {
		workDir   string
		filenames []string
		settings  []string
		arguments map[string]string
	}
	testArgs := args{
		filenames: []string{},
		settings:  []string{"testdata/kcl.yaml"},
		arguments: map[string]string{"image": "nginx:latest"},
	}
	tests := []struct {
		name    string
		args    args
		want    *intent.Intent
		wantErr bool
		prefunc func()
	}{
		{
			name:    "success",
			args:    testArgs,
			want:    &intent.Intent{Resources: []intent.Resource{}},
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
		mockey.PatchConvey(tt.name, t, func() {
			if tt.prefunc != nil {
				tt.prefunc()
			}

			g := &Builder{}
			got, err := g.Build(&builders.Options{
				WorkDir:   tt.args.workDir,
				Filenames: tt.args.filenames,
				Settings:  tt.args.settings,
				Arguments: tt.args.arguments,
			}, nil, fakeStack)
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
	mockey.PatchConvey("t1", t, func() {
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
				Documents: []kcl.KCLResult{map[string]interface{}{"a": "b"}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
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
	mockey.Mock(mockey.GetMethod(new(exec.Cmd), "Run")).To(func(_ *exec.Cmd) error {
		return nil
	}).Build()
	_, _, err := CompileUsingCmd([]string{}, "", map[string]string{"a": "b"}, []string{"kcl.yaml"})
	assert.Nil(t, err)
}

func TestOverwrite(t *testing.T) {
	mockey.Mock(kcl.OverrideFile).To(func(filename string, _, _ []string) (bool, error) {
		return false, nil
	}).Build()
	_, err := Overwrite("", []string{})
	assert.Nil(t, err)
}

func mockNew(mockErr error) {
	mockey.Mock(rest.New).To(func() (*rest.Client, error) {
		return nil, mockErr
	}).Build()
}

func mockRunFiles(mockErr error) {
	mockey.Mock(kcl.RunFiles).To(func(paths []string, opts ...kcl.Option) (*kcl.KCLResultList, error) {
		return &kcl.KCLResultList{}, mockErr
	}).Build()
}

func Test_appendCRDs(t *testing.T) {
	mockey.PatchConvey("append one CRD", t, func() {
		cs := &CompileResult{}
		err := appendCRDs("./testdata/crd", cs)
		assert.Nil(t, err)
		assert.NotNil(t, cs.Documents)
		assert.NotEmpty(t, cs.RawYAMLResult)
	})

	mockey.PatchConvey("no CRD to append", t, func() {
		cs := &CompileResult{}
		err := appendCRDs("./testdata", cs)
		assert.Nil(t, err)
		assert.Nil(t, cs.Documents)
		assert.Empty(t, cs.RawYAMLResult)
	})
}

func Test_readCRDsIfExists(t *testing.T) {
	mockey.PatchConvey("read CRDs", t, func() {
		crds, err := readCRDs("./testdata/crd")
		assert.Nil(t, err)
		assert.NotNil(t, crds)
	})
	mockey.PatchConvey("no CRDs", t, func() {
		crds, err := readCRDs("./testdata")
		assert.Nil(t, err)
		assert.Nil(t, crds)
	})
}
