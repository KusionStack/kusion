package workspace

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	yaml "gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/util/kfile"
)

const (
	defaultRelativeStoragePath = "workspaces"
	suffixYAML                 = ".yaml"
)

var (
	ErrEmptyStoragePath = errors.New("empty storage path")
	ErrUnexpectedDir    = errors.New("unexpected dir under storage path")
	ErrFileNotYAML      = errors.New("not yaml file under storage path")

	ErrEmptyWorkspace        = errors.New("empty workspace")
	ErrWorkspaceNotExist     = errors.New("workspace does not exist")
	ErrWorkspaceAlreadyExist = errors.New("workspace has already existed")
)

// CheckWorkspaceExistenceByDefaultOperator checks the workspace exists or not by default operator.
func CheckWorkspaceExistenceByDefaultOperator(name string) (bool, error) {
	operator, err := NewValidDefaultOperator()
	if err != nil {
		return false, err
	}
	return operator.WorkspaceExist(name), nil
}

// GetWorkspaceByDefaultOperator gets a workspace by default operator.
func GetWorkspaceByDefaultOperator(name string) (*v1.Workspace, error) {
	operator, err := NewValidDefaultOperator()
	if err != nil {
		return nil, err
	}
	return operator.GetWorkspace(name)
}

// GetWorkspaceNamesByDefaultOperator list all the workspace names by default operator.
func GetWorkspaceNamesByDefaultOperator() ([]string, error) {
	operator, err := NewValidDefaultOperator()
	if err != nil {
		return nil, err
	}
	return operator.GetWorkspaceNames()
}

// CreateWorkspaceByDefaultOperator creates a workspace by default operator.
func CreateWorkspaceByDefaultOperator(ws *v1.Workspace) error {
	operator, err := NewValidDefaultOperator()
	if err != nil {
		return err
	}
	return operator.CreateWorkspace(ws)
}

// UpdateWorkspaceByDefaultOperator updates a workspace by default operator.
func UpdateWorkspaceByDefaultOperator(ws *v1.Workspace) error {
	operator, err := NewValidDefaultOperator()
	if err != nil {
		return err
	}
	return operator.UpdateWorkspace(ws)
}

// DeleteWorkspaceByDefaultOperator deletes a workspace by default operator
func DeleteWorkspaceByDefaultOperator(name string) error {
	operator, err := NewValidDefaultOperator()
	if err != nil {
		return err
	}
	return operator.DeleteWorkspace(name)
}

// Operator is used to handle the CURD operations of workspace. Operator only supports local file
// system as backend for now.
type Operator struct {
	// storagePath is absolute path of the directory to store the workspace configs locally. The
	// storagePath should only include workspace configuration files.
	storagePath string
}

// NewOperator news an Operator with the specified storage path. If the directory of the storage
// path has not created, then create the directory.
func NewOperator(storagePath string) (*Operator, error) {
	_, err := os.Stat(storagePath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(storagePath, os.ModePerm); err != nil {
			return nil, fmt.Errorf("create directory %s failed, %w", storagePath, err)
		}
	}
	return &Operator{storagePath: storagePath}, nil
}

// NewDefaultOperator returns a default backend, whose storage path is the directory "workspace"
// under kfile.KusionDataFolder().
func NewDefaultOperator() (*Operator, error) {
	kusionDataDir, err := kfile.KusionDataFolder()
	if err != nil {
		return nil, fmt.Errorf("get kusion data folder failed, %w", err)
	}
	storagePath := path.Join(kusionDataDir, defaultRelativeStoragePath)
	return NewOperator(storagePath)
}

// NewValidDefaultOperator news a default operator and then do the validation work.
func NewValidDefaultOperator() (*Operator, error) {
	operator, err := NewDefaultOperator()
	if err != nil {
		return nil, err
	}
	if err = operator.Validate(); err != nil {
		return nil, err
	}
	return operator, nil
}

