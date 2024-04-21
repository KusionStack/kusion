package kfile

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
)

const mockHomeDir = "/home/test"

func TestFileExists(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "file exists",
			args: args{
				filename: "./testdata/test.txt",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "file does not exists",
			args: args{
				filename: "test.txt",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FileExists(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKusionDataFolder(t *testing.T) {
	// Run test
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			want:    filepath.Join(mockHomeDir, ".kusion"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			// Mock data
			os.Setenv(EnvKusionHome, "")
			mockUserCurrent()
			mockMkdirall()
			got, err := KusionDataFolder()
			if (err != nil) != tt.wantErr {
				t.Errorf("KusionDataFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("KusionDataFolder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockUserCurrent() {
	mockey.Mock(user.Current).To(func() (*user.User, error) {
		return &user.User{
			HomeDir: mockHomeDir,
		}, nil
	}).Build()
}

func mockMkdirall() {
	mockey.Mock(os.MkdirAll).To(func(path string, perm os.FileMode) error {
		return nil
	}).Build()
}
