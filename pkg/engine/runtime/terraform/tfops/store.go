package tfops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"kusionstack.io/kusion/pkg/engine/models"
)

// WorkspaceStore store Terraform workspaces.
type WorkspaceStore struct {
	Store map[string]*WorkSpace
	Fs    afero.Afero
}

// Create make Terraform workspace for given resources.
// convert kusion resource to hcl json and write to file
// and init in the workspace folder
func (ws *WorkspaceStore) Create(ctx context.Context, resource *models.Resource) error {
	w, ok := ws.Store[resource.ResourceKey()]
	if !ok {
		ws.Store[resource.ResourceKey()] = NewWorkSpace(resource, ws.Fs)
		w = ws.Store[resource.ResourceKey()]
	}
	// write hcl json to file
	if err := w.WriteHCL(); err != nil {
		return fmt.Errorf("write hcl error: %v", err)
	}

	// init workspace
	if err := w.InitWorkSpace(ctx); err != nil {
		return fmt.Errorf("init workspace error: %v", err)
	}
	return nil
}

// Remove delete workspace directory and delete its record from the store.
func (ws *WorkspaceStore) Remove(ctx context.Context, resource *models.Resource) error {
	w, ok := ws.Store[resource.ResourceKey()]
	if !ok {
		return nil
	}
	if err := w.fs.RemoveAll(w.dir); err != nil {
		return fmt.Errorf("remove workspace error %v", err)
	}
	delete(ws.Store, resource.ResourceKey())
	return nil
}

// GetWorkspaceStore find directory in the filesystem and store workspace
// return all terraform workspace record in the filesystem.
func GetWorkspaceStore(fs afero.Afero) (WorkspaceStore, error) {
	ws := WorkspaceStore{
		Store: make(map[string]*WorkSpace),
		Fs:    fs,
	}
	wd, _ := GetWorkSpaceDir()
	_, err := fs.Stat(wd)
	if err != nil {
		if os.IsNotExist(err) {
			if err = fs.MkdirAll(wd, os.ModePerm); err != nil {
				return ws, err
			}
		} else {
			return ws, err
		}
	}
	dirs, err := afero.ReadDir(fs, wd)
	if err != nil {
		return ws, err
	}
	for _, dir := range dirs {
		workspace := WorkSpace{
			fs:  fs,
			dir: filepath.Join(wd, dir.Name()),
		}
		ws.Store[dir.Name()] = &workspace
	}
	return ws, nil
}
