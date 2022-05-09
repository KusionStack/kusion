package states

import (
	"bytes"
	"encoding/json"
	"errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
)

var ErrOSSNoExist = errors.New("oss: key not exist")

type OssState struct {
	bucket *oss.Bucket
}

func NewOSSState(endPoint, accessKeyId, accessKeySecret, bucketName string) (*OssState, error) {
	var ossClient *oss.Client
	var err error
	ossClient, err = oss.New(endPoint, accessKeyId, accessKeySecret)
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

func (s *OssState) Apply(state *State) error {
	u, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	jsonByte, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	prefix := state.Tenant + "/" + state.Project + "/" + state.Stack
	err = s.bucket.PutObject(prefix+u.String(), bytes.NewReader(jsonByte))
	if err != nil {
		return err
	}
	return nil
}

func (s *OssState) Delete(id string) error {
	panic("implement me")
}

func (s *OssState) GetLatestState(query *StateQuery) (*State, error) {
	prefix := query.Tenant + "/" + query.Project + "/" + query.Stack
	objects, err := s.bucket.ListObjects(oss.Delimiter("/"), oss.Prefix(prefix))
	if err != nil {
		return nil, err
	}
	var result *oss.ObjectProperties
	for _, obj := range objects.Objects {
		if result == nil || result.LastModified.UnixNano() < obj.LastModified.UnixNano() {
			result = &obj
		}
	}
	if result == nil {
		return nil, ErrOSSNoExist
	}

	body, err := s.bucket.GetObject(result.Key)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	state := &State{}
	// JSON is a subset of YAML. Please check FileSystemState.GetLatestState for detail explanation
	err = yaml.Unmarshal(data, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}
