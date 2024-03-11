package storages

import (
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

// MysqlStorage is an implementation of state.Storage which uses mysql as storage.
type MysqlStorage struct {
	db                        *gorm.DB
	project, stack, workspace string
}

func NewMysqlStorage(db *gorm.DB, project, stack, workspace string) *MysqlStorage {
	return &MysqlStorage{
		db:        db,
		project:   project,
		stack:     stack,
		workspace: workspace,
	}
}

func (s *MysqlStorage) Get() (*v1.State, error) {
	stateDO, err := getState(s.db, s.project, s.stack, s.workspace)
	if err != nil {
		return nil, err
	}
	if stateDO == nil {
		return nil, nil
	}
	return convertFromDO(stateDO)
}

func (s *MysqlStorage) Apply(state *v1.State) error {
	exist, err := isStateExist(s.db, s.project, s.stack, s.workspace)
	if err != nil {
		return err
	}

	stateDO, err := convertToDO(state)
	if err != nil {
		return err
	}
	if exist {
		return updateState(s.db, stateDO)
	} else {
		return createState(s.db, stateDO)
	}
}

// State is the data object stored in the mysql db.
type State struct {
	Project   string
	Stack     string
	Workspace string
	Content   string
}

func (s State) TableName() string {
	return stateTable
}

func getState(db *gorm.DB, project, stack, workspace string) (*State, error) {
	q := &State{
		Project:   project,
		Stack:     stack,
		Workspace: workspace,
	}
	s := &State{}
	result := db.Where(q).First(s)
	// if no record, return nil
	if *s == (State{}) {
		s = nil
	}
	return s, result.Error
}

func isStateExist(db *gorm.DB, project, stack, workspace string) (bool, error) {
	s, err := getState(db, project, stack, workspace)
	if err != nil {
		return false, err
	}
	return s != nil, err
}

func createState(db *gorm.DB, s *State) error {
	return db.Create(s).Error
}

func updateState(db *gorm.DB, s *State) error {
	q := &State{
		Project:   s.Project,
		Stack:     s.Stack,
		Workspace: s.Workspace,
	}
	return db.Where(q).Updates(s).Error
}

func convertToDO(state *v1.State) (*State, error) {
	content, err := yaml.Marshal(state)
	if err != nil {
		return nil, err
	}
	return &State{
		Project:   state.Project,
		Stack:     state.Stack,
		Workspace: state.Workspace,
		Content:   string(content),
	}, nil
}

func convertFromDO(s *State) (*v1.State, error) {
	state := &v1.State{}
	if err := yaml.Unmarshal([]byte(s.Content), state); err != nil {
		return nil, err
	}
	return state, nil
}
