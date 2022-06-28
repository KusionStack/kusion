//go:build !arm64
// +build !arm64

package graph

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"kusionstack.io/kusion/pkg/engine/states/local"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"

	"kusionstack.io/kusion/pkg/engine/operation/types"

	"bou.ke/monkey"
	"github.com/hashicorp/terraform/dag"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/status"
)

func TestResourceNode_Execute(t *testing.T) {
	type fields struct {
		BaseNode baseNode
		Action   types.ActionType
		state    *models.Resource
	}
	type args struct {
		operation opsmodels.Operation
	}

	const Jack = "jack"
	const Pony = "pony"
	const Eric = "eric"
	mf := &models.Spec{Resources: []models.Resource{
		{
			ID: Pony,

			Attributes: map[string]interface{}{
				"c": "d",
			},
			DependsOn: []string{Jack},
		},
		{
			ID: Eric,

			Attributes: map[string]interface{}{
				"a": ImplicitRefPrefix + "jack.a.b",
			},
			DependsOn: []string{Pony},
		},
		{
			ID: Jack,

			Attributes: map[string]interface{}{
				"a": map[string]interface{}{
					"b": "c",
				},
			},
			DependsOn: nil,
		},
	}}

	priorStateResourceIndex := map[string]*models.Resource{}
	for i, resource := range mf.Resources {
		priorStateResourceIndex[resource.ResourceKey()] = &mf.Resources[i]
	}

	newResourceState := &models.Resource{
		ID: Eric,

		Attributes: map[string]interface{}{
			"a": ImplicitRefPrefix + "jack.a.b",
		},
		DependsOn: []string{Pony},
	}

	illegalResourceState := &models.Resource{
		ID: Eric,

		Attributes: map[string]interface{}{
			"a": ImplicitRefPrefix + "jack.notExist",
		},
		DependsOn: []string{Pony},
	}

	graph := &dag.AcyclicGraph{}
	graph.Add(&RootNode{})

	tests := []struct {
		name   string
		fields fields
		args   args
		want   status.Status
	}{
		{
			name: "update",
			fields: fields{
				BaseNode: baseNode{ID: Jack},
				Action:   types.Update,
				state:    newResourceState,
			},
			args: args{operation: opsmodels.Operation{
				OperationType:           types.Apply,
				StateStorage:            local.NewFileSystemState(),
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				MsgCh:                   make(chan opsmodels.Message),
				ResultState:             states.NewState(),
				Lock:                    &sync.Mutex{},
				Runtime:                 &runtime.KubernetesRuntime{},
			}},
			want: nil,
		},
		{
			name: "delete",
			fields: fields{
				BaseNode: baseNode{ID: Jack},
				Action:   types.Delete,
				state:    newResourceState,
			},
			args: args{operation: opsmodels.Operation{
				OperationType:           types.Apply,
				StateStorage:            local.NewFileSystemState(),
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				MsgCh:                   make(chan opsmodels.Message),
				ResultState:             states.NewState(),
				Lock:                    &sync.Mutex{},
				Runtime:                 &runtime.KubernetesRuntime{},
			}},
			want: nil,
		},
		{
			name: "illegalRef",
			fields: fields{
				BaseNode: baseNode{ID: Jack},
				Action:   types.Update,
				state:    illegalResourceState,
			},
			args: args{operation: opsmodels.Operation{
				OperationType:           types.Apply,
				StateStorage:            local.NewFileSystemState(),
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				MsgCh:                   make(chan opsmodels.Message),
				ResultState:             states.NewState(),
				Lock:                    &sync.Mutex{},
				Runtime:                 &runtime.KubernetesRuntime{},
			}},
			want: status.NewErrorStatusWithMsg(status.IllegalManifest, "can't find specified value in resource:jack by ref:jack.notExist"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rn := &ResourceNode{
				baseNode: &tt.fields.BaseNode,
				Action:   tt.fields.Action,
				state:    tt.fields.state,
			}
			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.operation.Runtime), "Apply",
				func(k *runtime.KubernetesRuntime, ctx context.Context, priorState, planState *models.Resource) (*models.Resource, status.Status) {
					mockState := *newResourceState
					mockState.Attributes["a"] = "c"
					return &mockState, nil
				})
			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.operation.Runtime), "Delete",
				func(k *runtime.KubernetesRuntime, ctx context.Context, priorState *models.Resource) status.Status {
					return nil
				})
			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.operation.Runtime), "Read",
				func(k *runtime.KubernetesRuntime, ctx context.Context, resourceState *models.Resource) (*models.Resource, status.Status) {
					return resourceState, nil
				})
			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.operation.StateStorage), "Apply",
				func(f *local.FileSystemState, state *states.State) error {
					return nil
				})
			defer monkey.UnpatchAll()

			assert.Equalf(t, tt.want, rn.Execute(&tt.args.operation), "Execute(%v)", tt.args.operation)
		})
	}
}
