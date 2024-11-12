package api

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/liu-hm19/pterm"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/infra/util/semaphore"
	"kusionstack.io/kusion/pkg/log"
	logutil "kusionstack.io/kusion/pkg/server/util/logging"
)

// The Apply function will apply the resources changes
// through the execution Kusion Engine, and will save
// the state to specified storage.
//
// You can customize the runtime of engine and the state
// storage through `runtime` and `storage` parameters.
func Apply(
	ctx context.Context,
	o *APIOptions,
	storage release.Storage,
	rel *apiv1.Release,
	gph *apiv1.Graph,
	changes *models.Changes,
	out io.Writer,
) (*apiv1.Release, error) {
	logger := logutil.GetLogger(ctx)
	// construct the apply operation
	ac := &operation.ApplyOperation{
		Operation: models.Operation{
			Stack:          changes.Stack(),
			ReleaseStorage: storage,
			MsgCh:          make(chan models.Message),
			IgnoreFields:   o.IgnoreFields,
			Sem:            semaphore.New(int64(o.MaxConcurrent)),
		},
	}

	// line summary
	var ls lineSummary

	// progress bar, print dag walk detail
	progressbar, err := pterm.DefaultProgressbar.
		WithMaxWidth(0). // Set to 0, the terminal width will be used
		WithTotal(len(changes.StepKeys)).
		WithWriter(out).
		WithRemoveWhenDone().
		Start()
	if err != nil {
		return nil, err
	}
	// wait msgCh close
	var wg sync.WaitGroup
	// receive msg and print detail
	go func() {
		defer func() {
			if p := recover(); p != nil {
				log.Errorf("failed to receive msg and print detail as %v", p)
			}
		}()
		wg.Add(1)

		for {
			select {
			case msg, ok := <-ac.MsgCh:
				if !ok {
					wg.Done()
					return
				}
				changeStep := changes.Get(msg.ResourceID)

				switch msg.OpResult {
				case models.Success, models.Skip:
					var title string
					if changeStep.Action == models.UnChanged {
						title = fmt.Sprintf("%s %s, %s",
							changeStep.Action.String(),
							pterm.Bold.Sprint(changeStep.ID),
							strings.ToLower(string(models.Skip)),
						)
					} else {
						title = fmt.Sprintf("%s %s %s",
							changeStep.Action.String(),
							pterm.Bold.Sprint(changeStep.ID),
							strings.ToLower(string(msg.OpResult)),
						)
					}
					pterm.Success.WithWriter(out).Println(title)
					progressbar.UpdateTitle(title)
					progressbar.Increment()
					ls.Count(changeStep.Action)
				case models.Failed:
					title := fmt.Sprintf("%s %s %s",
						changeStep.Action.String(),
						pterm.Bold.Sprint(changeStep.ID),
						strings.ToLower(string(msg.OpResult)),
					)
					pterm.Error.WithWriter(out).Printf("%s\n", title)
				default:
					title := fmt.Sprintf("%s %s %s",
						changeStep.Action.Ing(),
						pterm.Bold.Sprint(changeStep.ID),
						strings.ToLower(string(msg.OpResult)),
					)
					progressbar.UpdateTitle(title)
				}
			}
		}
	}()

	var upRel *apiv1.Release
	if o.DryRun {
		for _, r := range rel.Spec.Resources {
			ac.MsgCh <- models.Message{
				ResourceID: r.ResourceKey(),
				OpResult:   models.Success,
				OpErr:      nil,
			}
		}
		close(ac.MsgCh)
	} else {
		// parse cluster in arguments
		rsp, st := ac.Apply(&operation.ApplyRequest{
			Request: models.Request{
				Project: changes.Project(),
				Stack:   changes.Stack(),
			},
			Release: rel,
			Graph:   gph,
		})
		if v1.IsErr(st) {
			return nil, fmt.Errorf("apply failed, status:\n%v", st)
		}
		upRel = rsp.Release
	}

	// wait for msgCh closed
	wg.Wait()
	// print summary
	logger.Info(fmt.Sprintf("Apply complete! Resources: %d created, %d updated, %d deleted.", ls.created, ls.updated, ls.deleted))
	return upRel, nil
}

// Watch function will observe the changes of each resource
// by the execution engine.
//
// Example:
//
//	o := NewApplyOptions()
//	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
//	if err != nil {
//	    return err
//	}
//
//	Watch(o, kubernetesRuntime, planResources, changes, os.Stdout)
//	if err != nil {
//	    return err
//	}
func Watch(
	o *APIOptions,
	planResources *apiv1.Spec,
	changes *models.Changes,
) error {
	if o.DryRun {
		fmt.Println("NOTE: Watch doesn't work in DryRun mode")
		return nil
	}

	// filter out unchanged resources
	toBeWatched := apiv1.Resources{}
	for _, res := range planResources.Resources {
		if changes.ChangeOrder.ChangeSteps[res.ResourceKey()].Action != models.UnChanged {
			toBeWatched = append(toBeWatched, res)
		}
	}

	// watch operation
	wo := &operation.WatchOperation{}
	if err := wo.Watch(&operation.WatchRequest{
		Request: models.Request{
			Project: changes.Project(),
			Stack:   changes.Stack(),
		},
		Spec: &apiv1.Spec{Resources: toBeWatched},
	}); err != nil {
		return err
	}

	fmt.Println("Watch Finish! All resources have been reconciled.")
	return nil
}

type lineSummary struct {
	created, updated, deleted int
}

func (ls *lineSummary) Count(op models.ActionType) {
	switch op {
	case models.Create:
		ls.created++
	case models.Update:
		ls.updated++
	case models.Delete:
		ls.deleted++
	}
}
