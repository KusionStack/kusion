package s3

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
)

var ErrS3NoExist = errors.New("s3: key not exist")

const (
	deprecatedKusionStateFile = "kusion_state.json"
	KusionStateFile           = "kusion_state.yaml"
)

var _ states.StateStorage = &S3State{}

type S3State struct {
	sess       *session.Session
	bucketName string
}

func NewS3State(endpoint, accessKeyID, accessKeySecret, bucketName string, region string) (*S3State, error) {
	config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyID, accessKeySecret, ""),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false),
	}
	if endpoint != "" {
		config.Endpoint = aws.String(endpoint)
	}
	sess, err := session.NewSession(config)
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

	var prefix string
	if state.Tenant != "" {
		prefix = state.Tenant + "/" + state.Project + "/" + state.Stack + "/" + KusionStateFile
	} else {
		prefix = state.Project + "/" + state.Stack + "/" + KusionStateFile
	}

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
	var prefix string
	if query.Tenant != "" {
		prefix = query.Tenant + "/" + query.Project + "/" + query.Stack + "/" + KusionStateFile
	} else {
		prefix = query.Project + "/" + query.Stack + "/" + KusionStateFile
	}
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
		var deprecatedPrefix string
		deprecatedPrefix, err = s.usingDeprecatedStateFilePrefix(query)
		if err != nil {
			return nil, err
		}
		if deprecatedPrefix == "" {
			return nil, nil
		}
		prefix = deprecatedPrefix
		log.Infof("using deprecated s3 kusion state file %s", prefix)
	}

	out, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    &prefix,
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
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

func (s *S3State) usingDeprecatedStateFilePrefix(query *states.StateQuery) (string, error) {
	var prefix string
	if query.Tenant != "" {
		prefix = query.Tenant + "/" + query.Project + "/" + query.Stack + "/" + deprecatedKusionStateFile
	} else {
		prefix = query.Project + "/" + query.Stack + "/" + deprecatedKusionStateFile
	}
	s3Client := s3.New(s.sess)

	params := &s3.ListObjectsInput{
		Bucket:    aws.String(s.bucketName),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(prefix),
	}

	objects, err := s3Client.ListObjects(params)
	if err != nil {
		return "", err
	}
	if len(objects.Contents) == 0 {
		return "", nil
	}
	return prefix, nil
}
