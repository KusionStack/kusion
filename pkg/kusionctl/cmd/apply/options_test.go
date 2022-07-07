//go:build !arm64
// +build !arm64

package apply

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"kusionstack.io/kusion/pkg/engine/states/local"

	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"

	"kusionstack.io/kusion/pkg/engine/operation/types"

	"bou.ke/monkey"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/compile"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
)

func TestApplyOptions_Run(t *testing.T) {
	t.Run("Detail is true", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockCompileWithSpinner()
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewApplyOptions()
		o.Detail = true
		o.NoStyle = true
		err := o.Run()
		assert.Nil(t, err)
	})

	t.Run("OnlyPreview is true", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockCompileWithSpinner()
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewApplyOptions()
		o.OnlyPreview = true

		mockPromptOutput("yes")
		err := o.Run()
		assert.Nil(t, err)

		mockPromptOutput("no")
		err = o.Run()
		assert.Nil(t, err)
	})

	t.Run("DryRun is true", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockCompileWithSpinner()
		mockNewKubernetesRuntime()
		mockOperationPreview()
		mockOperationApply(opsmodels.Success)

		o := NewApplyOptions()
		o.DryRun = true
		mockPromptOutput("yes")
		err := o.Run()
		assert.Nil(t, err)
	})
}

var (
	project = &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name:   "testdata",
			Tenant: "admin",
		},
	}
	stack = &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{
			Name: "dev",
		},
	}
)

func mockDetectProjectAndStack() {
	monkey.Patch(projectstack.DetectProjectAndStack, func(stackDir string) (*projectstack.Project, *projectstack.Stack, error) {
		project.Path = stackDir
		stack.Path = stackDir
		return project, stack, nil
	})
}

func mockCompileWithSpinner() {
	monkey.Patch(compile.CompileWithSpinner,
		func(workDir string, filenames, settings, arguments, overrides []string, stack *projectstack.Stack,
		) (*models.Spec, *pterm.SpinnerPrinter, error) {
			sp := pterm.DefaultSpinner.
				WithSequence("⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ ").
				WithDelay(time.Millisecond * 100)

			sp, _ = sp.Start(fmt.Sprintf("Compiling in stack %s...", stack.Name))

			return &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, sp, nil
		})
}

func Test_preview(t *testing.T) {
	stateStorage := &local.FileSystemState{Path: filepath.Join("", local.KusionState)}
	t.Run("preview success", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockOperationPreview()

		o := NewApplyOptions()
		_, err := Preview(o, &fakerRuntime{}, stateStorage, &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, project, stack, os.Stdout)
		assert.Nil(t, err)
	})
}

func mockNewKubernetesRuntime() {
	monkey.Patch(runtime.NewKubernetesRuntime, func() (runtime.Runtime, error) {
		return &fakerRuntime{}, nil
	})
}

var _ runtime.Runtime = (*fakerRuntime)(nil)

type fakerRuntime struct{}

func (f *fakerRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakerRuntime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
	if request.Resource.ResourceKey() == "fake-id" {
		return &runtime.ReadResponse{
			Resource: nil,
			Status:   nil,
		}
	}
	return &runtime.ReadResponse{
		Resource: request.Resource,
		Status:   nil,
	}
}

