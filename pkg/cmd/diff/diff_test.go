//go:build !arm64
// +build !arm64

package diff

import (
	"errors"
	"os"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/gonvenience/ytbx"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	diffutil "kusionstack.io/kusion/pkg/util/diff"
	"kusionstack.io/kusion/third_party/diff"
)

func TestNewCmdDiff(t *testing.T) {
	mockey.PatchConvey("diff by files", t, func() {
		cmd := NewCmdDiff()
		cmd.SetArgs([]string{"testdata/pod1.yaml", "testdata/pod2.yaml"})
		err := cmd.Execute()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("diff by stdin with docs size < 2", t, func() {
		mockey.Mock(ytbx.IsStdin).To(func(location string) bool {
			return true
		}).Build()

		cmd := NewCmdDiff()
		cmd.SetArgs([]string{"testdata/namespace.yaml"})
		err := cmd.Execute()
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("diff by stdin with error diff-mode", t, func() {
		mockey.Mock(ytbx.IsStdin).To(func(location string) bool {
			return true
		}).Build()

		cmd := NewCmdDiff()
		assert.Nil(t, cmd.Flags().Set("diff-mode", "xxx"))
		cmd.SetArgs([]string{"testdata/pod-full.yaml"})
		err := cmd.Execute()
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("diff by stdin with error output", t, func() {
		mockey.Mock(ytbx.IsStdin).To(func(location string) bool {
			return true
		}).Build()

		cmd := NewCmdDiff()
		assert.Nil(t, cmd.Flags().Set("output", "xxx"))
		cmd.SetArgs([]string{"testdata/pod-full.yaml"})
		err := cmd.Execute()
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("diff by stdin with diff-mode", t, func() {
		mockey.Mock(ytbx.IsStdin).To(func(location string) bool {
			return true
		}).Build()

		err := mockStdin("testdata/pod-full.yaml")
		if err != nil {
			panic(err)
		}
		defer func() { os.Stdin = originalOSStdin }()

		cmd := NewCmdDiff()
		assert.Nil(t, cmd.Flags().Set("diff-mode", ModeLive))
		cmd.SetArgs([]string{"testdata/pod-full.yaml"})
		err = cmd.Execute()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("diff by files with flags", t, func() {
		cmd := NewCmdDiff()
		assert.Nil(t, cmd.Flags().Set("diff-mode", ModeIgnoreAdded))
		assert.Nil(t, cmd.Flags().Set("output", diffutil.OutputRaw))
		assert.Nil(t, cmd.Flags().Set("sort-by-kubernetes-resource", "true"))
		assert.Nil(t, cmd.Flags().Set("swap", "true"))
		cmd.SetArgs([]string{"testdata/pod1.yaml", "testdata/pod2.yaml"})
		err := cmd.Execute()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("diff by os.stdin with flags", t, func() {
		mockey.Mock(ytbx.IsStdin).To(func(location string) bool {
			return true
		}).Build()

		err := mockStdin("testdata/pod-full.yaml")
		if err != nil {
			panic(err)
		}
		defer func() { os.Stdin = originalOSStdin }()

		cmd := NewCmdDiff()
		assert.Nil(t, cmd.Flags().Set("diff-mode", ModeIgnoreAdded))
		assert.Nil(t, cmd.Flags().Set("output", diffutil.OutputRaw))
		assert.Nil(t, cmd.Flags().Set("sort-by-kubernetes-resource", "true"))
		assert.Nil(t, cmd.Flags().Set("swap", "true"))
		cmd.SetArgs([]string{"testdata/pod-full.yaml"})
		err = cmd.Execute()
		assert.Nil(t, err)
	})
}

func Test_liveDiffWithStdin(t *testing.T) {
	mockey.PatchConvey("create normalizer failed", t, func() {
		mockey.Mock(diff.NewDefaultIgnoreNormalizer).To(func(paths []string) (diff.Normalizer, error) {
			return nil, errors.New("mock create normalizer error")
		}).Build()

		err := mockStdin("testdata/pod-full.yaml")
		if err != nil {
			panic(err)
		}
		defer func() { os.Stdin = originalOSStdin }()

		err = liveDiffWithStdin()
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("calculate diff failed", t, func() {
		mockey.Mock(
			diff.Diff).To(
			func(config, live *unstructured.Unstructured, opts ...diff.Option) (*diff.DiffResult, error) {
				return nil, errors.New("mock calculate diff error")
			},
		).Build()

		err := mockStdin("testdata/pod-full.yaml")
		if err != nil {
			panic(err)
		}
		defer func() { os.Stdin = originalOSStdin }()

		err = liveDiffWithStdin()
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("liveDiffWithStdin success", t, func() {
		err := mockStdin("testdata/pod-full.yaml")
		if err != nil {
			panic(err)
		}
		defer func() { os.Stdin = originalOSStdin }()

		err = liveDiffWithStdin()
		assert.Nil(t, err)
	})
}

var originalOSStdin = os.Stdin

func mockStdin(file string) error {
	open, err := os.Open(file)
	if err != nil {
		return err
	}

	os.Stdin = open

	return nil
}

func Test_liveDiffWithFile(t *testing.T) {
	mockey.PatchConvey("read fromLocation error", t, func() {
		err := liveDiffWithFile("no/such/file", "no/such/file")
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("read toLocation error", t, func() {
		err := liveDiffWithFile("testdata/pod1.yaml", "no/such/file")
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("create normalizer failed", t, func() {
		mockey.Mock(diff.NewDefaultIgnoreNormalizer).To(func(paths []string) (diff.Normalizer, error) {
			return nil, errors.New("mock create normalizer error")
		}).Build()

		err := liveDiffWithFile("testdata/pod1.yaml", "testdata/pod1.yaml")
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("calculate diff failed", t, func() {
		mockey.Mock(
			diff.Diff).To(
			func(config, live *unstructured.Unstructured, opts ...diff.Option) (*diff.DiffResult, error) {
				return nil, errors.New("mock calculate diff error")
			},
		).Build()

		err := liveDiffWithFile("testdata/pod1.yaml", "testdata/pod1.yaml")
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("liveDiffWithFile success", t, func() {
		err := liveDiffWithFile("testdata/pod1.yaml", "testdata/pod1.yaml")
		assert.Nil(t, err)
	})
}

func Test_loadFile(t *testing.T) {
	mockey.PatchConvey("load failed", t, func() {
		_, err := loadFile("no/such/file")
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("load success", t, func() {
		_, err := loadFile("testdata/pod1.yaml")
		assert.Nil(t, err)
	})
}
