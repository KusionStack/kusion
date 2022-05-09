package oss

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"
)

var ErrFake = errors.New("fake error")

func mockOSSNew(mockErr error) {
	monkey.Patch(oss.New, func(endpoint, accessKeyID, accessKeySecret string, options ...oss.ClientOption) (*oss.Client, error) {
		return &oss.Client{}, mockErr
	})
}

func mockBucket(client oss.Client, mockErr error) {
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Bucket", func(_ oss.Client, _ string) (*oss.Bucket, error) {
		if mockErr == nil {
			return &oss.Bucket{
				Client:     oss.Client{},
				BucketName: "fake-bucket",
			}, nil
		} else {
			return nil, mockErr
		}
	})
}

func mockBucketIsObjectExist(bucket oss.Bucket, mockBool bool, mockErr error) {
	monkey.PatchInstanceMethod(reflect.TypeOf(bucket), "IsObjectExist", func(_ oss.Bucket, _ string, options ...oss.Option) (bool, error) {
		return mockBool, mockErr
	})
}

func mockBucketGetObject(bucket oss.Bucket, mockResult string, mockErr error) {
	monkey.PatchInstanceMethod(reflect.TypeOf(bucket), "GetObject", func(_ oss.Bucket, _ string, options ...oss.Option) (io.ReadCloser, error) {
		readerString := strings.NewReader(mockResult)
		readCloserString := ioutil.NopCloser(readerString)
		return readCloserString, mockErr
	})
}

func mockBucketGetObjectMeta(bucket oss.Bucket, mockErr error) {
	monkey.PatchInstanceMethod(reflect.TypeOf(bucket), "GetObjectMeta", func(_ oss.Bucket, _ string, options ...oss.Option) (http.Header, error) {
		h := http.Header{}
		h.Add("Content-Type", "application/json")
		return h, mockErr
	})
}

func mockBucketPutObjectFromFile(bucket oss.Bucket, mockErr error) {
	monkey.PatchInstanceMethod(reflect.TypeOf(bucket), "PutObjectFromFile", func(_ oss.Bucket, _, _ string, options ...oss.Option) error {
		return mockErr
	})
}

func mockBucketPutObject(bucket oss.Bucket, mockErr error) {
	monkey.PatchInstanceMethod(reflect.TypeOf(bucket), "PutObject", func(_ oss.Bucket, objectKey string, reader io.Reader, options ...oss.Option) error {
		return mockErr
	})
}

func mockBucketGetObjectToFile(bucket oss.Bucket, mockErr error) {
	monkey.PatchInstanceMethod(reflect.TypeOf(bucket), "GetObjectToFile", func(_ oss.Bucket, objectKey, filePath string, options ...oss.Option) error {
		return mockErr
	})
}

func mockBucketDeleteObject(bucket oss.Bucket, mockErr error) {
	monkey.PatchInstanceMethod(reflect.TypeOf(bucket), "DeleteObject", func(_ oss.Bucket, objectKey string, options ...oss.Option) error {
		return mockErr
	})
}

func mockBucketListObjectsV2(bucket oss.Bucket, mockErr error) {
	monkey.PatchInstanceMethod(reflect.TypeOf(bucket), "ListObjectsV2", func(_ oss.Bucket, options ...oss.Option) (oss.ListObjectsResultV2, error) {
		return oss.ListObjectsResultV2{}, mockErr
	})
}

