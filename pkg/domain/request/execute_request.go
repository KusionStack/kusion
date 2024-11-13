package request

import (
	"net/http"
)

type StackImportRequest struct {
	ImportedResources map[string]string `json:"importedResources"`
}

func (payload *StackImportRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}

type CreateRunRequest struct {
	Type              string             `json:"type"`
	StackID           uint               `json:"stackID"`
	Workspace         string             `json:"workspace"`
	ImportedResources StackImportRequest `json:"importedResources"`
}

type UpdateRunRequest struct {
	CreateRunRequest `json:",inline" yaml:",inline"`
}

type UpdateRunResultRequest struct {
	Result string `json:"result"`
	Status string `json:"status"`
	Logs   string `json:"logs"`
}

func (payload *CreateRunRequest) Decode(r *http.Request) error {
	return decode(r, payload)
}
