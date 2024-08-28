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
