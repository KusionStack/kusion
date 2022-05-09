package states

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/zclconf/go-cty/cty"
	"kusionstack.io/kusion/pkg/log"
)

func init() {
	AddToBackends("local", func() StateStorage {
		return NewFileSystemState()
	})
}

type FileSystemState struct {
	// state Path is in the same dir where command line is invoked
	Path string
}

func NewFileSystemState() StateStorage {
	return &FileSystemState{}
}

const KusionState = "kusion_state.json"

func (f *FileSystemState) ConfigSchema() cty.Type {
	config := map[string]cty.Type{
		"path": cty.String,
	}
	return cty.Object(config)
}

func (f *FileSystemState) Configure(obj cty.Value) error {
	var path cty.Value
	if path = obj.GetAttr("path"); !path.IsNull() && path.AsString() != "" {
		f.Path = path.AsString()
	} else {
		f.Path = KusionState
	}
	return nil
}

func (f *FileSystemState) GetLatestState(query *StateQuery) (*State, error) {
	// parse state
	file, err := os.OpenFile(f.Path, os.O_RDWR|os.O_CREATE, fs.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	jsonFile, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return nil, err
	}

	if len(jsonFile) != 0 {
		state := &State{}
		// JSON is a subset of YAML.
		// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
		// Go JSON library doesn't try to pick the right number type (int, float,
		// etc.) when unmarshalling to interface{}, it just picks float64 universally.
		// go-yaml does the right thing.
		err = yaml.Unmarshal(jsonFile, state)
		if err != nil {
			return nil, err
		}
		return state, nil
	} else {
		log.Infof("file %s is empty. Skip unmarshal json", f.Path)
		return nil, nil
	}
}

func (f *FileSystemState) Apply(state *State) error {
	now := time.Now()
	state.GmtCreate = now
	state.GmtModified = now
	jsonByte, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f.Path, jsonByte, fs.ModePerm)
}

func (f *FileSystemState) Delete(id string) error {
	log.Infof("Delete state file:%s", f.Path)
	err := os.Remove(f.Path)
	if err != nil {
		return err
	}
	return nil
}
