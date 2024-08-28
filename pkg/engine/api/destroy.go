package api

import (
	"fmt"
	"strings"
	"sync"

	"github.com/liu-hm19/pterm"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/release"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/infra/util/semaphore"
	"kusionstack.io/kusion/pkg/log"
)

func DestroyPreview(
	o *APIOptions,
	planResources *apiv1.Spec,
	priorResources *apiv1.State,
	proj *apiv1.Project,
	stack *apiv1.Stack,
	storage release.Storage,
) (*models.Changes, error) {
	log.Info("Start compute preview changes ...")

	// check and install terraform executable binary for
	// resources with the type of Terraform.
	tfInstaller := terraform.CLIInstaller{
		Intent: planResources,
	}
	if err := tfInstaller.CheckAndInstall(); err != nil {
		return nil, err
	}

	pc := &operation.PreviewOperation{
		Operation: models.Operation{
			OperationType:  models.DestroyPreview,
			Stack:          stack,
			ReleaseStorage: storage,
			ChangeOrder:    &models.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*models.ChangeStep{}},
			Sem:            semaphore.New(int64(o.MaxConcurrent)),
		},
	}

	log.Info("Start call pc.Preview() ...")

	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: models.Request{
			Project: proj,
			Stack:   stack,
		},
		Spec:  planResources,
		State: priorResources,
	})
	if v1.IsErr(s) {
		return nil, fmt.Errorf("preview failed, status: %v", s)
	}

	return models.NewChanges(proj, stack, rsp.Order), nil
}

func Destroy(
	o *APIOptions,
	rel *apiv1.Release,
	changes *models.Changes,
	storage release.Storage,
) (*apiv1.Release, error) {
	do := &operation.DestroyOperation{
		Operation: models.Operation{
			Stack:          changes.Stack(),
			ReleaseStorage: storage,
			MsgCh:          make(chan models.Message),
			Sem:            semaphore.New(int64(o.MaxConcurrent)),
		},
	}

	// line summary
	var deleted int

	// progress bar, print dag walk detail
	progressbar, err := pterm.DefaultProgressbar.
		WithTotal(len(changes.StepKeys)).
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
			case msg, ok := <-do.MsgCh:
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
					pterm.Success.Println(title)
					progressbar.UpdateTitle(title)
					progressbar.Increment()
					deleted++
				case models.Failed:
					title := fmt.Sprintf("%s %s %s",
						changeStep.Action.String(),
						pterm.Bold.Sprint(changeStep.ID),
						strings.ToLower(string(msg.OpResult)),
					)
					pterm.Error.Printf("%s\n", title)
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

	rsp, st := do.Destroy(&operation.DestroyRequest{
		Request: models.Request{
			Project: changes.Project(),
			Stack:   changes.Stack(),
		},
		Release: rel,
	})
	if v1.IsErr(st) {
		return nil, fmt.Errorf("destroy failed, status: %v", st)
	}
	upRel := rsp.Release

	// wait for msgCh closed
	wg.Wait()
	// print summary
	pterm.Println()
	pterm.Printf("Destroy complete! Resources: %d deleted.\n", deleted)
	return upRel, nil
}
