package diff

import (
	"reflect"
	"testing"

	"github.com/gonvenience/ytbx"
	"github.com/stretchr/testify/assert"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
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

func TestMaskSensitiveData(t *testing.T) {
	var nilResource *v1.Resource

	testcases := []struct {
		name                  string
		oldData               interface{}
		newData               interface{}
		expectedMaskedOldData interface{}
		expectedMaskedNewData interface{}
	}{
		{
			name: "same old and new data",
			oldData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"data": map[string]interface{}{
						"key": "dmFsdWUK",
					},
				},
			},
			newData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"data": map[string]interface{}{
						"key": "dmFsdWUK",
					},
				},
			},
			expectedMaskedOldData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"data": map[string]interface{}{
						"key": "*******",
					},
				},
			},
			expectedMaskedNewData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"data": map[string]interface{}{
						"key": "*******",
					},
				},
			},
		},
		{
			name: "different old and new data",
			oldData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"data": map[string]interface{}{
						"key": "dmFsdWUK",
					},
				},
			},
			newData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"data": map[string]interface{}{
						"key": "dmFsdWUtY2hhbmdlZAo=",
					},
				},
			},
			expectedMaskedOldData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"data": map[string]interface{}{
						"key": "***before***",
					},
				},
			},
			expectedMaskedNewData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"data": map[string]interface{}{
						"key": "***after****",
					},
				},
			},
		},
		{
			name:    "empty old data",
			oldData: nilResource,
			newData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"data": map[string]interface{}{
						"key": "dmFsdWUtY2hhbmdlZAo=",
					},
				},
			},
			expectedMaskedOldData: nilResource,
			expectedMaskedNewData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"data": map[string]interface{}{
						"key": "*******",
					},
				},
			},
		},
		{
			name: "not secret resource",
			oldData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Namespace",
				},
			},
			newData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Namespace",
				},
			},
			expectedMaskedOldData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Namespace",
				},
			},
			expectedMaskedNewData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Namespace",
				},
			},
		},
		{
			name: "secrets with string data",
			oldData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"stringData": map[string]interface{}{
						"key": "dmFsdWUK",
					},
				},
			},
			newData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"stringData": map[string]interface{}{
						"key": "dmFsdWUtY2hhbmdlZAo=",
					},
				},
			},
			expectedMaskedOldData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"stringData": map[string]interface{}{
						"key": "***before***",
					},
				},
			},
			expectedMaskedNewData: &v1.Resource{
				Type: v1.Kubernetes,
				Attributes: map[string]interface{}{
					"kind": "Secret",
					"stringData": map[string]interface{}{
						"key": "***after****",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualMaskedOldData, actualMaskedNewData := MaskSensitiveData(
				tc.oldData, tc.newData,
			)

			if !reflect.DeepEqual(actualMaskedOldData, tc.expectedMaskedOldData) {
				t.Errorf("masked old data does not match expected result, got: %v, wanted: %v",
					actualMaskedOldData, tc.expectedMaskedOldData)
			}

			if !reflect.DeepEqual(actualMaskedNewData, tc.expectedMaskedNewData) {
				t.Errorf("masked new data does not match expected result, get: %v, wanted: %v",
					actualMaskedNewData, tc.expectedMaskedNewData)
			}
		})
	}
}
