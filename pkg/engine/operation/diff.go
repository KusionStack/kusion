package operation

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gonvenience/wrap"
	"github.com/gonvenience/ytbx"
	"github.com/pkg/errors"
	yamlv3 "gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/engine/manifest"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/diff"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util"
	jsonUtil "kusionstack.io/kusion/pkg/util/json"
	"kusionstack.io/kusion/third_party/dyff"
)

type Diff struct {
	StateStorage states.StateStorage
}

type DiffRequest struct {
	Request
}

func (d *Diff) Diff(request *DiffRequest) (string, error) {
	log.Infof("invoke Diff")

	defer func() {
		if err := recover(); err != nil {
			log.Error("Diff panic:%v", err)
		}
	}()

	util.CheckNotNil(request, "request is nil")
	util.CheckNotNil(request.Manifest, "resource is nil")

	// Get plan state resources
	plan := request.Manifest
	// ignore id & privates
	// for _, resourceState := range plan {
	// 	for _, instance := range resourceState.Instances {
	// 		instance.Private = nil
	// 		instance.Attributes["id"] = ""
	// 	}
	// }

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
		log.Infof("can't find states by request: %v.", jsonUtil.MustMarshal2String(request))
	}
	// Get diff result
	return DiffWithRequestResourceAndState(plan, latestState)
}

func DiffWithRequestResourceAndState(plan *manifest.Manifest, latest *states.State) (string, error) {
	planString := jsonUtil.MustMarshal2String(plan.Resources)
	if latest == nil {
		return DiffReport("", planString, diff.OutputHuman)
	} else {
		latestResources := latest.Resources
		// ignore id & privates
		// TODO: use diff PathsToIgnoreAddition option
		// for i, resourceState := range latestResources {
		// 	for ii := range resourceState.Instances {
		// 		instance := &latestResources[i].Instances[ii]
		// 		instance.Private = nil
		// 		instance.Attributes["id"] = ""
		// 	}
		// }
		priorString := jsonUtil.MustMarshal2String(latestResources)
		return DiffReport(priorString, planString, diff.OutputHuman)
	}
}

func DiffReport(prior, plan, mode string) (string, error) {
	from, err := LoadFile(prior, "Last State")
	if err != nil {
		return "", err
	}
	to, err := LoadFile(plan, "Request State")
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

// LoadFile processes the provided input location to load it as one of the
// supported document formats, or plain text if nothing else works.
func LoadFile(yaml, location string) (ytbx.InputFile, error) {
	var (
		documents []*yamlv3.Node
		data      []byte
		err       error
	)

	data = []byte(yaml)
	if documents, err = ytbx.LoadDocuments(data); err != nil {
		return ytbx.InputFile{}, wrap.Errorf(err, "unable to parse data %v", data)
	}

	return ytbx.InputFile{
		Location:  location,
		Documents: documents,
	}, nil
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
