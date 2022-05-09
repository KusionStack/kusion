package operation

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/hashicorp/terraform/dag"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/manifest"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/status"
)

func TestResourceNode_Execute(t *testing.T) {
	type fields struct {
		BaseNode BaseNode
		Action   ActionType
		state    *states.ResourceState
	}
	type args struct {
		operation Operation
	}

	const Jack = "jack"
	const Pony = "pony"
	const Eric = "eric"
	mf := &manifest.Manifest{Resources: []states.ResourceState{
		{
			ID:   Pony,
			Mode: states.Managed,
			Attributes: map[string]interface{}{
				"c": "d",
			},
			DependsOn: []string{Jack},
		},
		{
			ID:   Eric,
			Mode: states.Managed,
			Attributes: map[string]interface{}{
				"a": ImplicitRefPrefix + "jack.a.b",
			},
			DependsOn: []string{Pony},
		},
		{
			ID:   Jack,
			Mode: states.Managed,
			Attributes: map[string]interface{}{
				"a": map[string]interface{}{
					"b": "c",
				},
			},
			DependsOn: nil,
		},
	}}

	priorStateResourceIndex := map[string]*states.ResourceState{}
	for i, resource := range mf.Resources {
		priorStateResourceIndex[resource.ResourceKey()] = &mf.Resources[i]
	}

	newResourceState := &states.ResourceState{
		ID:   Eric,
		Mode: states.Managed,
		Attributes: map[string]interface{}{
			"a": ImplicitRefPrefix + "jack.a.b",
		},
		DependsOn: []string{Pony},
	}

	illegalResourceState := &states.ResourceState{
		ID:   Eric,
		Mode: states.Managed,
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
				BaseNode: BaseNode{ID: Jack},
				Action:   Update,
				state:    newResourceState,
			},
			args: args{operation: Operation{
				OperationType:           Apply,
				StateStorage:            states.NewFileSystemState(),
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				MsgCh:                   make(chan Message),
				resultState:             states.NewState(),
				lock:                    &sync.Mutex{},
				Runtime:                 &runtime.KubernetesRuntime{},
			}},
			want: nil,
		},
		{
			name: "delete",
			fields: fields{
				BaseNode: BaseNode{ID: Jack},
				Action:   Delete,
				state:    newResourceState,
			},
			args: args{operation: Operation{
				OperationType:           Apply,
				StateStorage:            states.NewFileSystemState(),
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				MsgCh:                   make(chan Message),
				resultState:             states.NewState(),
				lock:                    &sync.Mutex{},
				Runtime:                 &runtime.KubernetesRuntime{},
			}},
			want: nil,
		},
		{
			name: "illegalRef",
			fields: fields{
				BaseNode: BaseNode{ID: Jack},
				Action:   Update,
				state:    illegalResourceState,
			},
			args: args{operation: Operation{
				OperationType:           Apply,
				StateStorage:            states.NewFileSystemState(),
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				MsgCh:                   make(chan Message),
				resultState:             states.NewState(),
				lock:                    &sync.Mutex{},
				Runtime:                 &runtime.KubernetesRuntime{},
			}},
			want: status.NewErrorStatusWithMsg(status.IllegalManifest, "can't find specified value in resource:jack by ref:jack.notExist"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rn := &ResourceNode{
				BaseNode: tt.fields.BaseNode,
				Action:   tt.fields.Action,
				state:    tt.fields.state,
			}
			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.operation.Runtime), "Apply",
				func(k *runtime.KubernetesRuntime, ctx context.Context, priorState, planState *states.ResourceState) (*states.ResourceState, status.Status) {
					mockState := *newResourceState
					mockState.Attributes["a"] = "c"
					return &mockState, nil
				})
			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.operation.Runtime), "Delete",
				func(k *runtime.KubernetesRuntime, ctx context.Context, priorState *states.ResourceState) status.Status {
					return nil
				})
			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.operation.StateStorage), "Apply",
				func(f *states.FileSystemState, state *states.State) error {
					return nil
				})
			defer monkey.UnpatchAll()

			assert.Equalf(t, tt.want, rn.Execute(tt.args.operation), "Execute(%v)", tt.args.operation)
		})
	}
}
