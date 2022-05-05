package io

import (
	"os"
	"testing"

	"bou.ke/monkey"
)

func TestCreateDirIfNotExist(t *testing.T) {
	defer monkey.UnpatchAll()
	monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
		return nil
	})
	err := CreateDirIfNotExist("./nondir/nonfile")
	if err != nil {
		t.Fatal(err)
	}
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
