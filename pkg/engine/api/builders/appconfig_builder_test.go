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

package builders

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
	"kcl-lang.io/kpm/pkg/api"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/modules"
)

func TestBuild(t *testing.T) {
	p, s := buildMockProjectAndStack()
	appName, app := buildMockApp()
	acg := &AppsConfigBuilder{
		Apps: map[string]v1.AppConfiguration{
			appName: *app,
		},
		Workspace: buildMockWorkspace(),
	}

	callMock := mockey.Mock(modules.CallGenerators).Return(nil).Build()
	defer func() {
		callMock.UnPatch()
	}()

	cwd, _ := os.Getwd()
	pkgPath := filepath.Join(cwd, "testdata")
	kclPkg, err := api.GetKclPackage(pkgPath)
	assert.NoError(t, err)

	intent, err := acg.Build(kclPkg, p, s)
	assert.NoError(t, err)
	assert.NotNil(t, intent)
}

func buildMockApp() (string, *v1.AppConfiguration) {
	return "app1", &v1.AppConfiguration{
		Workload: map[string]interface{}{
			"type": "Deployment",
			"ports": []map[string]any{
				{
					"port":     80,
					"protocol": "TCP",
				},
			},
		},
	}
}

func buildMockWorkspace() *v1.Workspace {
	return &v1.Workspace{
		Name: "test",
		Context: map[string]any{
			"Kubernetes": map[string]string{
				"Config": "/etc/kubeconfig.yaml",
			},
		},
	}
}

func buildMockProjectAndStack() (*v1.Project, *v1.Stack) {
	p := &v1.Project{
		Name: "test-project",
	}

	s := &v1.Stack{
		Name: "test-project",
	}

	return p, s
}