func TestNewRemoteClient(t *testing.T) {
	type args struct {
		endpoint      string
		accessKey     string
		secretKey     string
		securityToken string
		bucket        string
	}
	tests := []struct {
		name    string
		args    args
		want    *RemoteClient
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			args: args{
				endpoint:      "fake-endpoint",
				accessKey:     "fake-access-key",
				secretKey:     "fake-secret-key",
				securityToken: "fake-security-token",
				bucket:        "fake-bucket",
			},
			want: &RemoteClient{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			wantErr: false,
			preRun: func() {
				mockOSSNew(nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-new",
			args: args{
				endpoint:      "fake-endpoint",
				accessKey:     "fake-access-key",
				secretKey:     "fake-secret-key",
				securityToken: "fake-security-token",
				bucket:        "fake-bucket",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockOSSNew(ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-strconv",
			args: args{
				endpoint:      "fake-endpoint",
				accessKey:     "fake-access-key",
				secretKey:     "fake-secret-key",
				securityToken: "fake-security-token",
				bucket:        "fake-bucket",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				os.Setenv(EnvTimeout, "invalid-value")
				mockOSSNew(nil)
			},
			postRun: func() {
				os.Unsetenv(EnvTimeout)
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preRun()
			got, err := NewRemoteClient(tt.args.endpoint, tt.args.accessKey, tt.args.secretKey, tt.args.securityToken, tt.args.bucket)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRemoteClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRemoteClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteClient_WithBucket(t *testing.T) {
	type fields struct {
		ossClient  *oss.Client
		bucketName string
	}
	type args struct {
		bucket string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *RemoteClient
	}{
		{
			name: "success",
			fields: fields{
				ossClient: &oss.Client{},
			},
			args: args{
				bucket: "fake-bucket",
			},
			want: &RemoteClient{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RemoteClient{
				ossClient:  tt.fields.ossClient,
				bucketName: tt.fields.bucketName,
			}
			if got := c.WithBucket(tt.args.bucket); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoteClient.WithBucket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteClient_GetOssClient(t *testing.T) {
	type fields struct {
		ossClient  *oss.Client
		bucketName string
	}
	tests := []struct {
		name   string
		fields fields
		want   *oss.Client
	}{
		{
			name: "success",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			want: &oss.Client{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RemoteClient{
				ossClient:  tt.fields.ossClient,
				bucketName: tt.fields.bucketName,
			}
			if got := c.GetOssClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoteClient.GetOssClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteClient_Get(t *testing.T) {
	type fields struct {
		ossClient  *oss.Client
		bucketName string
	}
	type args struct {
		objectName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Payload
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				objectName: "fake-object",
			},
			want: &Payload{
				Data: []byte("fake-result"),
				MD5:  []byte{244, 37, 197, 253, 104, 169, 212, 94, 133, 112, 212, 65, 103, 132, 52, 53},
			},
			wantErr: false,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, true, nil)
				mockBucketGetObject(oss.Bucket{}, "fake-result", nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "success-for-empty-payload",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				objectName: "fake-object",
			},
			want:    nil,
			wantErr: false,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, true, nil)
				mockBucketGetObject(oss.Bucket{}, "", nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-bucket",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				objectName: "fake-object",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-IsObjectExist",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				objectName: "fake-object",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, true, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-IsObjectExist-2",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				objectName: "fake-object",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, false, nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-GetObject",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				objectName: "fake-object",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, true, nil)
				mockBucketGetObject(oss.Bucket{}, "", ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RemoteClient{
				ossClient:  tt.fields.ossClient,
				bucketName: tt.fields.bucketName,
			}
			tt.preRun()
			got, err := c.Get(tt.args.objectName)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoteClient.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoteClient.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteClient_GetObjectMeta(t *testing.T) {
	type fields struct {
		ossClient  *oss.Client
		bucketName string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    http.Header
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			want: map[string][]string{
				"Content-Type": {"application/json"},
			},
			wantErr: false,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, true, nil)
				mockBucketGetObjectMeta(oss.Bucket{}, nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-bucket",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-IsObjectExist",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, true, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-IsObjectExist-2",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, false, nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-GetObjectMeta",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, true, nil)
				mockBucketGetObjectMeta(oss.Bucket{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RemoteClient{
				ossClient:  tt.fields.ossClient,
				bucketName: tt.fields.bucketName,
			}
			tt.preRun()
			got, err := c.GetObjectMeta(tt.args.key)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoteClient.GetObjectMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoteClient.GetObjectMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteClient_ListObjects(t *testing.T) {
	type fields struct {
		ossClient  *oss.Client
		bucketName string
	}
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *oss.ListObjectsResultV2
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				prefix: "",
			},
			want:    &oss.ListObjectsResultV2{},
			wantErr: false,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketListObjectsV2(oss.Bucket{}, nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-bucket",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				prefix: "",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-ListObjectsV2",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				prefix: "",
			},
			want:    nil,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketListObjectsV2(oss.Bucket{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RemoteClient{
				ossClient:  tt.fields.ossClient,
				bucketName: tt.fields.bucketName,
			}
			tt.preRun()
			got, err := c.ListObjects(tt.args.prefix)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoteClient.ListObjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoteClient.ListObjects() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteClient_PutObjectFromFile(t *testing.T) {
	type fields struct {
		ossClient  *oss.Client
		bucketName string
	}
	type args struct {
		key  string
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key:  "fake-key",
				path: "fake-path",
			},
			wantErr: false,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketPutObjectFromFile(oss.Bucket{}, nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-bucket",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key:  "fake-key",
				path: "fake-path",
			},
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-PutObjectFromFile",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key:  "fake-key",
				path: "fake-path",
			},
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketPutObjectFromFile(oss.Bucket{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RemoteClient{
				ossClient:  tt.fields.ossClient,
				bucketName: tt.fields.bucketName,
			}
			tt.preRun()
			err := c.PutObjectFromFile(tt.args.key, tt.args.path)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoteClient.PutObjectFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoteClient_PutObject(t *testing.T) {
	type fields struct {
		ossClient  *oss.Client
		bucketName string
	}
	type args struct {
		key string
		val []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
				val: []byte("fake-value"),
			},
			wantErr: false,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketPutObject(oss.Bucket{}, nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-bucket",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
				val: []byte("fake-value"),
			},
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-PutObject",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
				val: []byte("fake-value"),
			},
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketPutObject(oss.Bucket{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RemoteClient{
				ossClient:  tt.fields.ossClient,
				bucketName: tt.fields.bucketName,
			}
			tt.preRun()
			err := c.PutObject(tt.args.key, tt.args.val)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoteClient.PutObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoteClient_GetObjectToFile(t *testing.T) {
	type fields struct {
		ossClient  *oss.Client
		bucketName string
	}
	type args struct {
		key     string
		dstPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key:     "fake-key",
				dstPath: "fake-dstPath",
			},
			wantErr: false,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketGetObjectToFile(oss.Bucket{}, nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-bucket",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key:     "fake-key",
				dstPath: "fake-dstPath",
			},
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-GetObjectToFile",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key:     "fake-key",
				dstPath: "fake-dstPath",
			},
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketGetObjectToFile(oss.Bucket{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RemoteClient{
				ossClient:  tt.fields.ossClient,
				bucketName: tt.fields.bucketName,
			}
			tt.preRun()
			err := c.GetObjectToFile(tt.args.key, tt.args.dstPath)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoteClient.GetObjectToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoteClient_DeleteObject(t *testing.T) {
	type fields struct {
		ossClient  *oss.Client
		bucketName string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			wantErr: false,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketDeleteObject(oss.Bucket{}, nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-bucket",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-DeleteObject",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketDeleteObject(oss.Bucket{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RemoteClient{
				ossClient:  tt.fields.ossClient,
				bucketName: tt.fields.bucketName,
			}
			tt.preRun()
			err := c.DeleteObject(tt.args.key)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoteClient.DeleteObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoteClient_IsObjectExist(t *testing.T) {
	type fields struct {
		ossClient  *oss.Client
		bucketName string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
		preRun  func()
		postRun func()
	}{
		{
			name: "success",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			want:    true,
			wantErr: false,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, true, nil)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-bucket",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			want:    false,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
		{
			name: "fail-for-IsObjectExist",
			fields: fields{
				ossClient:  &oss.Client{},
				bucketName: "fake-bucket",
			},
			args: args{
				key: "fake-key",
			},
			want:    false,
			wantErr: true,
			preRun: func() {
				mockBucket(oss.Client{}, nil)
				mockBucketIsObjectExist(oss.Bucket{}, false, ErrFake)
			},
			postRun: func() {
				defer monkey.UnpatchAll()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RemoteClient{
				ossClient:  tt.fields.ossClient,
				bucketName: tt.fields.bucketName,
			}
			tt.preRun()
			got, err := c.IsObjectExist(tt.args.key)
			tt.postRun()
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoteClient.IsObjectExist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RemoteClient.IsObjectExist() = %v, want %v", got, tt.want)
			}
		})
	}
}
