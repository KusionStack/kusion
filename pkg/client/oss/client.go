package oss

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"
)

const (
	UserAgent  = "Kusion"
	EnvTimeout = "OSS_TIMEOUT"
)

// Timeout default timeout is 5 second
var Timeout = 5

var ErrNoExist = errors.New("oss: key not exist")

type RemoteClient struct {
	ossClient  *oss.Client
	bucketName string
}

type Config struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	SecurityToken string
	BucketName    string
}

func NewRemoteClient(endpoint, accessKey, secretKey, securityToken, bucket string) (*RemoteClient, error) {
	var options []oss.ClientOption
	if securityToken != "" {
		options = append(options, oss.SecurityToken(securityToken))
	}

	options = append(options, oss.UserAgent(fmt.Sprintf("%s/%s", UserAgent, "FakeVersion")))

	client, err := oss.New(endpoint, accessKey, secretKey, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to new oss client with endpoint %s as %#v", endpoint, err)
	}

	if v := os.Getenv(EnvTimeout); len(v) > 0 {
		Timeout, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid OSS_TIMEOUT in env: %s", v)
		}
	}

	return &RemoteClient{ossClient: client, bucketName: bucket}, nil
}

func (c *RemoteClient) WithBucket(bucket string) *RemoteClient {
	c.bucketName = bucket
	return c
}

func (c *RemoteClient) GetOssClient() *oss.Client {
	return c.ossClient
}

// Get obtains target object from oss bucket.
func (c *RemoteClient) Get(objectName string) (*Payload, error) {
	bucket, err := c.ossClient.Bucket(c.bucketName)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket %s: %#v", c.bucketName, err)
	}

	if exist, err := bucket.IsObjectExist(objectName); err != nil {
		return nil, fmt.Errorf("estimating object %s is exist got an error: %#v", objectName, err)
	} else if !exist {
		return nil, errors.Wrapf(ErrNoExist, "key:[%s]", objectName)
	}

	var options []oss.Option

	output, err := bucket.GetObject(objectName, options...)
	if err != nil {
		return nil, fmt.Errorf("error getting object: %#v", err)
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, output); err != nil {
		return nil, fmt.Errorf("failed to read remote state: %s", err)
	}

	sum := md5.Sum(buf.Bytes())
	payload := &Payload{
		Data: buf.Bytes(),
		MD5:  sum[:],
	}

	// If there was no data, then return nil
	if len(payload.Data) == 0 {
		return nil, nil
	}

	return payload, nil
}

func (c *RemoteClient) GetObjectMeta(key string) (http.Header, error) {
	bucket, err := c.ossClient.Bucket(c.bucketName)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket %s: %#v", c.bucketName, err)
	}

	if exist, err := bucket.IsObjectExist(key); err != nil {
		return nil, fmt.Errorf("estimating object %s is exist got an error: %#v", key, err)
	} else if !exist {
		return nil, errors.Wrapf(ErrNoExist, "key:[%s]", key)
	}

	var options []oss.Option
	header, err := bucket.GetObjectMeta(key, options...)
	if err != nil {
		return nil, fmt.Errorf("error getting object: %#v", err)
	}

	return header, nil
}

func (c *RemoteClient) ListObjects(prefix string) (*oss.ListObjectsResultV2, error) {
	bucket, err := c.ossClient.Bucket(c.bucketName)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket %s: %#v", c.bucketName, err)
	}

	option := oss.Prefix(prefix)

	r, err := bucket.ListObjectsV2(option)
	if err != nil {
		return nil, fmt.Errorf("list object error: %#v", err)
	}

	return &r, nil
}

func (c *RemoteClient) PutObjectFromFile(key, path string) error {
	bc, err := c.ossClient.Bucket(c.bucketName)
	if err != nil {
		return fmt.Errorf("get bucket failed: %s, err: %#v", c.bucketName, err)
	}

	if err = bc.PutObjectFromFile(key, path); err != nil {
		return errors.Wrapf(err, "bucket %s", c.bucketName)
	}

	return nil
}

func (c *RemoteClient) PutObject(key string, val []byte) error {
	bc, err := c.ossClient.Bucket(c.bucketName)
	if err != nil {
		return fmt.Errorf("put bucket failed: %s, err: %#v", c.bucketName, err)
	}

	if err = bc.PutObject(key, bytes.NewReader(val)); err != nil {
		return errors.Wrapf(err, "bucket %s", c.bucketName)
	}

	return nil
}

func (c *RemoteClient) GetObjectToFile(key, dstPath string) error {
	bc, err := c.ossClient.Bucket(c.bucketName)
	if err != nil {
		return fmt.Errorf("get bucket failed: %s, err: %#v", c.bucketName, err)
	}

	if err = bc.GetObjectToFile(key, dstPath); err != nil {
		return errors.Wrapf(err, "bucket %s key %s dstPath %s", c.bucketName, key, dstPath)
	}

	return nil
}

func (c *RemoteClient) DeleteObject(key string) error {
	bc, err := c.ossClient.Bucket(c.bucketName)
	if err != nil {
		return fmt.Errorf("del bucket failed: %s, err: %#v", c.bucketName, err)
	}

	if err = bc.DeleteObject(key); err != nil {
		return errors.Wrapf(err, "bucket %s key %s ", c.bucketName, key)
	}

	return nil
}

func (c *RemoteClient) IsObjectExist(key string) (bool, error) {
	bucket, err := c.ossClient.Bucket(c.bucketName)
	if err != nil {
		return false, fmt.Errorf("error getting bucket %s: %#v", c.bucketName, err)
	}

	if exist, err := bucket.IsObjectExist(key); err != nil {
		return false, fmt.Errorf("estimating object %s is exist got an error: %#v", key, err)
	} else {
		return exist, nil
	}
}