// Validate is used to validate the Operator is valid or not.
func (o *Operator) Validate() error {
	if o.storagePath == "" {
		return ErrEmptyStoragePath
	}

	files, err := o.getWorkspaceFiles()
	if err != nil {
		return err
	}
	// under storage path, directories or files without suffix ".yaml" are not allowed.
	for _, file := range files {
		if file.IsDir() {
			return fmt.Errorf("%w, dir path: %s", ErrUnexpectedDir, file.Name())
		}
		if !strings.HasSuffix(file.Name(), suffixYAML) {
			return fmt.Errorf("%w, file path: %s", ErrFileNotYAML, file.Name())
		}
	}
	return nil
}

// WorkspaceExist checks the workspace exists or not.
func (o *Operator) WorkspaceExist(name string) bool {
	_, err := os.Stat(o.workspaceFilePath(name))
	return !os.IsNotExist(err)
}

// GetWorkspaceNames gets all the workspace names.
func (o *Operator) GetWorkspaceNames() ([]string, error) {
	files, err := o.getWorkspaceFiles()
	if err != nil {
		return nil, err
	}

	workspaces := make([]string, len(files))
	for i, file := range files {
		workspaces[i] = strings.TrimSuffix(file.Name(), suffixYAML)
	}
	return workspaces, nil
}

// GetWorkspace gets the workspace by name. The validity of the returned workspace is not guaranteed.
func (o *Operator) GetWorkspace(name string) (*v1.Workspace, error) {
	if name == "" {
		return nil, ErrEmptyWorkspaceName
	}
	content, err := os.ReadFile(o.workspaceFilePath(name))
	if os.IsNotExist(err) {
		return nil, ErrWorkspaceNotExist
	} else if err != nil {
		return nil, fmt.Errorf("read workspace file failed: %w", err)
	}

	ws := &v1.Workspace{}
	if err = yaml.Unmarshal(content, ws); err != nil {
		return nil, fmt.Errorf("yaml unmarshal failed: %w", err)
	}
	ws.Name = name
	return ws, nil
}

// CreateWorkspace creates a workspace. The validation of workspace should be done before creating.
func (o *Operator) CreateWorkspace(ws *v1.Workspace) error {
	if ws == nil {
		return ErrEmptyWorkspace
	}
	exist := o.WorkspaceExist(ws.Name)
	if exist {
		return ErrWorkspaceAlreadyExist
	}

	return o.writeWorkspaceFile(ws)
}

// UpdateWorkspace updates a workspace.The validation of workspace should be done before updating.
func (o *Operator) UpdateWorkspace(ws *v1.Workspace) error {
	if ws == nil {
		return ErrEmptyWorkspace
	}
	exist := o.WorkspaceExist(ws.Name)
	if !exist {
		return ErrWorkspaceNotExist
	}

	return o.writeWorkspaceFile(ws)
}

// DeleteWorkspace deletes a workspace.
func (o *Operator) DeleteWorkspace(name string) error {
	if name == "" {
		return ErrEmptyWorkspaceName
	}
	exist := o.WorkspaceExist(name)
	if !exist {
		return ErrWorkspaceNotExist
	}

	if err := os.Remove(o.workspaceFilePath(name)); err != nil {
		return fmt.Errorf("remove workspace file failed: %w", err)
	}
	return nil
}

func (o *Operator) workspaceFilePath(name string) string {
	return path.Join(o.storagePath, name+suffixYAML)
}

func (o *Operator) getWorkspaceFiles() ([]os.DirEntry, error) {
	files, err := os.ReadDir(o.storagePath)
	if err != nil {
		return nil, fmt.Errorf("read files under storage path %s failed: %w", o.storagePath, err)
	}
	return files, nil
}

func (o *Operator) writeWorkspaceFile(ws *v1.Workspace) error {
	content, err := yaml.Marshal(ws)
	if err != nil {
		return fmt.Errorf("yaml marshal workspace failed: %w", err)
	}
	if err = os.WriteFile(o.workspaceFilePath(ws.Name), content, 0o640); err != nil {
		return fmt.Errorf("write workspace file failed: %w", err)
	}
	return nil
}
