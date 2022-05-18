package apply

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/compile"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/manifest"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/states"
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
		mockOperationApply(operation.Success)

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
		) (*manifest.Manifest, *pterm.SpinnerPrinter, error) {
			sp := pterm.DefaultSpinner.
				WithSequence("⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ ").
				WithDelay(time.Millisecond * 100)

			sp, _ = sp.Start(fmt.Sprintf("Compiling in stack %s...", stack.Name))

			return &manifest.Manifest{Resources: []states.ResourceState{sa1, sa2, sa3}}, sp, nil
		})
}

func Test_preview(t *testing.T) {
	t.Run("preview success", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewApplyOptions()
		_, err := preview(o, &manifest.Manifest{Resources: []states.ResourceState{sa1, sa2, sa3}}, project, stack)
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

func (f *fakerRuntime) Apply(ctx context.Context, priorState, planState *states.ResourceState) (*states.ResourceState, status.Status) {
	return planState, nil
}

func (f *fakerRuntime) Read(ctx context.Context, resourceState *states.ResourceState) (*states.ResourceState, status.Status) {
	return resourceState, nil
}

func (f *fakerRuntime) Delete(ctx context.Context, resourceState *states.ResourceState) status.Status {
	return nil
}

func (f *fakerRuntime) Watch(ctx context.Context, resourceState *states.ResourceState) (*states.ResourceState, status.Status) {
	return resourceState, nil
}

func mockOperationPreview() {
	monkey.Patch((*operation.Operation).Preview,
		func(*operation.Operation, *operation.PreviewRequest, operation.Type) (rsp *operation.PreviewResponse, s status.Status) {
			return &operation.PreviewResponse{
				Order: &operation.ChangeOrder{
					StepKeys: []string{sa1.ID, sa2.ID, sa3.ID},
					ChangeSteps: map[string]*operation.ChangeStep{
						sa1.ID: {
							ID:     sa1.ID,
							Action: operation.Create,
							Old:    nil,
							New:    &sa1,
						},
						sa2.ID: {
							ID:     sa2.ID,
							Action: operation.UnChange,
							Old:    &sa2,
							New:    &sa2,
						},
						sa3.ID: {
							ID:     sa3.ID,
							Action: operation.Undefined,
							Old:    &sa3,
							New:    &sa1,
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

func newSA(name string) states.ResourceState {
	return states.ResourceState{
		ID:   engine.BuildIDForKubernetes(apiVersion, kind, namespace, name),
		Mode: "managed",
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
	t.Run("dry run", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockNewKubernetesRuntime()

		planResources := &manifest.Manifest{Resources: []states.ResourceState{sa1}}
		order := &operation.ChangeOrder{
			StepKeys: []string{sa1.ID},
			ChangeSteps: map[string]*operation.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: operation.Create,
					Old:    nil,
					New:    sa1,
				},
			},
		}
		changes := operation.NewChanges(project, stack, order)
		o := NewApplyOptions()
		o.DryRun = true
		err := apply(o, planResources, changes)
		assert.Nil(t, err)
	})
	t.Run("apply success", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockNewKubernetesRuntime()
		mockOperationApply(operation.Success)

		o := NewApplyOptions()
		planResources := &manifest.Manifest{Resources: []states.ResourceState{sa1, sa2}}
		order := &operation.ChangeOrder{
			StepKeys: []string{sa1.ID, sa2.ID},
			ChangeSteps: map[string]*operation.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: operation.Create,
					Old:    nil,
					New:    &sa1,
				},
				sa2.ID: {
					ID:     sa2.ID,
					Action: operation.UnChange,
					Old:    &sa2,
					New:    &sa2,
				},
			},
		}
		changes := operation.NewChanges(project, stack, order)

		err := apply(o, planResources, changes)
		assert.Nil(t, err)
	})
	t.Run("apply failed", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockNewKubernetesRuntime()
		mockOperationApply(operation.Failed)

		o := NewApplyOptions()
		planResources := &manifest.Manifest{Resources: []states.ResourceState{sa1}}
		order := &operation.ChangeOrder{
			StepKeys: []string{sa1.ID},
			ChangeSteps: map[string]*operation.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: operation.Create,
					Old:    nil,
					New:    &sa1,
				},
			},
		}
		changes := operation.NewChanges(project, stack, order)

		err := apply(o, planResources, changes)
		assert.NotNil(t, err)
	})
}

func mockOperationApply(res operation.OpResult) {
	monkey.Patch((*operation.Operation).Apply,
		func(o *operation.Operation, request *operation.ApplyRequest) (*operation.ApplyResponse, status.Status) {
			var err error
			if res == operation.Failed {
				err = errors.New("mock error")
			}
			for _, r := range request.Manifest.Resources {
				// ing -> $res
				o.MsgCh <- operation.Message{
					ResourceID: r.ResourceKey(),
					OpResult:   "",
					OpErr:      nil,
				}
				o.MsgCh <- operation.Message{
					ResourceID: r.ResourceKey(),
					OpResult:   res,
					OpErr:      err,
				}
			}
			close(o.MsgCh)
			if res == operation.Failed {
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
