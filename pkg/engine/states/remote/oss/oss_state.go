package oss

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/engine/states"
)

var ErrOSSNoExist = errors.New("oss: key not exist")

const OSSStateName = "kusion_state.json"

var _ states.StateStorage = &OssState{}

type OssState struct {
	bucket *oss.Bucket
}

func NewOSSState(endPoint, accessKeyID, accessKeySecret, bucketName string) (*OssState, error) {
	var ossClient *oss.Client
	var err error
	ossClient, err = oss.New(endPoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}
	var ossBucket *oss.Bucket
	ossBucket, err = ossClient.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	ossState := &OssState{
		bucket: ossBucket,
	}
	return ossState, nil
}

func (s *OssState) Apply(state *states.State) error {
	jsonByte, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	prefix := state.Tenant + "/" + state.Project + "/" + state.Stack + "/" + OSSStateName
	err = s.bucket.PutObject(prefix, bytes.NewReader(jsonByte))
	if err != nil {
		return err
	}
	return nil
}

func (s *OssState) Delete(id string) error {
	panic("implement me")
}

func (s *OssState) GetLatestState(query *states.StateQuery) (*states.State, error) {
	prefix := query.Tenant + "/" + query.Project + "/" + query.Stack + "/" + OSSStateName
	objects, err := s.bucket.ListObjects(oss.Delimiter("/"), oss.Prefix(prefix))
	if err != nil {
		return nil, err
	}

	if len(objects.Objects) == 0 {
		return nil, nil
	}

	body, err := s.bucket.GetObject(prefix)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	state := &states.State{}
	// JSON is a subset of YAML. Please check FileSystemState.GetLatestState for detail explanation
	err = yaml.Unmarshal(data, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}
