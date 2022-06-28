package remote

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/engine/states"
)

var ErrOSSNoExist = errors.New("oss: key not exist")

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

// ConfigSchema returns a description of the expected configuration
// structure for the receiving backend.
func (s *OssState) ConfigSchema() cty.Type {
	return cty.Type{}
}

// Configure uses the provided configuration to set configuration fields
// within the OssState backend.
func (s *OssState) Configure(obj cty.Value) error {
	return nil
}

func (s *OssState) Apply(state *states.State) error {
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

func (s *OssState) GetLatestState(query *states.StateQuery) (*states.State, error) {
	prefix := query.Tenant + "/" + query.Project + "/" + query.Stack
	objects, err := s.bucket.ListObjects(oss.Delimiter("/"), oss.Prefix(prefix))
	if err != nil {
		return nil, err
	}

	var result *oss.ObjectProperties
	for i := 0; i < len(objects.Objects); i++ {
		if result == nil || result.LastModified.UnixNano() < objects.Objects[i].LastModified.UnixNano() {
			result = &objects.Objects[i]
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
	state := &states.State{}
	// JSON is a subset of YAML. Please check FileSystemState.GetLatestState for detail explanation
	err = yaml.Unmarshal(data, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}