func (f *fakerRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fakerRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

func mockOperationPreview() {
	monkey.Patch((*operation.PreviewOperation).Preview,
		func(*operation.PreviewOperation, *operation.PreviewRequest) (rsp *operation.PreviewResponse, s status.Status) {
			return &operation.PreviewResponse{
				Order: &opsmodels.ChangeOrder{
					StepKeys: []string{sa1.ID, sa2.ID, sa3.ID},
					ChangeSteps: map[string]*opsmodels.ChangeStep{
						sa1.ID: {
							ID:       sa1.ID,
							Action:   types.Create,
							Original: nil,
							Modified: &sa1,
						},
						sa2.ID: {
							ID:       sa2.ID,
							Action:   types.UnChange,
							Original: &sa2,
							Modified: &sa2,
						},
						sa3.ID: {
							ID:       sa3.ID,
							Action:   types.Undefined,
							Original: &sa3,
							Modified: &sa1,
						},
					},
				},
			}, nil
		},
	)
}

const (
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"
)

var (
	sa1 = newSA("sa1")
	sa2 = newSA("sa2")
	sa3 = newSA("sa3")
)

func newSA(name string) models.Resource {
	return models.Resource{
		ID: engine.BuildIDForKubernetes(apiVersion, kind, namespace, name),

		Attributes: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
}

func Test_apply(t *testing.T) {
	stateStorage := &local.FileSystemState{Path: filepath.Join("", local.KusionState)}
	t.Run("dry run", func(t *testing.T) {
		defer monkey.UnpatchAll()

		planResources := &models.Spec{Resources: []models.Resource{sa1}}
		order := &opsmodels.ChangeOrder{
			StepKeys: []string{sa1.ID},
			ChangeSteps: map[string]*opsmodels.ChangeStep{
				sa1.ID: {
					ID:       sa1.ID,
					Action:   types.Create,
					Original: nil,
					Modified: sa1,
				},
			},
		}
		changes := opsmodels.NewChanges(project, stack, order)
		o := NewApplyOptions()
		o.DryRun = true
		err := Apply(o, &fakerRuntime{}, stateStorage, planResources, changes, os.Stdout)
		assert.Nil(t, err)
	})
	t.Run("apply success", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockOperationApply(opsmodels.Success)

		o := NewApplyOptions()
		planResources := &models.Spec{Resources: []models.Resource{sa1, sa2}}
		order := &opsmodels.ChangeOrder{
			StepKeys: []string{sa1.ID, sa2.ID},
			ChangeSteps: map[string]*opsmodels.ChangeStep{
				sa1.ID: {
					ID:       sa1.ID,
					Action:   types.Create,
					Original: nil,
					Modified: &sa1,
				},
				sa2.ID: {
					ID:       sa2.ID,
					Action:   types.UnChange,
					Original: &sa2,
					Modified: &sa2,
				},
			},
		}
		changes := opsmodels.NewChanges(project, stack, order)

		err := Apply(o, &fakerRuntime{}, stateStorage, planResources, changes, os.Stdout)
		assert.Nil(t, err)
	})
	t.Run("apply failed", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockOperationApply(opsmodels.Failed)

		o := NewApplyOptions()
		planResources := &models.Spec{Resources: []models.Resource{sa1}}
		order := &opsmodels.ChangeOrder{
			StepKeys: []string{sa1.ID},
			ChangeSteps: map[string]*opsmodels.ChangeStep{
				sa1.ID: {
					ID:       sa1.ID,
					Action:   types.Create,
					Original: nil,
					Modified: &sa1,
				},
			},
		}
		changes := opsmodels.NewChanges(project, stack, order)

		err := Apply(o, &fakerRuntime{}, stateStorage, planResources, changes, os.Stdout)
		assert.NotNil(t, err)
	})
}

func mockOperationApply(res opsmodels.OpResult) {
	monkey.Patch((*operation.ApplyOperation).Apply,
		func(o *operation.ApplyOperation, request *operation.ApplyRequest) (*operation.ApplyResponse, status.Status) {
			var err error
			if res == opsmodels.Failed {
				err = errors.New("mock error")
			}
			for _, r := range request.Spec.Resources {
				// ing -> $res
				o.MsgCh <- opsmodels.Message{
					ResourceID: r.ResourceKey(),
					OpResult:   "",
					OpErr:      nil,
				}
				o.MsgCh <- opsmodels.Message{
					ResourceID: r.ResourceKey(),
					OpResult:   res,
					OpErr:      err,
				}
			}
			close(o.MsgCh)
			if res == opsmodels.Failed {
				return nil, status.NewErrorStatus(err)
			}
			return &operation.ApplyResponse{}, nil
		})
}

func Test_prompt(t *testing.T) {
	t.Run("prompt error", func(t *testing.T) {
		monkey.Patch(
			survey.AskOne,
			func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
				return errors.New("mock error")
			},
		)
		_, err := prompt(false)
		assert.NotNil(t, err)
	})

	t.Run("prompt yes", func(t *testing.T) {
		mockPromptOutput("yes")
		_, err := prompt(false)
		assert.Nil(t, err)
	})
}

func mockPromptOutput(res string) {
	monkey.Patch(
		survey.AskOne,
		func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			reflect.ValueOf(response).Elem().Set(reflect.ValueOf(res))
			return nil
		},
	)
}
