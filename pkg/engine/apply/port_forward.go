package apply

import (
	"fmt"

	applystate "kusionstack.io/kusion/pkg/engine/apply/state"
	"kusionstack.io/kusion/pkg/engine/operation"
)

// PortForward function will forward the specified port from local to the project Kubernetes Service.
//
// Example:
//
// o := newApplyOptions()
// spec, err := generate.GenerateSpecWithSpinner(o.RefProject, o.RefStack, o.RefWorkspace, nil, o.NoStyle)
//
//	if err != nil {
//		 return err
//	}
//
// err = PortForward(o, spec)
//
//	if err != nil {
//	  return err
//	}
func PortForward(
	state *applystate.State,
) error {
	if state.DryRun {
		fmt.Println("NOTE: Portforward doesn't work in DryRun mode")
		return nil
	}

	state.PortForwardReady, state.PortForwardStop = make(chan struct{}, 1), make(chan struct{}, 1)

	// portforward operation
	wo := &operation.PortForwardOperation{}
	if err := wo.PortForward(&operation.PortForwardRequest{
		Spec: state.TargetRel.Spec,
		Port: state.PortForward,
	}, state.PortForwardStop, state.PortForwardReady); err != nil {
		return err
	}

	fmt.Println("Portforward has been completed!")
	return nil
}
