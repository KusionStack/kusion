package deps

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/projectstack"
)

var workDir string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	workDir = filepath.Join(cwd, "testdata")
}

func TestNewCmdDeps(t *testing.T) {
	cmd := NewCmdDeps()
	if err := cmd.Execute(); err == nil {
		t.Fatal(err)
	}
}

func TestDepsOptions_Validate(t *testing.T) {
	tCases := []struct {
		name    string
		only    string
		direct  string
		workDir string
		focus   []string
		ignore  []string
		errMsg  string
	}{

		{
			name:   "invalid output filter",
			errMsg: "invalid output downstream type. supported types: project, stack",
		},
		{
			name:   "invalid direct",
			only:   "project",
			errMsg: "invalid output direction of the dependency inspection. supported directions: up, down",
		},
		{
			name:   "invalid workdir",
			only:   "project",
			direct: "up",
			errMsg: "invalid work dir: stat : no such file or directory",
		},
		{
			name:    "invalid focus",
			only:    "project",
			direct:  "up",
			workDir: workDir,
			errMsg:  "invalid focus paths. cannot be empty",
		},
		{
			name:    "invalid ignore",
			only:    "project",
			direct:  "up",
			workDir: workDir,
			focus:   []string{"file.k"},
			ignore:  []string{"invalid_path.k"},
			errMsg:  fmt.Sprintf("invalid ignore path. need to be valid relative path from the workdir: stat %s: no such file or directory", filepath.Join(workDir, "invalid_path.k")),
		},
		{
			name:    "valid",
			only:    "project",
			direct:  "up",
			workDir: workDir,
			focus:   []string{"file.k"},
			errMsg:  "",
		},
	}
	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := DepsOptions{
				workDir: tc.workDir,
				Direct:  tc.direct,
				Only:    tc.only,
				Focus:   tc.focus,
				Ignore:  tc.ignore,
			}
			err := opt.Validate()
			if err != nil && err.Error() != tc.errMsg {
				t.Fatalf("wrong validate errMsg, actual: %s, expect: %s", err.Error(), tc.errMsg)
			}
			if err == nil && tc.errMsg != "" {
				t.Fatalf("wrong validate result: actual: validate success, expect err: %s", tc.errMsg)
			}
		})
	}
}

func TestDepsOptions_Complete(t *testing.T) {
	tCases := []struct {
		name      string
		args      []string
		completed string
	}{
		{
			name:      "omit workdir in args",
			completed: filepath.Dir(workDir),
		},
		{
			name:      "specify workdir in args",
			args:      []string{"workdir path"},
			completed: "workdir path",
		},
	}
	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			opt := NewDepsOptions()
			opt.Complete(tc.args)
			if opt.workDir != tc.completed {
				t.Fatalf("wrong completed workdir, actual: %s, expect: %s", opt.workDir, tc.completed)
			}
		})
	}

}

func TestDepsOptions_Run(t *testing.T) {
	tCases := []struct {
		workDir string
		direct  string
		only    string
	}{
		{
			workDir: workDir,
			direct:  "up",
		},
		{
			workDir: workDir,
			direct:  "down",
			only:    "project",
		},
	}
	for _, tc := range tCases {
		t.Run(tc.direct, func(t *testing.T) {
			opt := DepsOptions{
				workDir: tc.workDir,
				Direct:  tc.direct,
				Only:    tc.only,
			}
			if err := opt.Run(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDepsOptions_Run2(t *testing.T) {
	opt := DepsOptions{
		workDir: workDir,
		Direct:  "up",
		Focus: []string{
			"appops/projectC/dev/main.k",
		},
	}
	err := opt.Run()
	if err != nil {
		t.Fatal(err)
	}
}

var downstreamTestCases = []struct {
	name               string
	focus              []string
	downStreamProjects []string
	ignore             []string
}{
	{
		name: "change base and entrance files",
		focus: []string{
			"base/frontend/container/container_port.k",
			"appops/projectA/base/base.k",
			"appops/projectA/dev/main.k",
			"appops/projectA/dev/datafile.sql",
		},
		downStreamProjects: []string{
			"appops/projectA",
			"appops/projectB",
			"appops/projectC",
		},
	},
	{
		name:  "change common base: container port",
		focus: []string{"base/frontend/container/container_port.k"},
		downStreamProjects: []string{
			"appops/projectA",
			"appops/projectB",
			"appops/projectC",
		},
	},
	{
		name:  "change server render base",
		focus: []string{"base/render/server/server_render.k"},
		downStreamProjects: []string{
			"appops/projectA",
			"appops/projectC",
		},
	},
	{
		name:  "change job render base",
		focus: []string{"base/render/job/job_render.k"},
		downStreamProjects: []string{
			"appops/projectB",
		},
	},
	{
		name: "ignore job render base",
		focus: []string{
			"base/render/server/server_render.k",
			"base/render/job/job_render.k",
		},
		ignore: []string{
			"base/render/job/job_render.k",
		},
		downStreamProjects: []string{
			"appops/projectA",
			"appops/projectC",
		},
	},
	{
		name: "only change entrance files",
		focus: []string{
			"appops/projectA/base/base.k",
			"appops/projectA/dev/main.k",
			"appops/projectB/dev/main.k",
		},
		downStreamProjects: []string{
			"appops/projectA",
			"appops/projectB",
		},
	},
	{
		name: "only change data files",
		focus: []string{
			"appops/projectA/dev/datafile.sql",
		},
		downStreamProjects: []string{
			"appops/projectA",
		},
	},
	{
		name: "delete files",
		focus: []string{
			"appops/projectC/dev/non_exist.sql",
			"base/frontend/non_exist/non_exist.k",
		},
		downStreamProjects: []string{
			"appops/projectC",
		},
	},
}

func TestFindDownStreams(t *testing.T) {
	projects, err := projectstack.FindAllProjectsFrom(workDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range downstreamTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := findDownStreams(workDir, projects, toSet(tc.focus), toSet(tc.ignore), true)
			if err != nil {
				t.Fatal(err)
			}
			assert.ElementsMatch(t, result.toSlice(), tc.downStreamProjects, "test result mismatch")
		})
	}
}

func BenchmarkDownStream(b *testing.B) {
	tc := downstreamTestCases[0]
	for i := 0; i < b.N; i++ {
		projects, err := projectstack.FindAllProjectsFrom(workDir)
		if err != nil {
			b.Fatal(err)
		}
		result, err := findDownStreams(workDir, projects, toSet(tc.focus), toSet(tc.ignore), true)
		if err != nil {
			b.Fatal(err)
		}
		assert.ElementsMatch(b, result.toSlice(), tc.downStreamProjects, "test result mismatch")
	}
}
