package states

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

var ErrS3NoExist = errors.New("s3: key not exist")

type S3State struct {
	sess       *session.Session
	bucketName string
}

func NewS3State(endPoint, accessKeyId, accessKeySecret, bucketName string, region string) (*S3State, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyId, accessKeySecret, ""),
		Endpoint:         aws.String(endPoint),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false),
	})
	if err != nil {
		return nil, err
	}
	s3State := &S3State{
		sess:       sess,
		bucketName: bucketName,
	}
	return s3State, nil
}

func (s *S3State) Apply(state *State) error {
	u, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	jsonByte, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	prefix := state.Tenant + "/" + state.Project + "/" + state.Stack
	svc := s3.New(s.sess)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(prefix + u.String()),
		Body:   bytes.NewReader(jsonByte),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *S3State) Delete(id string) error {
	panic("implement me")
}

func (s *S3State) GetLatestState(query *StateQuery) (*State, error) {
	prefix := query.Tenant + "/" + query.Project + "/" + query.Stack
	svc := s3.New(s.sess)

	params := &s3.ListObjectsInput{
		Bucket:    aws.String(s.bucketName),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(prefix),
	}

	objects, err := svc.ListObjects(params)
	if err != nil {
		return nil, err
	}

	var result *s3.Object
	if len(objects.Contents) == 0 {
		return nil, ErrS3NoExist
	}
	for _, obj := range objects.Contents {
		if result == nil || result.LastModified.UnixNano() < obj.LastModified.UnixNano() {
			result = obj
		}
	}

	if result == nil {
		return nil, ErrS3NoExist
	}

	out, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    result.Key,
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()

	data, err := ioutil.ReadAll(out.Body)
	if err != nil {
		return nil, err
	}
	state := &State{}
	err = json.Unmarshal(data, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}
