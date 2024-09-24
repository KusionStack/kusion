package storages

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func mockS3Storage() *S3Storage {
	return &S3Storage{
		s3:     &s3.S3{},
		bucket: "",
		prefix: "release",
	}
}

func TestNewS3Storage(t *testing.T) {
	type args struct {
		s3     *s3.S3
		bucket string
		prefix string
	}
	tests := []struct {
		name string
		args args
		want *S3Storage
	}{
		{name: "", args: args{
			s3:     &s3.S3{},
			bucket: "",
			prefix: "",
		}, want: &S3Storage{
			s3:     &s3.S3{},
			bucket: "",
			prefix: "",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewS3Storage(tt.args.s3, tt.args.bucket, tt.args.prefix); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewS3Storage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS3Storage_Get(t *testing.T) {
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
			mockey.PatchConvey("mock s3 operation", t, func() {
				isTruncated := false
				dir1 := "releases/project1/"
				dir2 := "releases/project2/"
				commonPrefix := []*s3.CommonPrefix{{Prefix: &dir1}, {Prefix: &dir2}}

				mockey.Mock((*s3.S3).ListObjectsV2).Return(&s3.ListObjectsV2Output{
					CommonPrefixes: commonPrefix,
					IsTruncated:    &isTruncated,
				}, nil).Build()
				r, err := mockS3Storage().Get()
				assert.NoError(t, err)
				assert.Equal(t, tt.want, r)
			})
		})
	}
}
