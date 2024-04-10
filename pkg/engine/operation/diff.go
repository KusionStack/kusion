package operation

import (
	"github.com/pkg/errors"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/state"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/diff"
	"kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/third_party/dyff"
)

type Diff struct {
	StateStorage state.Storage
}

type DiffRequest struct {
	models.Request
}

func (d *Diff) Diff(request *DiffRequest) (string, error) {
	log.Infof("invoke Diff")

	defer func() {
		if err := recover(); err != nil {
			log.Error("Diff panic:%v", err)
		}
	}()

	util.CheckNotNil(request, "request is nil")
	util.CheckNotNil(request.Intent, "resource is nil")

	// Get plan state resources
	plan := request.Intent

	// Get the state resources
	priorState, err := d.StateStorage.Get()
	if err != nil {
		return "", errors.Wrap(err, "GetLatestState failed")
	}
	if priorState == nil {
		log.Infof("can't find states by request: %v.", json.MustMarshal2String(request))
	}
	// Get diff result
	return DiffWithRequestResourceAndState(plan, priorState)
}

func DiffWithRequestResourceAndState(plan *v1.Spec, priorState *v1.DeprecatedState) (string, error) {
	planString := json.MustMarshal2String(plan.Resources)
	var report *dyff.Report
	var err error
	if priorState == nil {
		report, err = diff.ToReport("", planString)
	} else {
		latestResources := priorState.Resources
		priorString := json.MustMarshal2String(latestResources)
		report, err = diff.ToReport(priorString, priorString)
	}
	if err != nil {
		return "", err
	}
	return diff.ToHumanString(diff.NewHumanReport(report))
}
