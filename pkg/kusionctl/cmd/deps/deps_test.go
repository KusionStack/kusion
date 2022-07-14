package deps

import (
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
