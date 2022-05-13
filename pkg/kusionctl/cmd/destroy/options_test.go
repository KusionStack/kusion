package destroy

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

func TestDestroyOptions_Run(t *testing.T) {
	t.Run("Detail is true", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockCompileWithSpinner()
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewDestroyOptions()
		o.Detail = true
		err := o.Run()
		assert.Nil(t, err)
	})

	t.Run("prompt no", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockCompileWithSpinner()
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewDestroyOptions()
		mockPromptOutput("no")
		err := o.Run()
		assert.Nil(t, err)
	})

	t.Run("prompt yes", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockDetectProjectAndStack()
		mockCompileWithSpinner()
		mockNewKubernetesRuntime()
		mockOperationPreview()
		mockOperationDestroy(operation.Success)

		o := NewDestroyOptions()
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

			return &manifest.Manifest{Resources: []states.ResourceState{sa1}}, sp, nil
		})
}

func Test_preview(t *testing.T) {
	t.Run("preview success", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockNewKubernetesRuntime()
		mockOperationPreview()

		o := NewDestroyOptions()
		_, err := o.preview(&manifest.Manifest{Resources: []states.ResourceState{sa1}}, project, stack)
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
				ChangeSteps: map[string]*operation.ChangeStep{
					sa1.ID: {
						ID:     sa1.ID,
						Action: operation.Delete,
						Old:    &sa1,
						New:    nil,
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

func Test_destroy(t *testing.T) {
	t.Run("destroy success", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockNewKubernetesRuntime()
		mockOperationDestroy(operation.Success)

		o := NewDestroyOptions()
		planResources := &manifest.Manifest{Resources: []states.ResourceState{sa2}}
		changeSteps := map[string]*operation.ChangeStep{
			sa1.ID: {
				ID:     sa1.ID,
				Action: operation.Delete,
				Old:    &sa1,
				New:    nil,
			},
			sa2.ID: {
				ID:     sa2.ID,
				Action: operation.UnChange,
				Old:    &sa2,
				New:    &sa2,
			},
		}
		changes := operation.NewChanges(project, stack, changeSteps)

		err := o.destroy(planResources, changes)
		assert.Nil(t, err)
	})
	t.Run("destroy failed", func(t *testing.T) {
		defer monkey.UnpatchAll()
		mockNewKubernetesRuntime()
		mockOperationDestroy(operation.Failed)

		o := NewDestroyOptions()
		planResources := &manifest.Manifest{Resources: []states.ResourceState{sa1}}
		changeSteps := map[string]*operation.ChangeStep{
			sa1.ID: {
				ID:     sa1.ID,
				Action: operation.Delete,
				Old:    &sa1,
				New:    nil,
			},
		}
		changes := operation.NewChanges(project, stack, changeSteps)

		err := o.destroy(planResources, changes)
		assert.NotNil(t, err)
	})
}

func mockOperationDestroy(res operation.OpResult) {
	monkey.Patch((*operation.Operation).Destroy,
		func(o *operation.Operation, request *operation.DestroyRequest) status.Status {
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
				return status.NewErrorStatus(err)
			}
			return nil
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
		_, err := prompt()
		assert.NotNil(t, err)
	})

	t.Run("prompt yes", func(t *testing.T) {
		mockPromptOutput("yes")
		_, err := prompt()
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
