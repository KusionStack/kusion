// Copyright 2024 KusionStack Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package meta

import (
	"github.com/spf13/cobra"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend"
	"kusionstack.io/kusion/pkg/project"
	"kusionstack.io/kusion/pkg/util/i18n"
)

// MetaFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// MetaOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type MetaFlags struct {
	Workspace *string

	Backend *string

	WorkDir *string
}

// MetaOptions are the meta-options that are available on all or most commands.
type MetaOptions struct {
	// RefProject references the project for this CLI invocation.
	RefProject *v1.Project

	// RefStack referenced the stack for this CLI invocation.
	RefStack *v1.Stack

	// RefWorkspace referenced the workspace for this CLI invocation.
	RefWorkspace *v1.Workspace

	// StorageBackend referenced the target storage backend for this CLI invocation.
	StorageBackend backend.Backend
}

// NewMetaFlags provides default flags and values for use in other commands.
func NewMetaFlags() *MetaFlags {
	workspace := ""
	backendType := ""
	workDir := ""

	return &MetaFlags{
		Workspace: &workspace,
		Backend:   &backendType,
		WorkDir:   &workDir,
	}
}

// AddFlags registers flags for a cli.
func (f *MetaFlags) AddFlags(cmd *cobra.Command) {
	if f.Workspace != nil {
		cmd.Flags().StringVarP(f.Workspace, "workspace", "", *f.Workspace, i18n.T("The name of target workspace to operate in."))
	}
	if f.Backend != nil {
		cmd.Flags().StringVarP(f.Backend, "backend", "", *f.Backend, i18n.T("The backend to use, supports 'local', 'oss' and 's3'."))
	}
	if f.WorkDir != nil {
		cmd.Flags().StringVarP(f.WorkDir, "workdir", "w", *f.WorkDir, i18n.T("The work directory to run Kusion CLI."))
	}
}

// ToOptions converts MetaFlags to MetaOptions.
func (f *MetaFlags) ToOptions() (*MetaOptions, error) {
	opts := &MetaOptions{}

	// Parse project and currentStack of work directory
	var refProject *v1.Project
	var refStack *v1.Stack
	var err error
	if f.WorkDir != nil && *f.WorkDir != "" {
		refProject, refStack, err = project.DetectProjectAndStackFrom(*f.WorkDir)
	} else {
		refProject, refStack, err = project.DetectProjectAndStacks()
	}

	if err != nil {
		return nil, err
	}

	opts.RefProject = refProject
	opts.RefStack = refStack

	storageBackend, err := f.ParseBackend()
	if err != nil {
		return nil, err
	}
	opts.StorageBackend = storageBackend

	// Get current workspace from backend
	workspace, err := f.ParseWorkspace(storageBackend)
	if err != nil {
		return nil, err
	}
	opts.RefWorkspace = workspace

	return opts, nil
}

func (f *MetaFlags) ParseWorkspace(storageBackend backend.Backend) (*v1.Workspace, error) {
	if f.Workspace != nil && storageBackend != nil {
		workspaceStorage, err := storageBackend.WorkspaceStorage()
		if err != nil {
			return nil, err
		}
		refWorkspace, err := workspaceStorage.Get(*f.Workspace)
		if err != nil {
			return nil, err
		}
		return refWorkspace, nil
	}
	return nil, nil
}

func (f *MetaFlags) ParseBackend() (backend.Backend, error) {
	var storageBackend backend.Backend
	var err error
	if f.Backend != nil {
		storageBackend, err = backend.NewBackend(*f.Backend)
		if err != nil {
			return nil, err
		}
	}
	return storageBackend, nil
}
