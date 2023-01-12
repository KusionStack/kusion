package graph

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/models"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/third_party/terraform/dag"
)

func TestResourceNode_Execute(t *testing.T) {
	type fields struct {
		BaseNode baseNode
		Action   opsmodels.ActionType
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
			ID:   Pony,
			Type: runtime.Kubernetes,
			Attributes: map[string]interface{}{
				"c": "d",
			},
			DependsOn: []string{Jack},
		},
		{
			ID:   Eric,
			Type: runtime.Kubernetes,
			Attributes: map[string]interface{}{
				"a": ImplicitRefPrefix + "jack.a.b",
			},
			DependsOn: []string{Pony},
		},
		{
			ID:   Jack,
			Type: runtime.Kubernetes,
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
		ID:   Eric,
		Type: runtime.Kubernetes,
		Attributes: map[string]interface{}{
			"a": ImplicitRefPrefix + "jack.a.b",
		},
		DependsOn: []string{Pony},
	}

	illegalResourceState := &models.Resource{
		ID:   Eric,
		Type: runtime.Kubernetes,
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
				Action:   opsmodels.Update,
				state:    newResourceState,
			},
			args: args{operation: opsmodels.Operation{
				OperationType:           opsmodels.Apply,
				StateStorage:            local.NewFileSystemState(),
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				IgnoreFields:            []string{"not_exist_field"},
				MsgCh:                   make(chan opsmodels.Message),
				ResultState:             states.NewState(),
				Lock:                    &sync.Mutex{},
				RuntimeMap:              map[models.Type]runtime.Runtime{runtime.Kubernetes: &runtime.KubernetesRuntime{}},
			}},
			want: nil,
		},
		{
			name: "delete",
			fields: fields{
				BaseNode: baseNode{ID: Jack},
				Action:   opsmodels.Delete,
				state:    newResourceState,
			},
			args: args{operation: opsmodels.Operation{
				OperationType:           opsmodels.Apply,
				StateStorage:            local.NewFileSystemState(),
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				MsgCh:                   make(chan opsmodels.Message),
				ResultState:             states.NewState(),
				Lock:                    &sync.Mutex{},
				RuntimeMap:              map[models.Type]runtime.Runtime{runtime.Kubernetes: &runtime.KubernetesRuntime{}},
			}},
			want: nil,
		},
		{
			name: "illegalRef",
			fields: fields{
				BaseNode: baseNode{ID: Jack},
				Action:   opsmodels.Update,
				state:    illegalResourceState,
			},
			args: args{operation: opsmodels.Operation{
				OperationType:           opsmodels.Apply,
				StateStorage:            local.NewFileSystemState(),
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				MsgCh:                   make(chan opsmodels.Message),
				ResultState:             states.NewState(),
				Lock:                    &sync.Mutex{},
				RuntimeMap:              map[models.Type]runtime.Runtime{runtime.Kubernetes: &runtime.KubernetesRuntime{}},
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
			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.operation.RuntimeMap[runtime.Kubernetes]), "Apply",
				func(k *runtime.KubernetesRuntime, ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
					mockState := *newResourceState
					mockState.Attributes["a"] = "c"
					return &runtime.ApplyResponse{
						Resource: &mockState,
					}
				})
			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.operation.RuntimeMap[runtime.Kubernetes]), "Delete",
				func(k *runtime.KubernetesRuntime, ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
					return &runtime.DeleteResponse{Status: nil}
				})
			monkey.PatchInstanceMethod(reflect.TypeOf(tt.args.operation.RuntimeMap[runtime.Kubernetes]), "Read",
				func(k *runtime.KubernetesRuntime, ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
					return &runtime.ReadResponse{Resource: request.PriorResource}
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
