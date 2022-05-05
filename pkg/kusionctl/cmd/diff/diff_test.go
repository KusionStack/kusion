package diff

import (
	"errors"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/gonvenience/ytbx"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"kusionstack.io/kusion/third_party/diff"
)

func TestNewCmdDiff(t *testing.T) {
	t.Run("diff by files", func(t *testing.T) {
		cmd := NewCmdDiff()
		cmd.SetArgs([]string{"testdata/pod1.yaml", "testdata/pod2.yaml"})
		err := cmd.Execute()
		assert.Nil(t, err)
	})

	t.Run("diff by stdin with docs size < 2", func(t *testing.T) {
		monkey.Patch(ytbx.IsStdin, func(location string) bool {
			return true
		})
		defer monkey.UnpatchAll()

		cmd := NewCmdDiff()
		cmd.SetArgs([]string{"testdata/namespace.yaml"})
		err := cmd.Execute()
		assert.NotNil(t, err)
	})

	t.Run("diff by stdin with error diff-mode", func(t *testing.T) {
		monkey.Patch(ytbx.IsStdin, func(location string) bool {
			return true
		})
		defer monkey.UnpatchAll()

		cmd := NewCmdDiff()
		assert.Nil(t, cmd.Flags().Set("diff-mode", "xxx"))
		cmd.SetArgs([]string{"testdata/pod-full.yaml"})
		err := cmd.Execute()
		assert.NotNil(t, err)
	})

	t.Run("diff by stdin with error output", func(t *testing.T) {
		monkey.Patch(ytbx.IsStdin, func(location string) bool {
			return true
		})
		defer monkey.UnpatchAll()

		cmd := NewCmdDiff()
		assert.Nil(t, cmd.Flags().Set("output", "xxx"))
		cmd.SetArgs([]string{"testdata/pod-full.yaml"})
		err := cmd.Execute()
		assert.NotNil(t, err)
	})

	t.Run("diff by stdin with diff-mode", func(t *testing.T) {
		monkey.Patch(ytbx.IsStdin, func(location string) bool {
			return true
		})
		defer monkey.UnpatchAll()

		err := mockStdin("testdata/pod-full.yaml")
		if err != nil {
			panic(err)
		}
		defer func() { os.Stdin = originalOSStdin }()

		cmd := NewCmdDiff()
		assert.Nil(t, cmd.Flags().Set("diff-mode", DiffModeLive))
		cmd.SetArgs([]string{"testdata/pod-full.yaml"})
		err = cmd.Execute()
		assert.Nil(t, err)
	})

	t.Run("diff by files with flags", func(t *testing.T) {
		cmd := NewCmdDiff()
		assert.Nil(t, cmd.Flags().Set("diff-mode", DiffModeIgnoreAdded))
		assert.Nil(t, cmd.Flags().Set("output", OutputRaw))
		assert.Nil(t, cmd.Flags().Set("sort-by-kubernetes-resource", "true"))
		assert.Nil(t, cmd.Flags().Set("swap", "true"))
		cmd.SetArgs([]string{"testdata/pod1.yaml", "testdata/pod2.yaml"})
		err := cmd.Execute()
		assert.Nil(t, err)
	})

	t.Run("diff by os.stdin with flags", func(t *testing.T) {
		monkey.Patch(ytbx.IsStdin, func(location string) bool {
			return true
		})
		defer monkey.UnpatchAll()

		err := mockStdin("testdata/pod-full.yaml")
		if err != nil {
			panic(err)
		}
		defer func() { os.Stdin = originalOSStdin }()

		cmd := NewCmdDiff()
		assert.Nil(t, cmd.Flags().Set("diff-mode", DiffModeIgnoreAdded))
		assert.Nil(t, cmd.Flags().Set("output", OutputRaw))
		assert.Nil(t, cmd.Flags().Set("sort-by-kubernetes-resource", "true"))
		assert.Nil(t, cmd.Flags().Set("swap", "true"))
		cmd.SetArgs([]string{"testdata/pod-full.yaml"})
		err = cmd.Execute()
		assert.Nil(t, err)
	})
}

func Test_liveDiffWithStdin(t *testing.T) {
	t.Run("create normalizer failed", func(t *testing.T) {
		monkey.Patch(diff.NewDefaultIgnoreNormalizer, func(paths []string) (diff.Normalizer, error) {
			return nil, errors.New("mock create normalizer error")
		})
		defer monkey.UnpatchAll()

		err := mockStdin("testdata/pod-full.yaml")
		if err != nil {
			panic(err)
		}
		defer func() { os.Stdin = originalOSStdin }()

		err = liveDiffWithStdin()
		assert.NotNil(t, err)
	})

	t.Run("calculate diff failed", func(t *testing.T) {
		monkey.Patch(
			diff.Diff,
			func(config, live *unstructured.Unstructured, opts ...diff.Option) (*diff.DiffResult, error) {
				return nil, errors.New("mock calculate diff error")
			},
		)
		defer monkey.UnpatchAll()

		err := mockStdin("testdata/pod-full.yaml")
		if err != nil {
			panic(err)
		}
		defer func() { os.Stdin = originalOSStdin }()

		err = liveDiffWithStdin()
		assert.NotNil(t, err)
	})

	t.Run("liveDiffWithStdin success", func(t *testing.T) {
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
	t.Run("read fromLocation error", func(t *testing.T) {
		err := liveDiffWithFile("no/such/file", "no/such/file")
		assert.NotNil(t, err)
	})

	t.Run("read toLocation error", func(t *testing.T) {
		err := liveDiffWithFile("testdata/pod1.yaml", "no/such/file")
		assert.NotNil(t, err)
	})

	t.Run("create normalizer failed", func(t *testing.T) {
		monkey.Patch(diff.NewDefaultIgnoreNormalizer, func(paths []string) (diff.Normalizer, error) {
			return nil, errors.New("mock create normalizer error")
		})
		defer monkey.UnpatchAll()

		err := liveDiffWithFile("testdata/pod1.yaml", "testdata/pod1.yaml")
		assert.NotNil(t, err)
	})

	t.Run("calculate diff failed", func(t *testing.T) {
		monkey.Patch(
			diff.Diff,
			func(config, live *unstructured.Unstructured, opts ...diff.Option) (*diff.DiffResult, error) {
				return nil, errors.New("mock calculate diff error")
			},
		)
		defer monkey.UnpatchAll()

		err := liveDiffWithFile("testdata/pod1.yaml", "testdata/pod1.yaml")
		assert.NotNil(t, err)
	})

	t.Run("liveDiffWithFile success", func(t *testing.T) {
		err := liveDiffWithFile("testdata/pod1.yaml", "testdata/pod1.yaml")
		assert.Nil(t, err)
	})
}

func Test_loadFile(t *testing.T) {
	t.Run("load failed", func(t *testing.T) {
		_, err := loadFile("no/such/file")
		assert.NotNil(t, err)
	})

	t.Run("load success", func(t *testing.T) {
		_, err := loadFile("testdata/pod1.yaml")
		assert.Nil(t, err)
	})
}
