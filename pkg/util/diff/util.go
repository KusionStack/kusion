package diff

import (
	"bytes"

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
