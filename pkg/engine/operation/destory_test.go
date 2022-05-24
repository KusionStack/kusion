//go:build !arm64
// +build !arm64

package operation

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/states"
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
	o := &Operation{
		OperationType: Destroy,
		StateStorage:  &states.FileSystemState{Path: filepath.Join("test_data", states.KusionState)},
		Runtime:       &runtime.KubernetesRuntime{},
	}
	r := &DestroyRequest{
		Request{
			Tenant:   tenant,
			Stack:    stack,
			Project:  project,
			Operator: operator,
			Manifest: mf,
		},
	}

	t.Run("destroy success", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch((*ResourceNode).Execute, func(rn *ResourceNode, operation *Operation) status.Status {
			return nil
		})
		o.MsgCh = make(chan Message, 1)
		go readMsgCh(o.MsgCh)
		st := o.Destroy(r)
		assert.Nil(t, st)
	})

	t.Run("destroy failed", func(t *testing.T) {
		defer monkey.UnpatchAll()
		monkey.Patch((*ResourceNode).Execute, func(rn *ResourceNode, operation *Operation) status.Status {
			return status.NewErrorStatus(errors.New("mock error"))
		})

		o.MsgCh = make(chan Message, 1)
		go readMsgCh(o.MsgCh)
		st := o.Destroy(r)
		assert.True(t, status.IsErr(st))
	})
}

func readMsgCh(ch chan Message) {
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
