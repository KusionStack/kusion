package diff

import (
	"testing"

	"github.com/gonvenience/ytbx"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/third_party/dyff"
)

func TestToReportString(t *testing.T) {
	t.Run("human report", func(t *testing.T) {
		actual, err := ToReportString(NewHumanReport(&dyff.Report{
			From:  ytbx.InputFile{},
			To:    ytbx.InputFile{},
			Diffs: []dyff.Diff{},
		}), OutputHuman)
		assert.Nil(t, err)
		assert.Equal(t, "\n", actual)
	})

	t.Run("raw report", func(t *testing.T) {
		actual, err := ToReportString(NewHumanReport(&dyff.Report{
			From:  ytbx.InputFile{},
			To:    ytbx.InputFile{},
			Diffs: []dyff.Diff{},
		}), OutputRaw)
		assert.Nil(t, err)
		assert.Equal(t, "diffs: []\n", actual)
	})
}

func TestToHumanString(t *testing.T) {
	t.Run("no diff", func(t *testing.T) {
		report, err := ToReport("", "")
		assert.Nil(t, err)

		humanReport, err := ToHumanString(NewHumanReport(report))
		assert.Nil(t, err)
		assert.Equal(t, "\n", humanReport)
	})

	t.Run("one diff", func(t *testing.T) {
		report, err := ToReport(map[string]interface{}{"a": "foo"}, map[string]interface{}{"a": "Foo"})
		assert.Nil(t, err)

		humanReport, err := ToHumanString(NewHumanReport(report))
		assert.Nil(t, err)
		assert.Equal(t, `
a
  ± value change
    - foo
    + Foo

`, humanReport)
	})
}

func TestToRawString(t *testing.T) {
	t.Run("no diff", func(t *testing.T) {
		report, err := ToReport("", "")
		assert.Nil(t, err)

		humanReport, err := ToRawString(NewHumanReport(report))
		assert.Nil(t, err)
		assert.Equal(t, "diffs: []\n", humanReport)
	})

	t.Run("one diff", func(t *testing.T) {
		report, err := ToReport(map[string]interface{}{"a": "foo"}, map[string]interface{}{"a": "Foo"})
		assert.Nil(t, err)

		humanReport, err := ToRawString(NewHumanReport(report))
		assert.Nil(t, err)
		assert.Equal(t, `diffs:
    - path:
        documentidx: 0
        pathelements:
            - idx: -1
              key: ""
              name: a
      details:
        - kind: 177
          from: foo
          to: Foo
`, humanReport)
	})
}

func TestToReport(t *testing.T) {
	t.Run("compare string, 1 diff", func(t *testing.T) {
		report, err := ToReport("a: foo", "a: Foo")
		assert.Nil(t, err)
		assert.Equal(t, len(report.Diffs), 1)
	})

	t.Run("compare struct type, 2 diff", func(t *testing.T) {
		report, err := ToReport(
			map[string]interface{}{
				"a": "foo",
				"b": 1,
			},
			map[string]interface{}{
				"a": "Foo",
				"b": 2,
			})
		assert.Nil(t, err)
		assert.Equal(t, len(report.Diffs), 2)
	})
}
