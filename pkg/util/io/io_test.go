//go:build !arm64
// +build !arm64

package io

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"bou.ke/monkey"

	"github.com/stretchr/testify/assert"
)

func TestReadStdinInput(t *testing.T) {
	defer monkey.UnpatchAll()
	monkey.Patch((*os.File).Stat, func(file *os.File) (os.FileInfo, error) {
		return mockInfo{}, nil
	})
	input := "hello world!"
	monkey.Patch(bufio.NewReader, func(rd io.Reader) *bufio.Reader {
		return bufio.NewReaderSize(bytes.NewReader([]byte(input)), 4096)
	})
	result, err := ReadStdinInput()
	assert.Equal(t, input, result)
	assert.Nil(t, err)
}

type mockInfo struct{}

func (m mockInfo) Name() string {
	return ""
}

func (m mockInfo) Size() int64 {
	return 4096
}

func (m mockInfo) Mode() os.FileMode {
	return os.FileMode(0o777)
}

func (m mockInfo) IsDir() bool {
	return false
}

func (m mockInfo) Sys() interface{} {
	return nil
}

func (m mockInfo) ModTime() time.Time {
	return time.Time{}
}
