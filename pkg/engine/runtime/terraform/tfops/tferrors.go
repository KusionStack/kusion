package tfops

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// TerraformInfo represent fields of a Terraform CLI JSON-formatted log line
type TerraformInfo struct {
	Level      string     `json:"@level"`
	Message    string     `json:"@message"`
	Module     string     `json:"@module"`
	Timestamp  time.Time  `json:"@timestamp"`
	Diagnostic Diagnostic `json:"diagnostic"`
	Type       string     `json:"type"`
}

// Diagnostic schema from https://github.com/hashicorp/terraform/blob/main/internal/command/views/json/diagnostic.go.

// Diagnostic represents relevant fields of a Terraform CLI JSON-formatted
// log line diagnostic info
type Diagnostic struct {
	Severity string  `json:"severity"`
	Summary  string  `json:"summary"`
	Detail   string  `json:"detail"`
	Range    Range   `json:"range"`
	Snippet  Snippet `json:"snippet"`
}

// Pos represents a position in the source code.
type Pos struct {
	// Line is a one-based count for the line in the indicated file.
	Line int `json:"line"`

	// Column is a one-based count of Unicode characters from the start of the line.
	Column int `json:"column"`

	// Byte is a zero-based offset into the indicated file.
	Byte int `json:"byte"`
}

// DiagnosticRange represents the filename and position of the diagnostic
// subject. This defines the range of the source to be highlighted in the
// output. Note that the snippet may include additional surrounding source code
// if the diagnostic has a context range.
//
// The Start position is inclusive, and the End position is exclusive. Exact
// positions are intended for highlighting for human interpretation only and
// are subject to change.
type Range struct {
	Filename string `json:"filename"`
	Start    Pos    `json:"start"`
	End      Pos    `json:"end"`
}

// Snippet represents source code information about the diagnostic.
// It is possible for a diagnostic to have a source (and therefore a range) but
// no source code can be found. In this case, the range field will be present and
// the snippet field will not.
type Snippet struct {
	Context              string        `json:"context"`
	Code                 string        `json:"code"`
	StartLine            int           `json:"start_line"`
	HighlightStartOffset int           `json:"highlight_start_offset"`
	HighlightEndOffset   int           `json:"highlight_end_offset"`
	Values               []interface{} `json:"values"`
}

// Parse Terraform CLI output infos
func parseTerraformInfo(infos []byte) ([]*TerraformInfo, error) {
	info := strings.Split(string(infos), "\n")
	tfInfos := make([]*TerraformInfo, 0, len(info))
	for _, v := range info {
		terraformInfo := &TerraformInfo{}
		if v == "" {
			continue
		}
		err := json.Unmarshal([]byte(v), terraformInfo)
		if err != nil {
			return nil, err
		}
		tfInfos = append(tfInfos, terraformInfo)
	}
	return tfInfos, nil
}

// TFError parse Terraform CLI output infos
// return error with given infos
func TFError(infos []byte) error {
	// todo @Markliby TF error outputs are formatted as TerraformInfo only when TF_LOG is TRACE or higher.
	// The output often looks like this when the log level is INFO:
	// 2022-11-07T17:41:29.643+0800 [INFO]  Terraform version: 1.3.4
	// 2022-11-07T17:41:29.643+0800 [INFO]  Go runtime version: go1.19.3
	// 2022-11-07T17:41:29.643+0800 [INFO]  CLI args: []string{"terraform", "apply", "-auto-approve", "-json", "-lock=false"}
	// ... ...
	// {"@level":"info","@message":"Terraform 1.3.4","@module":"terraform.ui","@timestamp":"2022-11-07T17:41:29.647389+08:00",
	// "terraform":"1.3.4","type":"version","ui":"1.0"}
	// we should figure out a more user-friendly way to handle TF errors
	// return errors.New(string(infos))
	//
	// So, we can check the TF_LOG level here and unmarshall the output to a struct.

	tfInfo, err := parseTerraformInfo(infos)
	if err != nil {
		return err
	}
	for _, v := range tfInfo {
		if v == nil || v.Level != "error" {
			continue
		}
		if v.Diagnostic.Severity == "error" {
			msg := fmt.Sprintf("%s: %s", v.Diagnostic.Summary, v.Diagnostic.Detail)
			return errors.New(msg)
		}
	}
	return nil
}
