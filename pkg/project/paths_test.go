package project

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"

	"kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
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

func TestFindAllProjectsFrom(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []*v1.Project
		wantErr bool
	}{
		{
			name: "given-project-path",
			args: args{
				path: "./testdata/appops/http-echo",
			},
			want: []*v1.Project{
				{
					Name: TestProjectA,
					Path: filepath.Join(TestCurrentDir, TestProjectPathA),
					Stacks: []*v1.Stack{
						{
							Name: TestStackA,
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
			want: []*v1.Project{
				{
					Name: TestProjectB,
					Path: filepath.Join(TestCurrentDir, TestProjectPathB),
					Stacks: []*v1.Stack{
						{
							Name: TestStackA,
							Path: filepath.Join(TestCurrentDir, TestStackPathBA),
						},
						{
							Name: TestStackB,
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

func TestDetectProjectAndStack(t *testing.T) {
	FakeProject := &v1.Project{
		Name: TestProjectA,
		Path: filepath.Join(TestCurrentDir, TestProjectPathA),
		Stacks: []*v1.Stack{
			{
				Name: TestStackA,
				Path: filepath.Join(TestCurrentDir, TestStackPathAA),
			},
		},
	}
	FakeStack := &v1.Stack{
		Name: TestStackA,
		Path: filepath.Join(TestCurrentDir, TestStackPathAA),
	}

	type args struct {
		stackDir string
	}
	tests := []struct {
		name    string
		args    args
		project *v1.Project
		stack   *v1.Stack
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
			project, stack, err := DetectProjectAndStackFrom(tt.args.stackDir)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectProjectAndStackFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(project, tt.project) {
				t.Errorf("DetectProjectAndStackFrom() got = %v, want %v", project, tt.project)
			}
			if !reflect.DeepEqual(stack, tt.stack) {
				t.Errorf("DetectProjectAndStackFrom() gosuccess = %v, want %v", stack, tt.stack)
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
	mockey.Mock(GetStackFrom).To(func(_ string) (*v1.Stack, error) {
		if mockErr == nil {
			return &v1.Stack{}, nil
		}
		return nil, mockErr
	}).Build()
}

func mockGetProjectFrom(mockErr error) {
	mockey.Mock(getProjectFrom).To(func(_ string) (*v1.Project, error) {
		if mockErr == nil {
			return &v1.Project{}, nil
		}
		return nil, mockErr
	}).Build()
}

func mockFindProjectPathFrom(mockProjectDir string, mockErr error) {
	mockey.Mock(findProjectPathFrom).To(func(_ string) (string, error) {
		if mockErr == nil {
			return mockProjectDir, nil
		}
		return "", mockErr
	}).Build()
}
