package oss

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
)

var ErrOSSNoExist = errors.New("oss: key not exist")

const (
	deprecatedKusionStateFile = "kusion_state.json"
	KusionStateFile           = "kusion_state.yaml"
)

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

	var prefix string
	if state.Tenant != "" {
		prefix = state.Tenant + "/" + state.Project + "/" + state.Stack + "/" + KusionStateFile
	} else {
		prefix = state.Project + "/" + state.Stack + "/" + KusionStateFile
	}

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
	var prefix string
	if query.Tenant != "" {
		prefix = query.Tenant + "/" + query.Project + "/" + query.Stack + "/" + KusionStateFile
	} else {
		prefix = query.Project + "/" + query.Stack + "/" + KusionStateFile
	}

	objects, err := s.bucket.ListObjects(oss.Delimiter("/"), oss.Prefix(prefix))
	if err != nil {
		return nil, err
	}

	if len(objects.Objects) == 0 {
		var deprecatedPrefix string
		deprecatedPrefix, err = s.usingDeprecatedStateFilePrefix(query)
		if err != nil {
			return nil, err
		}
		if deprecatedPrefix == "" {
			return nil, nil
		}
		prefix = deprecatedPrefix
		log.Infof("using deprecated oss kusion state file %s", prefix)
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

func (s *OssState) usingDeprecatedStateFilePrefix(query *states.StateQuery) (string, error) {
	var prefix string
	if query.Tenant != "" {
		prefix = query.Tenant + "/" + query.Project + "/" + query.Stack + "/" + deprecatedKusionStateFile
	} else {
		prefix = query.Project + "/" + query.Stack + "/" + deprecatedKusionStateFile
	}

	objects, err := s.bucket.ListObjects(oss.Delimiter("/"), oss.Prefix(prefix))
	if err != nil {
		return "", err
	}
	if len(objects.Objects) == 0 {
		return "", nil
	}
	return prefix, nil
}
