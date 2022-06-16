package operation

import (
	"bytes"
	"fmt"
	"strings"

	"kusionstack.io/kusion/pkg/engine/operation/utils"

	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"

	"github.com/gonvenience/wrap"
	"github.com/pkg/errors"
	yamlv3 "gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/engine/models"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/diff"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util"
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
			Stack:   request.Stack,
			Project: request.Project,
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
	if latest == nil {
		return DiffReport("", planString, diff.OutputHuman)
	} else {
		latestResources := latest.Resources
		priorString := jsonutil.MustMarshal2String(latestResources)
		return DiffReport(priorString, planString, diff.OutputHuman)
	}
}

func DiffReport(prior, plan, mode string) (string, error) {
	from, err := utils.LoadFile(prior, "Last State")
	if err != nil {
		return "", err
	}
	to, err := utils.LoadFile(plan, "Request State")
	if err != nil {
		return "", err
	}

	report, err := dyff.CompareInputFiles(from, to, dyff.IgnoreOrderChanges(true))
	if err != nil {
		return "", wrap.Errorf(err, "failed to compare input files")
	}
	return buildReport(mode, report)
}

func buildReport(mode string, report dyff.Report) (string, error) {
	switch strings.ToLower(mode) {
	case diff.OutputHuman:
		return writeReport(report)
	case diff.OutputRaw:
		// output stdout/file
		reportMap := map[string]interface{}{
			"diffs": report.Diffs,
		}
		reportYAML, err := yamlv3.Marshal(reportMap)
		if err != nil {
			return "", wrap.Errorf(err, "failed to marshal report diffs")
		}
		return string(reportYAML), nil
	default:
		return "", fmt.Errorf("invalid output style `%s`", mode)
	}
}

// WriteReport writes a human-readable report to the provided writer
func writeReport(report dyff.Report) (string, error) {
	reportWriter := &dyff.HumanReport{
		Report:               report,
		MinorChangeThreshold: 0.1,
	}

	buffer := new(bytes.Buffer)
	if err := reportWriter.WriteReport(buffer); err != nil {
		return "", err
	}
	return buffer.String(), nil
}
