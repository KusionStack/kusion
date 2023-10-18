//go:build !arm64
// +build !arm64

package io

import (
	"os"
	"testing"

	"github.com/bytedance/mockey"
)

func TestCreateDirIfNotExist(t *testing.T) {
	mockey.PatchConvey("test create dir if not exist", t, func() {
		mockey.Mock(os.MkdirAll).To(func(path string, perm os.FileMode) error {
			return nil
		}).Build()
		err := CreateDirIfNotExist("./nondir/nonfile")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestValidDir(t *testing.T) {
	_, err := OutDir("./")
	if err != nil {
		t.Fatal(err)
	}
}

func TestInvalidDir(t *testing.T) {
	_, err := OutDir("./nondir")
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestNotDir(t *testing.T) {
	_, err := OutDir("./dir_test.go")
	if err == nil {
		t.Fatal("expected an error")
	}
}
