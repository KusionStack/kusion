package crd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrdVisitor(t *testing.T) {
	t.Run("read single file", func(t *testing.T) {
		visitor := crdVisitor{
			Path: "./testdata/one.yaml",
		}
		objs, err := visitor.Visit()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(objs))
	})

	t.Run("read multi files", func(t *testing.T) {
		visitor := crdVisitor{
			Path: "./testdata",
		}
		objs, err := visitor.Visit()
		assert.Nil(t, err)
		assert.Equal(t, 3, len(objs))
	})
}

func Test_ignoreFile(t *testing.T) {
	t.Run("not ignore .YAML file", func(t *testing.T) {
		flag := ignoreFile("foo.YAML", FileExtensions)
		assert.False(t, flag)
	})

	t.Run("not ignore .yaml file", func(t *testing.T) {
		flag := ignoreFile("foo.yaml", FileExtensions)
		assert.False(t, flag)
	})

	t.Run("not ignore .yml file", func(t *testing.T) {
		flag := ignoreFile("foo.yml", FileExtensions)
		assert.False(t, flag)
	})

	t.Run("ignore .go file", func(t *testing.T) {
		flag := ignoreFile("bar.go", FileExtensions)
		assert.True(t, flag)
	})
}
