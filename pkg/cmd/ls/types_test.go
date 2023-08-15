//go:build !arm64
// +build !arm64

package ls

import (
	"github.com/bytedance/mockey"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/projectstack"
)

var (
	workDir     = project.Path
	report      = NewLsReport(workDir, []*projectstack.Project{project})
	humanReport = `┌───────────────────────────────────────────────────────────────┐
| Type       | Name                                             |
| Stack Name | dev                                              |
| Stack Path | ../../projectstack/testdata/appops/http-echo/dev |
└───────────────────────────────────────────────────────────────┘`
	treeReport = `└─┬http-echo
  └──dev
`
)

func Test_lsReport_Human(t *testing.T) {
	mockey.PatchConvey("test le report human", t, func() {
		mockPromptOutput()

		got, err := report.Human()
		assert.Nil(t, err)
		assert.Equal(t, humanReport, pterm.RemoveColorFromString(got))
	})
}

func Test_lsReport_Tree(t *testing.T) {
	got, err := report.Tree()
	assert.Nil(t, err)
	assert.Equal(t, treeReport, pterm.RemoveColorFromString(got))
}
