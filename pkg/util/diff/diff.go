package diff

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/gonvenience/wrap"
	"github.com/gonvenience/ytbx"
	yamlv3 "gopkg.in/yaml.v3"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/util/yaml"
	"kusionstack.io/kusion/third_party/dyff"
)

// Supported output option values
const (
	OutputHuman = "human"
	OutputRaw   = "raw"
)

// Placeholders for masking sensitive information.
const (
	maskStr       = "*******"
	maskStrBefore = "***before***"
	maskStrAfter  = "***after****"
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
	// Mask the sensitive data in Kubernetes Secret before generating the
	// diff report.
	maskedOldData, maskedNewData := maskSensitiveData(oldData, newData)

	from, err := LoadFile(yaml.MergeToOneYAML(maskedOldData), "Old item")
	if err != nil {
		return nil, err
	}

	to, err := LoadFile(yaml.MergeToOneYAML(maskedNewData), "New item")
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

// maskSensitiveData masks the sensitive data with placeholders before generating
// the diff report.
func maskSensitiveData(oldData, newData interface{}) (interface{}, interface{}) {
	from, ok1 := oldData.(*v1.Resource)
	to, ok2 := newData.(*v1.Resource)

	// Return directly if oldData or newData can not be transferred into v1.Resource.
	if !(ok1 && ok2) {
		return oldData, newData
	}

	// Record whether we need to mask the old or the new object and the 'data'
	// and 'stringData' attributes of the secret resource.
	maskOld, maskNew := false, false
	fromSecData, toSecData := map[string]interface{}{}, map[string]interface{}{}
	fromSecStrData, toSecStrData := map[string]interface{}{}, map[string]interface{}{}

	// Prepare the masked old secret resource and masked new secret resource.
	maskedOldData, maskedNewData := &v1.Resource{}, &v1.Resource{}

	// Check if the resource type is Kubernetes Secret.
	if from != nil {
		if _, ok := from.Attributes["kind"]; ok {
			fromKind, ok := from.Attributes["kind"].(string)
			if from.Type == v1.Kubernetes && ok && fromKind == "Secret" {
				// Set masking old data to true.
				maskOld = true
				deepCopyResource(from, maskedOldData)

				// Append 'data' and 'stringData' attributes of the old secret resource.
				if _, ok := from.Attributes["data"]; ok {
					data, ok := from.Attributes["data"].(map[string]interface{})
					if ok {
						for k, v := range data {
							fromSecData[k] = v
						}
					}
				}

				if _, ok := from.Attributes["stringData"]; ok {
					strData, ok := from.Attributes["stringData"].(map[string]interface{})
					if ok {
						for k, v := range strData {
							fromSecStrData[k] = v
						}
					}
				}
			}
		}
	} else {
		maskedOldData = nil
	}

	if to != nil {
		if _, ok := to.Attributes["kind"]; ok {
			toKind, ok := to.Attributes["kind"].(string)
			if to.Type == v1.Kubernetes && ok && toKind == "Secret" {
				// Set masking new data to true.
				maskNew = true
				deepCopyResource(to, maskedNewData)

				// Append 'data' and 'stringData' attributes of the new secret resource.
				if _, ok := to.Attributes["data"]; ok {
					data, ok := to.Attributes["data"].(map[string]interface{})
					if ok {
						for k, v := range data {
							toSecData[k] = v
						}
					}
				}

				if _, ok := to.Attributes["stringData"]; ok {
					strData, ok := to.Attributes["stringData"].(map[string]interface{})
					if ok {
						for k, v := range strData {
							toSecStrData[k] = v
						}
					}
				}
			}
		}
	} else {
		maskedNewData = nil
	}

	// Return the original data if do not need to mask the sensitive data.
	if !maskOld && !maskNew {
		return oldData, newData
	}

	// Replace the 'data' and 'stringData' attributes of the old secret resource.
	if maskOld {
		for k, v := range fromSecData {
			var secStr string
			secStr, ok := v.(string)
			if !ok {
				continue
			}

			if toV, ok := toSecData[k]; ok && secStr != toV.(string) {
				maskedOldData.Attributes["data"].(map[string]interface{})[k] = maskStrBefore
			} else {
				maskedOldData.Attributes["data"].(map[string]interface{})[k] = maskStr
			}
		}

		for k, v := range fromSecStrData {
			var secStr string
			secStr, ok := v.(string)
			if !ok {
				continue
			}

			if toV, ok := toSecStrData[k]; ok && secStr != toV.(string) {
				maskedOldData.Attributes["stringData"].(map[string]interface{})[k] = maskStrBefore
			} else {
				maskedOldData.Attributes["stringData"].(map[string]interface{})[k] = maskStr
			}
		}
	}

	// Replace the 'data' and 'stringData' attributes of the new secret resource.
	if maskNew {
		for k, v := range toSecData {
			var secStr string
			secStr, ok := v.(string)
			if !ok {
				continue
			}

			if fromV, ok := fromSecData[k]; ok && secStr != fromV.(string) {
				maskedNewData.Attributes["data"].(map[string]interface{})[k] = maskStrAfter
			} else {
				maskedNewData.Attributes["data"].(map[string]interface{})[k] = maskStr
			}
		}

		for k, v := range toSecStrData {
			var secStr string
			secStr, ok := v.(string)
			if !ok {
				continue
			}

			if fromV, ok := fromSecStrData[k]; ok && secStr != fromV.(string) {
				maskedNewData.Attributes["stringData"].(map[string]interface{})[k] = maskStrAfter
			} else {
				maskedNewData.Attributes["stringData"].(map[string]interface{})[k] = maskStr
			}
		}
	}

	return maskedOldData, maskedNewData
}

// deepCopyResource deeply copies the old Resource into a new one.
func deepCopyResource(from, to *v1.Resource) error {
	to.ID = from.ID
	to.Type = from.Type

	var err error
	if len(from.Attributes) != 0 {
		if to.Attributes, err = deepCopyMap(from.Attributes); err != nil {
			return err
		}
	}

	if len(from.Extensions) != 0 {
		if to.Extensions, err = deepCopyMap(from.Extensions); err != nil {
			return err
		}
	}

	if len(from.DependsOn) != 0 {
		if len(to.DependsOn) == 0 {
			to.DependsOn = make([]string, len(from.DependsOn))
		}
		copy(to.DependsOn, from.DependsOn)
	}

	return nil
}

// deepCopyMap deeply copies the map[string]interface{}.
func deepCopyMap(src map[string]interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	var dest map[string]interface{}
	if err = json.Unmarshal(jsonBytes, &dest); err != nil {
		return nil, err
	}

	return dest, nil
}
