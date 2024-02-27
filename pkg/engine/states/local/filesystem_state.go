package local

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
)

var _ states.StateStorage = &FileSystemState{}

type FileSystemState struct {
	// state Path is in the same dir where command line is invoked
	Path string
}

func NewFileSystemState() states.StateStorage {
	return &FileSystemState{}
}

const (
	deprecatedKusionStateFile = "kusion_state.json" // deprecated default kusion state file
	KusionStateFileFile       = "kusion_state.yaml"
)

func (f *FileSystemState) GetLatestState(query *states.StateQuery) (*states.State, error) {
	filePath := f.Path
	// if the file of specified path does not exist, use deprecated kusion state file.
	if deprecatedPath := f.usingDeprecatedKusionStateFilePath(); deprecatedPath != "" {
		filePath = deprecatedPath
		log.Infof("use deprecated kusion state file %s", filePath)
	}

	// create a new state file if no file exists
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, fs.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if len(yamlFile) != 0 {
		state := &states.State{}
		// JSON is a subset of YAML.
		// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
		// Go JSON library doesn't try to pick the right number type (int, float,
		// etc.) when unmarshalling to interface{}, it just picks float64 universally.
		// go-yaml does the right thing.
		err = yaml.Unmarshal(yamlFile, state)
		if err != nil {
			return nil, err
		}
		return state, nil
	} else {
		log.Infof("file %s is empty. Skip unmarshal", filePath)
		return nil, nil
	}
}

func (f *FileSystemState) Apply(state *states.State) error {
	now := time.Now()

	// don't change createTime in the state
	oldState, err := f.GetLatestState(nil)
	if err != nil {
		return err
	}

	if oldState == nil || oldState.CreateTime.IsZero() {
		state.CreateTime = now
	} else {
		state.CreateTime = oldState.CreateTime
	}

	state.ModifiedTime = now
	yamlByte, err := yaml.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(f.Path, yamlByte, fs.ModePerm)
}

func (f *FileSystemState) Delete(id string) error {
	log.Infof("Delete state file:%s", f.Path)
	err := os.Remove(f.Path)
	if err != nil {
		return err
	}
	// if deprecated kusion state file exists, also delete
	if f.deprecatedKusionStateFileExist() {
		deprecatedPath := filepath.Join(filepath.Dir(f.Path), deprecatedKusionStateFile)
		if err = os.Remove(deprecatedPath); err != nil {
			return err
		}
		log.Infof("delete deprecated state file %s", deprecatedPath)
	}
	return nil
}

func (f *FileSystemState) usingDeprecatedKusionStateFilePath() string {
	_, err := os.Stat(f.Path)
	if os.IsNotExist(err) {
		dir := filepath.Dir(f.Path)
		deprecatedPath := filepath.Join(dir, deprecatedKusionStateFile)
		if _, err = os.Stat(deprecatedPath); err == nil {
			return deprecatedPath
		}
	}
	return ""
}

func (f *FileSystemState) deprecatedKusionStateFileExist() bool {
	dir := filepath.Dir(f.Path)
	_, err := os.Stat(filepath.Join(dir, deprecatedKusionStateFile))
	return err == nil
}
