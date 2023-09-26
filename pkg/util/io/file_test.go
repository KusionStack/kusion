//go:build !arm64
// +build !arm64

package io

import (
	"github.com/bytedance/mockey"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDir(t *testing.T) {
	type args struct {
		path string
	}
	type checkFunc func(t *testing.T, result bool, err error)
	tests := []struct {
		name  string
		args  args
		check checkFunc
	}{
		{
			name: "t1",
			args: args{
				path: "./",
			},
			check: func(t *testing.T, result bool, err error) {
				assert.True(t, result)
				assert.Nil(t, err)
			},
		},
		{
			name: "t2",
			args: args{
				path: "./nonDir",
			},
			check: func(t *testing.T, result bool, err error) {
				assert.False(t, result)
				assert.NotNil(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsDir(tt.args.path)
			tt.check(t, result, err)
		})
	}
}

func TestIsFile(t *testing.T) {
	type args struct {
		path string
	}
	type checkFunc func(t *testing.T, result bool, err error)
	tests := []struct {
		name  string
		args  args
		check checkFunc
	}{
		{
			name: "t1",
			args: args{
				path: "./",
			},
			check: func(t *testing.T, result bool, err error) {
				assert.False(t, result)
				assert.Nil(t, err)
			},
		},
		{
			name: "t2",
			args: args{
				path: "./nonDir",
			},
			check: func(t *testing.T, result bool, err error) {
				assert.False(t, result)
				assert.NotNil(t, err)
			},
		},
		{
			name: "t3",
			args: args{
				path: "./file.go",
			},
			check: func(t *testing.T, result bool, err error) {
				assert.True(t, result)
				assert.Nil(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsFile(tt.args.path)
			tt.check(t, result, err)
		})
	}
}

func TestIsFileOrDirExist(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "t1",
			args: args{
				path: "./",
			},
			want: true,
		},
		{
			name: "t2",
			args: args{
				path: "./nonDir",
			},
			want: false,
		},
		{
			name: "t3",
			args: args{
				path: "./file.go",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFileOrDirExist(tt.args.path)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestRenamePath(t *testing.T) {
	type args struct {
		oldPath string
		newPath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t1",
			args: args{
				oldPath: "./",
				newPath: "./nondir",
			},
		},
	}

	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			mockey.Mock(os.MkdirAll).To(func(string, os.FileMode) error {
				return nil
			}).Build()
			mockey.Mock(os.Rename).To(func(oldpath, newpath string) error {
				return nil
			}).Build()
			err := RenamePath(tt.args.oldPath, tt.args.newPath)
			assert.Nil(t, err)
		})
	}
}
