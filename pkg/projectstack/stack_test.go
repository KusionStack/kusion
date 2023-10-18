//go:build !arm64
// +build !arm64

package projectstack

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"

	"kusionstack.io/kusion/pkg/util/json"
)

func TestFindStackPath(t *testing.T) {
	_ = os.Chdir(filepath.Join(TestStackPathAA, "ci-test"))
	defer os.Chdir(TestCurrentDir)

	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			want:    filepath.Join(TestCurrentDir, TestStackPathAA),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindStackPath()
			if (err != nil) != tt.wantErr {
				t.Errorf("FindStackPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindStackPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindStackPathFrom(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: "./testdata/appops/http-echo/dev/ci-test",
			},
			want:    "testdata/appops/http-echo/dev",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindStackPathFrom(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindStackPathFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindStackPathFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStack(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is-stack",
			args: args{
				path: "./testdata/appops/http-echo/dev/ci-test",
			},
			want: false,
		},
		{
			name: "is-not-stack",
			args: args{
				path: "./testdata/appops/http-echo/dev",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStack(tt.args.path); got != tt.want {
				t.Errorf("IsStack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseStackConfiguration(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *StackConfiguration
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: "./testdata/appops/http-echo/dev",
			},
			want: &StackConfiguration{
				Name: TestStackA,
			},
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				path: "./testdata/appops/http-echo",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStackConfiguration(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStackConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseStackConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindAllStacks(t *testing.T) {
	_ = os.Chdir(TestProjectPathA)
	defer os.Chdir(TestCurrentDir)

	tests := []struct {
		name    string
		want    []*Stack
		wantErr bool
	}{
		{
			name: "given-project-path",
			want: []*Stack{
				{
					StackConfiguration: StackConfiguration{
						Name: TestStackA,
					},
					Path: filepath.Join(TestCurrentDir, TestStackPathAA),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindAllStacks()
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAllStacks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAllStacks() = %v, want %v", json.MustMarshal2PrettyString(got), json.MustMarshal2PrettyString(tt.want))
			}
		})
	}
}

func TestFindAllStacksFrom(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []*Stack
		wantErr bool
	}{
		{
			name: "given-project-path",
			args: args{
				path: "./testdata/appops/http-echo",
			},
			want: []*Stack{
				{
					StackConfiguration: StackConfiguration{
						Name: TestStackA,
					},
					Path: filepath.Join(TestCurrentDir, TestStackPathAA),
				},
			},
			wantErr: false,
		},
		{
			name: "give-project-path-with-two-stacks",
			args: args{
				path: "./testdata/appops/nginx-example",
			},
			want: []*Stack{
				{
					StackConfiguration: StackConfiguration{
						Name: TestStackA,
					},
					Path: filepath.Join(TestCurrentDir, TestStackPathBA),
				},
				{
					StackConfiguration: StackConfiguration{
						Name: TestStackB,
					},
					Path: filepath.Join(TestCurrentDir, TestStackPathBB),
				},
			},
			wantErr: false,
		},
		{
			name: "given-stack-path",
			args: args{
				path: "./testdata/appops/http-echo/dev/",
			},
			want: []*Stack{
				{
					StackConfiguration: StackConfiguration{
						Name: TestStackA,
					},
					Path: filepath.Join(TestCurrentDir, TestStackPathAA),
				},
			},
			wantErr: false,
		},
		{
			name: "given-no-stack-path",
			args: args{
				path: "./testdata/appops/http-echo/gray/",
			},
			want:    []*Stack{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindAllStacksFrom(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAllStacksFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAllStacksFrom() = %v, want %v", json.MustMarshal2PrettyString(got), json.MustMarshal2PrettyString(tt.want))
			}
		})
	}
}

func TestGetStack(t *testing.T) {
	_ = os.Chdir(TestStackPathAA)
	defer os.Chdir(TestCurrentDir)

	tests := []struct {
		name    string
		want    *Stack
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			want: &Stack{
				StackConfiguration: StackConfiguration{
					Name: TestStackA,
				},
				Path: filepath.Join(TestCurrentDir, TestStackPathAA),
			},
			wantErr: false,
			preRun:  func() {},
			postRun: func() {},
		},
		{
			name:    "fail-for-GetStackFrom",
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockGetStackFrom(ErrFake)
			},
			postRun: func() {},
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			tt.preRun()
			got, err := GetStack()
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStackFrom(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Stack
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: "./testdata/appops/http-echo/dev/",
			},
			want: &Stack{
				StackConfiguration: StackConfiguration{
					Name: TestStackA,
				},
				Path: filepath.Join(TestCurrentDir, TestStackPathAA),
			},
			wantErr: false,
		},
		{
			name: "failed-because-no-stack-path",
			args: args{
				path: TestProjectPathA,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "failed-because-not-existed-path",
			args: args{
				path: "./testdata-not-exist",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStackFrom(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStackFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStackFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStackFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is-stack-file",
			args: args{
				path: filepath.Join(TestStackPathAA, StackFile),
			},
			want: true,
		},
		{
			name: "is-not-stack-file",
			args: args{
				path: TestStackPathAA,
			},
			want: false,
		},
		{
			name: "is-not-stack-file-2",
			args: args{
				path: filepath.Join(TestStackPathAA, "main.k"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStackFile(tt.args.path); got != tt.want {
				t.Errorf("IsStackFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
