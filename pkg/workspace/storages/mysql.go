package storages

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

// MysqlStorage is an implementation of workspace.Storage which uses mysql as storage.
type MysqlStorage struct {
	db *gorm.DB
}

// NewMysqlStorage news mysql workspace storage and init default workspace.
func NewMysqlStorage(db *gorm.DB) (*MysqlStorage, error) {
	s := &MysqlStorage{db: db}

	return s, s.initDefaultWorkspaceIf()
}

func (s *MysqlStorage) Get(name string) (*v1.Workspace, error) {
	w, err := getWorkspaceFromMysql(s.db, name)
	if err != nil {
		return nil, fmt.Errorf("get workspace from mysql database failed: %w", err)
	}
	if w == nil {
		return nil, ErrWorkspaceNotExist
	}

	ws, err := convertFromMysqlDO(w)
	if err != nil {
		return nil, err
	}
	ws.Name = name
	return ws, nil
}

func (s *MysqlStorage) Create(ws *v1.Workspace) error {
	exist, err := checkWorkspaceExistenceInMysql(s.db, ws.Name)
	if err != nil {
		return err
	}
	if exist {
		return ErrWorkspaceAlreadyExist
	}

	w, err := convertToMysqlDO(ws)
	if err != nil {
		return err
	}
	return createWorkspaceInMysql(s.db, w)
}

func (s *MysqlStorage) Update(ws *v1.Workspace) error {
	if ws.Name == "" {
		name, err := getCurrentWorkspaceNameFromMysql(s.db)
		if err != nil {
			return err
		}
		ws.Name = name
	}
	exist, err := checkWorkspaceExistenceInMysql(s.db, ws.Name)
	if err != nil {
		return err
	}
	if !exist {
		return ErrWorkspaceNotExist
	}

	w, err := convertToMysqlDO(ws)
	if err != nil {
		return err
	}
	return updateWorkspaceInMysql(s.db, w)
}

func (s *MysqlStorage) Delete(name string) error {
	if name == "" {
		var err error
		name, err = getCurrentWorkspaceNameFromMysql(s.db)
		if err != nil {
			return err
		}
	}
	return deleteWorkspaceInMysql(s.db, name)
}

func (s *MysqlStorage) GetNames() ([]string, error) {
	return getWorkspaceNamesFromMysql(s.db)
}

func (s *MysqlStorage) GetCurrent() (string, error) {
	return getCurrentWorkspaceNameFromMysql(s.db)
}

func (s *MysqlStorage) SetCurrent(name string) error {
	exist, err := checkWorkspaceExistenceInMysql(s.db, name)
	if err != nil {
		return err
	}
	if !exist {
		return ErrWorkspaceNotExist
	}

	return alterCurrentWorkspaceInMysql(s.db, name)
}

func (s *MysqlStorage) initDefaultWorkspaceIf() error {
	exist, err := checkWorkspaceExistenceInMysql(s.db, DefaultWorkspace)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	w := &WorkspaceMysqlDO{Name: DefaultWorkspace}
	currentName, err := getCurrentWorkspaceNameFromMysql(s.db)
	if err != nil {
		return err
	}
	if currentName == "" {
		isCurrent := true
		w.IsCurrent = &isCurrent
	}

	return createWorkspaceInMysql(s.db, w)
}

// WorkspaceMysqlDO is the data object stored in the mysql db.
type WorkspaceMysqlDO struct {
	Name      string
	Content   string
	IsCurrent *bool
}

func (s WorkspaceMysqlDO) TableName() string {
	return workspaceTable
}

func getWorkspaceFromMysql(db *gorm.DB, name string) (*WorkspaceMysqlDO, error) {
	q := &WorkspaceMysqlDO{Name: name}
	if name == "" {
		isCurrent := true
		q.IsCurrent = &isCurrent
	}
	w := &WorkspaceMysqlDO{}
	result := db.Where(q).First(w)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return w, result.Error
}

func getCurrentWorkspaceNameFromMysql(db *gorm.DB) (string, error) {
	isCurrent := true
	q := &WorkspaceMysqlDO{IsCurrent: &isCurrent}
	w := &WorkspaceMysqlDO{}
	result := db.Select("name").Where(q).First(w)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return w.Name, result.Error
}

func getWorkspaceNamesFromMysql(db *gorm.DB) ([]string, error) {
	var wList []*WorkspaceMysqlDO
	result := db.Select("name").Find(wList)
	if result.Error != nil {
		return nil, result.Error
	}
	names := make([]string, len(wList))
	for i, w := range wList {
		names[i] = w.Name
	}
	return names, nil
}

func checkWorkspaceExistenceInMysql(db *gorm.DB, name string) (bool, error) {
	q := &WorkspaceMysqlDO{Name: name}
	w := &WorkspaceMysqlDO{}
	result := db.Select("name").Where(q).First(w)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return result.Error == nil, result.Error
}

func createWorkspaceInMysql(db *gorm.DB, w *WorkspaceMysqlDO) error {
	return db.Create(w).Error
}

func updateWorkspaceInMysql(db *gorm.DB, w *WorkspaceMysqlDO) error {
	q := &WorkspaceMysqlDO{Name: w.Name}
	return db.Where(q).Updates(w).Error
}

func deleteWorkspaceInMysql(db *gorm.DB, name string) error {
	q := &WorkspaceMysqlDO{Name: name}
	return db.Where(q).Delete(&WorkspaceMysqlDO{}).Error
}

func alterCurrentWorkspaceInMysql(db *gorm.DB, name string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// set the current workspace now to not current
		isCurrent := true
		q := &WorkspaceMysqlDO{IsCurrent: &isCurrent}
		notCurrent := false
		w := &WorkspaceMysqlDO{IsCurrent: &notCurrent}
		result := tx.Where(q).Updates(w)
		if result.Error != nil {
			return result.Error
		}

		// set current of the specified workspace
		q = &WorkspaceMysqlDO{Name: name}
		w = &WorkspaceMysqlDO{IsCurrent: &isCurrent}
		result = tx.Where(q).Updates(w)
		return result.Error
	})
}

func convertToMysqlDO(ws *v1.Workspace) (*WorkspaceMysqlDO, error) {
	content, err := yaml.Marshal(ws)
	if err != nil {
		return nil, fmt.Errorf("yaml marshal workspace failed: %w", err)
	}
	return &WorkspaceMysqlDO{
		Name:    ws.Name,
		Content: string(content),
	}, nil
}

func convertFromMysqlDO(w *WorkspaceMysqlDO) (*v1.Workspace, error) {
	ws := &v1.Workspace{}
	if err := yaml.Unmarshal([]byte(w.Content), ws); err != nil {
		return nil, fmt.Errorf("yaml unmarshal workspace failed: %w", err)
	}
	return ws, nil
}
