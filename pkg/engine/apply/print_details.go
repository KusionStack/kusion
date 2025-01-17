package apply

import (
	"errors"
	"fmt"

	"github.com/liu-hm19/pterm"
	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	applystate "kusionstack.io/kusion/pkg/engine/apply/state"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/resource/graph"
	"kusionstack.io/kusion/pkg/util/pretty"
)

// PrintApplyDetails function will receive the messages of the apply operation and print the details.
func PrintApplyDetails(
	state *applystate.State,
	msgChan chan models.Message,
	applyResult chan error,
	changesWriterMap map[string]*pterm.SpinnerPrinter,
	progressbar *pterm.ProgressbarPrinter,
	changes *models.Changes,
) {
	var err error
	defer func() {
		applyResult <- err
		close(applyResult)
	}()
	defer cmdutil.RecoverErr(&err)

	for {
		select {
		// Get operation results from the message channel.
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			changeStep := changes.Get(msg.ResourceID)

			// Update the progressbar and spinner printer according to the operation result.
			switch msg.OpResult {
			case models.Success, models.Skip:
				var title string
				if changeStep.Action == models.UnChanged {
					title = fmt.Sprintf("Skipped %s", pterm.Bold.Sprint(changeStep.ID))
					changesWriterMap[msg.ResourceID].Success(title)
				} else {
					if state.Watch && !state.DryRun {
						title = fmt.Sprintf("%s %s",
							changeStep.Action.Ing(),
							pterm.Bold.Sprint(changeStep.ID),
						)
						changesWriterMap[msg.ResourceID].UpdateText(title)
					} else {
						changesWriterMap[msg.ResourceID].Success(fmt.Sprintf("Succeeded %s", pterm.Bold.Sprint(msg.ResourceID)))
					}
				}

				// Update resource status
				if !state.DryRun && changeStep.Action != models.UnChanged {
					gphResource := graph.FindGraphResourceByID(state.Gph.Resources, msg.ResourceID)
					if gphResource != nil {
						// Delete resource from the graph if it's deleted during apply
						if changeStep.Action == models.Delete {
							graph.RemoveResource(state.Gph, gphResource)
						} else {
							gphResource.Status = apiv1.ApplySucceed
						}
					}
				}

				progressbar.Increment()
				state.Ls.Count(changeStep.Action)
			case models.Failed:
				title := fmt.Sprintf("Failed %s", pterm.Bold.Sprint(changeStep.ID))
				changesWriterMap[msg.ResourceID].Fail(title)
				errStr := pretty.ErrorT.Sprintf("apply %s failed as: %s\n", msg.ResourceID, msg.OpErr.Error())
				err = errors.Join(err, errors.New(errStr))
				if !state.DryRun {
					// Update resource status, in case anything like update fail happened
					gphResource := graph.FindGraphResourceByID(state.Gph.Resources, msg.ResourceID)
					if gphResource != nil {
						gphResource.Status = apiv1.ApplyFail
					}
				}
			default:
				title := fmt.Sprintf("%s %s",
					changeStep.Action.Ing(),
					pterm.Bold.Sprint(changeStep.ID),
				)
				changesWriterMap[msg.ResourceID].UpdateText(title)
			}
		}
	}
}
