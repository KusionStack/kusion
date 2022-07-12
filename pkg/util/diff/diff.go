package diff

import (
	"bytes"
	"fmt"

	"github.com/gonvenience/wrap"
	"github.com/gonvenience/ytbx"
	yamlv3 "gopkg.in/yaml.v3"

	"kusionstack.io/kusion/pkg/util/yaml"
	"kusionstack.io/kusion/third_party/dyff"
)

// Supported output option values
const (
	OutputHuman = "human"
	OutputRaw   = "raw"
)

// NewHumanReport return a default *dyff.HumanReport with head omitted
func NewHumanReport(report *dyff.Report) *dyff.HumanReport {
	return &dyff.HumanReport{
		NoTableStyle:         false,
		DoNotInspectCerts:    false,
		OmitHeader:           true,
		UseGoPatchPaths:      false,
		MinorChangeThreshold: 0.1,
		Report:               *report,
	}
}

// ToReportString return a report string base on mode, valid mode: "human" and "raw"
func ToReportString(humanReport *dyff.HumanReport, mode string) (string, error) {
	switch mode {
	case OutputHuman:
		return ToHumanString(humanReport)
	case OutputRaw:
		return ToRawString(humanReport)
	default:
		return "", fmt.Errorf("invalid output style `%s`", mode)
	}
}

func ToHumanString(humanReport *dyff.HumanReport) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	err := humanReport.WriteReport(buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func ToRawString(humanReport *dyff.HumanReport) (string, error) {
	reportMap := map[string]interface{}{
		"diffs": humanReport.Diffs,
	}
	reportYAML, err := yamlv3.Marshal(reportMap)
	if err != nil {
		return "", wrap.Errorf(err, "failed to marshal report diffs")
	}
	return string(reportYAML), nil
}

// ToReport compares objects, oldData and newData,
// and returns a report with the list of differences.
func ToReport(oldData, newData interface{}) (*dyff.Report, error) {
	from, err := LoadFile(yaml.MergeToOneYAML(oldData), "Old item")
	if err != nil {
		return nil, err
	}

	to, err := LoadFile(yaml.MergeToOneYAML(newData), "New item")
	if err != nil {
		return nil, err
	}

	report, err := dyff.CompareInputFiles(from, to, dyff.IgnoreOrderChanges(true))
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// LoadFile reads the provided input data slice as a YAML, JSON, or TOML
// file with potential multiple documents.
func LoadFile(input, location string) (ytbx.InputFile, error) {
	var (
		documents []*yamlv3.Node
		data      []byte
		err       error
	)

	data = []byte(input)
	if documents, err = ytbx.LoadDocuments(data); err != nil {
		return ytbx.InputFile{}, wrap.Errorf(err, "unable to parse data %v", data)
	}

	return ytbx.InputFile{
		Location:  location,
		Documents: documents,
	}, nil
}
