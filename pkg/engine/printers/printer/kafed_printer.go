package printer

import (
	"fmt"

	"kusionstack.io/kube-api/apps/v1alpha1"
)

func AddCollaSetHandlers(h PrintHandler) {
	_ = h.TableHandler(printCollaSet)
}

func printCollaSet(obj *v1alpha1.CollaSet) (string, bool) {
	desired := obj.Spec.Replicas
	current := obj.Status.Replicas
	updated := obj.Status.UpdatedReplicas
	updatedReady := obj.Status.UpdatedReadyReplicas
	updateAvailable := obj.Status.UpdatedAvailableReplicas
	return fmt.Sprintf("Desired: %d, Current: %d, Updated: %d, UpdatedReady: %d, UpdatedAvailable: %d",
		*desired, current, updated, updatedReady, updateAvailable), *desired == updateAvailable
}
