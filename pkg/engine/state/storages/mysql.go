package storages

import (
	"errors"

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
	stateDO, err := getStateFromMysql(s.db, s.project, s.stack, s.workspace)
	if err != nil {
		return nil, err
	}
	if stateDO == nil {
		return nil, nil
	}
	return convertFromMysqlDO(stateDO)
}

func (s *MysqlStorage) Apply(state *v1.State) error {
	exist, err := checkStateExistenceInMysql(s.db, s.project, s.stack, s.workspace)
	if err != nil {
		return err
	}

	stateDO, err := convertToMysqlDO(state)
	if err != nil {
		return err
	}
	if exist {
		return updateStateInMysql(s.db, stateDO)
	} else {
		return createStateInMysql(s.db, stateDO)
	}
}

// StateMysqlDO is the data object stored in the mysql db.
type StateMysqlDO struct {
	Project   string
	Stack     string
	Workspace string
	Content   string
}

func (s StateMysqlDO) TableName() string {
	return stateTable
}

func getStateFromMysql(db *gorm.DB, project, stack, workspace string) (*StateMysqlDO, error) {
	q := &StateMysqlDO{
		Project:   project,
		Stack:     stack,
		Workspace: workspace,
	}
	s := &StateMysqlDO{}
	result := db.Where(q).First(s)
	// if no record, return nil state and nil error
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return s, result.Error
}

func checkStateExistenceInMysql(db *gorm.DB, project, stack, workspace string) (bool, error) {
	q := &StateMysqlDO{
		Project:   project,
		Stack:     stack,
		Workspace: workspace,
	}
	s := &StateMysqlDO{}
	result := db.Select("project").Where(q).First(s)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return result.Error == nil, result.Error
}

func createStateInMysql(db *gorm.DB, s *StateMysqlDO) error {
	return db.Create(s).Error
}

func updateStateInMysql(db *gorm.DB, s *StateMysqlDO) error {
	q := &StateMysqlDO{
		Project:   s.Project,
		Stack:     s.Stack,
		Workspace: s.Workspace,
	}
	return db.Where(q).Updates(s).Error
}

func convertToMysqlDO(state *v1.State) (*StateMysqlDO, error) {
	content, err := yaml.Marshal(state)
	if err != nil {
		return nil, err
	}
	return &StateMysqlDO{
		Project:   state.Project,
		Stack:     state.Stack,
		Workspace: state.Workspace,
		Content:   string(content),
	}, nil
}

func convertFromMysqlDO(s *StateMysqlDO) (*v1.State, error) {
	state := &v1.State{}
	if err := yaml.Unmarshal([]byte(s.Content), state); err != nil {
		return nil, err
	}
	return state, nil
}
