package diff

import (
	"bytes"

	"kusionstack.io/kusion/pkg/engine/operation/utils"
	"kusionstack.io/kusion/pkg/util/yaml"
	"kusionstack.io/kusion/third_party/dyff"
)

func ToReportString(report dyff.Report) (string, error) {
	reportWriter := &dyff.HumanReport{
		Report:               report,
		DoNotInspectCerts:    false,
		NoTableStyle:         false,
		OmitHeader:           true,
		UseGoPatchPaths:      false,
		MinorChangeThreshold: 0.1,
	}
	buf := bytes.NewBuffer([]byte{})
	err := reportWriter.WriteReport(buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ToReport compares objects, oldData and newData,
// and returns a report with the list of differences.
func ToReport(oldData, newData interface{}) (*dyff.Report, error) {
	from, err := utils.LoadFile(yaml.MergeToOneYAML(oldData), "Old item")
	if err != nil {
		return nil, err
	}

	to, err := utils.LoadFile(yaml.MergeToOneYAML(newData), "New item")
	if err != nil {
		return nil, err
	}

	report, err := dyff.CompareInputFiles(from, to, dyff.IgnoreOrderChanges(true))
	if err != nil {
		return nil, err
	}
	return &report, nil
}
