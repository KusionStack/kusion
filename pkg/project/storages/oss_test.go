package storages

import (
	"reflect"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func mockOssStorage() *OssStorage {
	return &OssStorage{bucket: &oss.Bucket{}, prefix: "release"}
}

func TestNewOssStorage(t *testing.T) {
	type args struct {
		bucket *oss.Bucket
		prefix string
	}
	tests := []struct {
		name string
		args args
		want *OssStorage
	}{
		{
			name: "",
			args: args{
				bucket: &oss.Bucket{},
				prefix: "",
			},
			want: &OssStorage{
				bucket: &oss.Bucket{},
				prefix: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewOssStorage(tt.args.bucket, tt.args.prefix); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOssStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOssStorage_Get(t *testing.T) {
	tests := []struct {
		name    string
		want    map[string][]string
		wantErr bool
	}{
		{
			name:    "Get projects successfully",
			want:    map[string][]string{"": {"releases/project1", "releases/project2"}, "releases/project1": {"releases/project2"}, "releases/project2": {"releases/project1"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockey.PatchConvey("mock oss operation", t, func() {
				mockey.Mock(oss.Bucket.ListObjects).Return(oss.ListObjectsResult{
					CommonPrefixes: []string{"releases/project1/", "releases/project2/"},
					IsTruncated:    false,
				}, nil).Build()
				r, err := mockOssStorage().Get()
				assert.NoError(t, err)
				assert.Equal(t, tt.want, r)
			})
		})
	}
}
