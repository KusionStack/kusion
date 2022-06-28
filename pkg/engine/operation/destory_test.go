//go:build !arm64
// +build !arm64

package operation

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"kusionstack.io/kusion/pkg/engine/states/local"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"

	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/operation/types"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/status"
)

func TestOperation_Destroy(t *testing.T) {
	var (
		tenant   = "tenant_name"
		stack    = "dev"
		project  = "project_name"
		operator = "foo_user"
	)
	resourceState := models.Resource{
		ID: "id1",

		Attributes: map[string]interface{}{
			"foo": "bar",
		},
		DependsOn: nil,
	}
	mf := &models.Spec{Resources: []models.Resource{resourceState}}
	o := &DestroyOperation{
		opsmodels.Operation{
			OperationType: types.Destroy,
			StateStorage:  &local.FileSystemState{Path: filepath.Join("test_data", local.KusionState)},
			Runtime:       &runtime.KubernetesRuntime{},
		},
	}
	r := &DestroyRequest{
		opsmodels.Request{
			Tenant:   tenant,
			Stack:    stack,
			Project:  project,
			Operator: operator,
			Spec:     mf,
		},
	}

	t.Run("destroy success", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch((*graph.ResourceNode).Execute, func(rn *graph.ResourceNode, operation *opsmodels.Operation) status.Status {
			return nil
		})
		o.MsgCh = make(chan opsmodels.Message, 1)
		go readMsgCh(o.MsgCh)
		st := o.Destroy(r)
		assert.Nil(t, st)
	})

	t.Run("destroy failed", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch((*graph.ResourceNode).Execute, func(rn *graph.ResourceNode, operation *opsmodels.Operation) status.Status {
			return status.NewErrorStatus(errors.New("mock error"))
		})

		o.MsgCh = make(chan opsmodels.Message, 1)
		go readMsgCh(o.MsgCh)
		st := o.Destroy(r)
		assert.True(t, status.IsErr(st))
	})
}

func readMsgCh(ch chan opsmodels.Message) {
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Printf("msg: %+v\n", msg)
		}
	}
}
