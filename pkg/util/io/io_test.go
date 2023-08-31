//go:build !arm64
// +build !arm64

package io

import (
	"bufio"
	"bytes"
	"github.com/bytedance/mockey"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadStdinInput(t *testing.T) {
	mockey.PatchConvey("test read stdin input", t, func() {
		mockey.Mock((*os.File).Stat).To(func(file *os.File) (os.FileInfo, error) {
			return mockInfo{}, nil
		}).Build()
		input := "hello world!"
		mockey.Mock(bufio.NewReader).To(func(rd io.Reader) *bufio.Reader {
			return bufio.NewReaderSize(bytes.NewReader([]byte(input)), 4096)
		}).Build()
		result, err := ReadStdinInput()
		assert.Equal(t, input, result)
		assert.Nil(t, err)
	})
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
