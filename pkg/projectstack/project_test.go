//go:build !arm64
// +build !arm64

package projectstack

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"kusionstack.io/kusion/pkg/util/json"
)

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

func TestParseProjectConfiguration(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *ProjectConfiguration
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: "./testdata/appops/http-echo/",
			},
			want: &ProjectConfiguration{
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
			got, err := ParseProjectConfiguration(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProjectConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseProjectConfiguration() = %v, want %v", got, tt.want)
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
					ProjectConfiguration: ProjectConfiguration{
						Name: TestProjectA,
					},
					Path: filepath.Join(TestCurrentDir, TestProjectPathA),
					Stacks: []*Stack{
						{
							StackConfiguration: StackConfiguration{
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
					ProjectConfiguration: ProjectConfiguration{
						Name: TestProjectA,
					},
					Path: filepath.Join(TestCurrentDir, TestProjectPathA),
					Stacks: []*Stack{
						{
							StackConfiguration: StackConfiguration{
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
					ProjectConfiguration: ProjectConfiguration{
						Name: TestProjectB,
					},
					Path: filepath.Join(TestCurrentDir, TestProjectPathB),
					Stacks: []*Stack{
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
				ProjectConfiguration: ProjectConfiguration{
					Name: TestProjectA,
				},
				Path: filepath.Join(TestCurrentDir, TestProjectPathA),
				Stacks: []*Stack{
					{
						StackConfiguration: StackConfiguration{
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
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				ProjectConfiguration: ProjectConfiguration{
					Name: TestProjectA,
				},
				Path: filepath.Join(TestCurrentDir, TestProjectPathA),
				Stacks: []*Stack{
					{
						StackConfiguration: StackConfiguration{
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
				path: filepath.Join(TestStackPathAA, StackFile),
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
		ProjectConfiguration: ProjectConfiguration{
			Name: TestProjectA,
		},
		Path: filepath.Join(TestCurrentDir, TestProjectPathA),
		Stacks: []*Stack{
			{
				StackConfiguration: StackConfiguration{
					Name: TestStackA,
				},
				Path: filepath.Join(TestCurrentDir, TestStackPathAA),
			},
		},
	}
	FakeStack := &Stack{
		StackConfiguration: StackConfiguration{
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
		stack   *Stack
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
			postRun: func() {
				defer monkey.UnpatchAll()
			},
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
			postRun: func() {
				defer monkey.UnpatchAll()
			},
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
			postRun: func() {
				defer monkey.UnpatchAll()
			},
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
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func mockAbs(mockAbs string, mockErr error) {
	monkey.Patch(filepath.Abs, func(_ string) (string, error) {
		return mockAbs, mockErr
	})
}

func mockGetStackFrom(mockErr error) {
	monkey.Patch(GetStackFrom, func(_ string) (*Stack, error) {
		if mockErr == nil {
			return &Stack{}, nil
		}
		return nil, mockErr
	})
}

func mockGetProjectFrom(mockErr error) {
	monkey.Patch(GetProjectFrom, func(_ string) (*Project, error) {
		if mockErr == nil {
			return &Project{}, nil
		}
		return nil, mockErr
	})
}

func mockFindProjectPathFrom(mockProjectDir string, mockErr error) {
	monkey.Patch(FindProjectPathFrom, func(_ string) (string, error) {
		if mockErr == nil {
			return mockProjectDir, nil
		}
		return "", mockErr
	})
}
