package operation

import (
	"github.com/pkg/errors"

	"kusionstack.io/kusion/pkg/engine/models"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util"
	"kusionstack.io/kusion/pkg/util/diff"
	jsonutil "kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/third_party/dyff"
)

type Diff struct {
	StateStorage states.StateStorage
}

type DiffRequest struct {
	opsmodels.Request
}

func (d *Diff) Diff(request *DiffRequest) (string, error) {
	log.Infof("invoke Diff")

	defer func() {
		if err := recover(); err != nil {
			log.Error("Diff panic:%v", err)
		}
	}()

	util.CheckNotNil(request, "request is nil")
	util.CheckNotNil(request.Spec, "resource is nil")

	// Get plan state resources
	plan := request.Spec

	// Get the latest state resources
	latestState, err := d.StateStorage.GetLatestState(
		&states.StateQuery{
			Tenant:  request.Tenant,
			Stack:   request.Stack.Name,
			Project: request.Project.Name,
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "GetLatestState failed")
	}
	if latestState == nil {
		log.Infof("can't find states by request: %v.", jsonutil.MustMarshal2String(request))
	}
	// Get diff result
	return DiffWithRequestResourceAndState(plan, latestState)
}

func DiffWithRequestResourceAndState(plan *models.Spec, latest *states.State) (string, error) {
	planString := jsonutil.MustMarshal2String(plan.Resources)
	var report *dyff.Report
	var err error
	if latest == nil {
		report, err = diff.ToReport("", planString)
	} else {
		latestResources := latest.Resources
		priorString := jsonutil.MustMarshal2String(latestResources)
		report, err = diff.ToReport(priorString, priorString)
	}
	if err != nil {
		return "", err
	}
	return diff.ToHumanString(diff.NewHumanReport(report))
}
