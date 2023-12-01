package project

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/util/json"
)

// merge project tests and stack tests together to reuse test data
var (
	// Inject into the TestMain
	TestCurrentDir   string
	TestProjectPathA string
	TestStackPathAA  string
	TestProjectPathB string
	TestStackPathBA  string
	TestStackPathBB  string
	ErrFake          = errors.New("fake error")
)

const (
	TestProjectA string = "http-echo"
	TestProjectB string = "nginx-example"
	TestStackA   string = "dev"
	TestStackB   string = "prod"
)

func TestMain(m *testing.M) {
	TestCurrentDir, _ = os.Getwd()
	TestProjectPathA = filepath.Join("testdata", "appops", TestProjectA)
	TestStackPathAA = filepath.Join("testdata", "appops", TestProjectA, TestStackA)
	TestProjectPathB = filepath.Join("testdata", "appops", TestProjectB)
	TestStackPathBA = filepath.Join("testdata", "appops", TestProjectB, TestStackA)
	TestStackPathBB = filepath.Join("testdata", "appops", TestProjectB, TestStackB)

	os.Exit(m.Run())
}

func TestFindProjectPath(t *testing.T) {
	_ = os.Chdir(TestStackPathAA)
	defer os.Chdir(TestCurrentDir)

	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			want:    filepath.Join(TestCurrentDir, TestProjectPathA),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindProjectPath()
			if (err != nil) != tt.wantErr {
				t.Errorf("FindProjectPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindProjectPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindProjectPathFrom(t *testing.T) {
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
			want:    "testdata/appops/http-echo",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindProjectPathFrom(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindProjectPathFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindProjectPathFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsProject(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is-project",
			args: args{
				path: "./testdata/appops/http-echo/dev/ci-test",
			},
			want: false,
		},
		{
			name: "is-not-project",
			args: args{
				path: "./testdata/appops/http-echo",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsProject(tt.args.path); got != tt.want {
				t.Errorf("IsProject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseConfiguration(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Configuration
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: "./testdata/appops/http-echo/",
			},
			want: &Configuration{
				Name:   TestProjectA,
				Tenant: "",
			},
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				path: "./testdata/appops/http-echo/dev",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfiguration(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindAllProjects(t *testing.T) {
	_ = os.Chdir(TestProjectPathA)
	defer os.Chdir(TestCurrentDir)

	tests := []struct {
		name    string
		want    []*Project
		wantErr bool
	}{
		{
			name: "given-project-path",
			want: []*Project{
				{
					Configuration: Configuration{
						Name: TestProjectA,
					},
					Path: filepath.Join(TestCurrentDir, TestProjectPathA),
					Stacks: []*stack.Stack{
						{
							Configuration: stack.Configuration{
								Name: TestStackA,
							},
							Path: filepath.Join(TestCurrentDir, TestStackPathAA),
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindAllProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAllProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAllProjects() = %v, want %v", json.MustMarshal2PrettyString(got), json.MustMarshal2PrettyString(tt.want))
			}
		})
	}
}

func TestFindAllProjectsFrom(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []*Project
		wantErr bool
	}{
		{
			name: "given-project-path",
			args: args{
				path: "./testdata/appops/http-echo",
			},
			want: []*Project{
				{
					Configuration: Configuration{
						Name: TestProjectA,
					},
					Path: filepath.Join(TestCurrentDir, TestProjectPathA),
					Stacks: []*stack.Stack{
						{
							Configuration: stack.Configuration{
								Name: TestStackA,
							},
							Path: filepath.Join(TestCurrentDir, TestStackPathAA),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "give-project-path-with-two-stacks",
			args: args{
				path: "./testdata/appops/nginx-example",
			},
			want: []*Project{
				{
					Configuration: Configuration{
						Name: TestProjectB,
					},
					Path: filepath.Join(TestCurrentDir, TestProjectPathB),
					Stacks: []*stack.Stack{
						{
							Configuration: stack.Configuration{
								Name: TestStackA,
							},
							Path: filepath.Join(TestCurrentDir, TestStackPathBA),
						},
						{
							Configuration: stack.Configuration{
								Name: TestStackB,
							},
							Path: filepath.Join(TestCurrentDir, TestStackPathBB),
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindAllProjectsFrom(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAllProjectsFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAllProjectsFrom() = %v, want %v", json.MustMarshal2PrettyString(got), json.MustMarshal2PrettyString(tt.want))
			}
		})
	}
}

func TestGetProject(t *testing.T) {
	_ = os.Chdir(TestProjectPathA)
	defer os.Chdir(TestCurrentDir)

	tests := []struct {
		name    string
		want    *Project
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			want: &Project{
				Configuration: Configuration{
					Name: TestProjectA,
				},
				Path: filepath.Join(TestCurrentDir, TestProjectPathA),
				Stacks: []*stack.Stack{
					{
						Configuration: stack.Configuration{
							Name: TestStackA,
						},
						Path: filepath.Join(TestCurrentDir, TestStackPathAA),
					},
				},
			},
			wantErr: false,
			preRun:  func() {},
			postRun: func() {},
		},
		{
			name:    "fail-for-GetProjectFrom",
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockGetProjectFrom(ErrFake)
			},
			postRun: func() {},
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			tt.preRun()
			got, err := GetProject()
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProjectFrom(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Project
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: filepath.Join(TestCurrentDir, TestProjectPathA),
			},
			want: &Project{
				Configuration: Configuration{
					Name: TestProjectA,
				},
				Path: filepath.Join(TestCurrentDir, TestProjectPathA),
				Stacks: []*stack.Stack{
					{
						Configuration: stack.Configuration{
							Name: TestStackA,
						},
						Path: filepath.Join(TestCurrentDir, TestStackPathAA),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "failed-because-no-project-path",
			args: args{
				path: filepath.Join(TestCurrentDir, TestStackPathAA),
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
			got, err := GetProjectFrom(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProjectFrom() = %v, want %v", json.MustMarshal2PrettyString(got), json.MustMarshal2PrettyString(tt.want))
			}
		})
	}
}

func TestIsProjectFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is-project-file",
			args: args{
				path: filepath.Join(TestProjectPathA, ProjectFile),
			},
			want: true,
		},
		{
			name: "is-not-project-file",
			args: args{
				path: TestProjectPathA,
			},
			want: false,
		},
		{
			name: "is-not-project-file-2",
			args: args{
				path: filepath.Join(TestStackPathAA, stack.File),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsProjectFile(tt.args.path); got != tt.want {
				t.Errorf("IsProjectFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectProjectAndStack(t *testing.T) {
	FakeProject := &Project{
		Configuration: Configuration{
			Name: TestProjectA,
		},
		Path: filepath.Join(TestCurrentDir, TestProjectPathA),
		Stacks: []*stack.Stack{
			{
				Configuration: stack.Configuration{
					Name: TestStackA,
				},
				Path: filepath.Join(TestCurrentDir, TestStackPathAA),
			},
		},
	}
	FakeStack := &stack.Stack{
		Configuration: stack.Configuration{
			Name: TestStackA,
		},
		Path: filepath.Join(TestCurrentDir, TestStackPathAA),
	}

	type args struct {
		stackDir string
	}
	tests := []struct {
		name    string
		args    args
		project *Project
		stack   *stack.Stack
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			args: args{
				stackDir: "./testdata/appops/http-echo/dev/",
			},
			project: FakeProject,
			stack:   FakeStack,
			wantErr: false,
			preRun:  func() {},
			postRun: func() {},
		},
		{
			name: "fail-for-abs",
			args: args{
				stackDir: "./testdata/appops/http-echo/dev/",
			},
			project: nil,
			stack:   nil,
			wantErr: true,
			preRun: func() {
				mockAbs("", ErrFake)
			},
			postRun: func() {},
		},
		{
			name: "fail-for-GetStackFrom",
			args: args{
				stackDir: "./testdata/appops/http-echo/dev/",
			},
			project: nil,
			stack:   nil,
			wantErr: true,
			preRun: func() {
				mockGetStackFrom(ErrFake)
			},
			postRun: func() {},
		},
		{
			name: "fail-for-FindProjectPathFrom",
			args: args{
				stackDir: "./testdata/appops/http-echo/dev/",
			},
			project: nil,
			stack:   nil,
			wantErr: true,
			preRun: func() {
				mockFindProjectPathFrom("", ErrFake)
			},
			postRun: func() {},
		},
		{
			name: "fail-for-GetProjectFrom",
			args: args{
				stackDir: "./testdata/appops/http-echo/dev/",
			},
			project: nil,
			stack:   nil,
			wantErr: true,
			preRun: func() {
				mockGetProjectFrom(ErrFake)
			},
			postRun: func() {},
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			tt.preRun()
			got, gosuccess, err := DetectProjectAndStack(tt.args.stackDir)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectProjectAndStack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.project) {
				t.Errorf("DetectProjectAndStack() got = %v, want %v", got, tt.project)
			}
			if !reflect.DeepEqual(gosuccess, tt.stack) {
				t.Errorf("DetectProjectAndStack() gosuccess = %v, want %v", gosuccess, tt.stack)
			}
		})
	}
}

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
			got, err := stack.FindStackPath()
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
			got, err := stack.FindStackPathFrom(tt.args.path)
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
			if got := stack.IsStack(tt.args.path); got != tt.want {
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
		want    *stack.Configuration
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: "./testdata/appops/http-echo/dev",
			},
			want: &stack.Configuration{
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
			got, err := stack.ParseStackConfiguration(tt.args.path)
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
		want    []*stack.Stack
		wantErr bool
	}{
		{
			name: "given-project-path",
			want: []*stack.Stack{
				{
					Configuration: stack.Configuration{
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
			got, err := stack.FindAllStacks()
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
		want    []*stack.Stack
		wantErr bool
	}{
		{
			name: "given-project-path",
			args: args{
				path: "./testdata/appops/http-echo",
			},
			want: []*stack.Stack{
				{
					Configuration: stack.Configuration{
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
			want: []*stack.Stack{
				{
					Configuration: stack.Configuration{
						Name: TestStackA,
					},
					Path: filepath.Join(TestCurrentDir, TestStackPathBA),
				},
				{
					Configuration: stack.Configuration{
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
			want: []*stack.Stack{
				{
					Configuration: stack.Configuration{
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
			want:    []*stack.Stack{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stack.FindAllStacksFrom(tt.args.path)
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
		want    *stack.Stack
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			want: &stack.Stack{
				Configuration: stack.Configuration{
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
			got, err := stack.GetStack()
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
		want    *stack.Stack
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: "./testdata/appops/http-echo/dev/",
			},
			want: &stack.Stack{
				Configuration: stack.Configuration{
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
			got, err := stack.GetStackFrom(tt.args.path)
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
				path: filepath.Join(TestStackPathAA, stack.File),
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
			if got := stack.IsStackFile(tt.args.path); got != tt.want {
				t.Errorf("IsStackFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewStack(t *testing.T) {
	type args struct {
		config *stack.Configuration
		path   string
	}
	tests := []struct {
		name string
		args args
		want *stack.Stack
	}{
		{
			name: "success",
			args: args{
				config: &stack.Configuration{
					Name: TestStackA,
				},
				path: TestStackPathAA,
			},
			want: &stack.Stack{
				Configuration: stack.Configuration{
					Name: TestStackA,
				},
				Path: TestStackPathAA,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stack.NewStack(tt.args.config, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStack_GetName(t *testing.T) {
	type fields struct {
		Configuration stack.Configuration
		Path          string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				Configuration: stack.Configuration{
					Name: TestStackA,
				},
				Path: TestStackPathAA,
			},
			want: TestStackA,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stack.Stack{
				Configuration: tt.fields.Configuration,
				Path:          tt.fields.Path,
			}
			if got := s.GetName(); got != tt.want {
				t.Errorf("Stack.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStack_GetPath(t *testing.T) {
	type fields struct {
		Configuration stack.Configuration
		Path          string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				Configuration: stack.Configuration{
					Name: TestStackA,
				},
				Path: TestStackPathAA,
			},
			want: TestStackPathAA,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stack.Stack{
				Configuration: tt.fields.Configuration,
				Path:          tt.fields.Path,
			}
			if got := s.GetPath(); got != tt.want {
				t.Errorf("Stack.GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStack_TableReport(t *testing.T) {
	type fields struct {
		Configuration stack.Configuration
		Path          string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				Configuration: stack.Configuration{
					Name: TestStackA,
				},
				Path: TestStackPathAA,
			},
			want: `┌────────────────────────────────────────────┐
| Type       | Name                          |
| Stack Name | dev                           |
| Stack Path | testdata/appops/http-echo/dev |
└────────────────────────────────────────────┘`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stack.Stack{
				Configuration: tt.fields.Configuration,
				Path:          tt.fields.Path,
			}
			got := pterm.RemoveColorFromString(s.TableReport())
			if got != tt.want {
				t.Errorf("Stack.TableReport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockAbs(mockAbs string, mockErr error) {
	mockey.Mock(filepath.Abs).To(func(_ string) (string, error) {
		return mockAbs, mockErr
	}).Build()
}

func mockGetStackFrom(mockErr error) {
	mockey.Mock(stack.GetStackFrom).To(func(_ string) (*stack.Stack, error) {
		if mockErr == nil {
			return &stack.Stack{}, nil
		}
		return nil, mockErr
	}).Build()
}

func mockGetProjectFrom(mockErr error) {
	mockey.Mock(GetProjectFrom).To(func(_ string) (*Project, error) {
		if mockErr == nil {
			return &Project{}, nil
		}
		return nil, mockErr
	}).Build()
}

func mockFindProjectPathFrom(mockProjectDir string, mockErr error) {
	mockey.Mock(FindProjectPathFrom).To(func(_ string) (string, error) {
		if mockErr == nil {
			return mockProjectDir, nil
		}
		return "", mockErr
	}).Build()
}
