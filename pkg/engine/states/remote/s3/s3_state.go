package s3

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"

	"kusionstack.io/kusion/pkg/engine/states"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var ErrS3NoExist = errors.New("s3: key not exist")

const S3StateName = "kusion_state.json"

var _ states.StateStorage = &S3State{}

type S3State struct {
	sess       *session.Session
	bucketName string
}

func NewS3State(endPoint, accessKeyID, accessKeySecret, bucketName string, region string) (*S3State, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyID, accessKeySecret, ""),
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

func (s *S3State) Apply(state *states.State) error {
	jsonByte, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	prefix := state.Tenant + "/" + state.Project + "/" + state.Stack + "/" + S3StateName
	s3Client := s3.New(s.sess)
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(prefix),
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

func (s *S3State) GetLatestState(query *states.StateQuery) (*states.State, error) {
	prefix := query.Tenant + "/" + query.Project + "/" + query.Stack + "/" + S3StateName
	s3Client := s3.New(s.sess)

	params := &s3.ListObjectsInput{
		Bucket:    aws.String(s.bucketName),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(prefix),
	}

	objects, err := s3Client.ListObjects(params)
	if err != nil {
		return nil, err
	}

	if len(objects.Contents) == 0 {
		return nil, nil
	}

	out, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    &prefix,
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()

	data, err := ioutil.ReadAll(out.Body)
	if err != nil {
		return nil, err
	}
	state := &states.State{}
	err = json.Unmarshal(data, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}
