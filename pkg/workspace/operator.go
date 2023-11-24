package workspace

import (
	"path"

	"kusionstack.io/kusion/pkg/util/kfile"
)

const defaultStoragePath = ".workspace"

// Operator is used to handle the operations about workspace.
type Operator struct {
	// storagePath is the place to store the workspace config locally.
	storagePath string
}

func NewDefaultOperator() *Operator {
	kusionDataDir, _ := kfile.KusionDataFolder()
	return &Operator{
		storagePath: path.Join(kusionDataDir, defaultStoragePath),
	}
}

/*
todo: The following are the functions to get provided.
func (o *Operator) GetWorkspaces() ([]string, error) {}
func (o *Operator) GetWorkspaceConfig(name string) (*Config, error) {}
func (o *Operator) GetUnstructuredWorkspaceConfig(name string) (*unstructuredConfig, error) {}
func (o *Operator) SetWorkspaceConfig(name string, config *Config) error {}
*/
