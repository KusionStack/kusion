package http

import (
	"errors"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"kusionstack.io/kusion/pkg/engine/states"
)

type HTTPBackend struct {
	HTTPState
}

func NewHTTPBackend() states.Backend {
	return &HTTPBackend{}
}

// ConfigSchema is an implementation of StateStorage.ConfigSchema
func (b *HTTPBackend) ConfigSchema() cty.Type {
	config := map[string]cty.Type{
		"urlPrefix":          cty.String,
		"applyURLFormat":     cty.String,
		"getLatestURLFormat": cty.String,
	}
	return cty.Object(config)
}

// Configure is an implementation of StateStorage.Configure
func (b *HTTPBackend) Configure(obj cty.Value) error {
	var url cty.Value

	if url = obj.GetAttr("urlPrefix"); url.IsNull() || url.AsString() == "" {
		return errors.New("urlPrefix can not be empty")
	} else {
		b.urlPrefix = url.AsString()
	}

	if applyFormat := obj.GetAttr("applyURLFormat"); applyFormat.IsNull() || url.AsString() == "" {
		return errors.New("applyURLFormat can not be empty")
	} else {
		asString := applyFormat.AsString()
		count := strings.Count(asString, "%s")
		if count != ParamsCounts {
			return errors.New("applyURLFormat must contains 4 \"%s\" placeholders for tenant, project, " +
				"stack and cluster. Current format:" + asString)
		}
		b.applyURLFormat = asString
	}

	if getLatest := obj.GetAttr("getLatestURLFormat"); getLatest.IsNull() && getLatest.AsString() == "" {
		return errors.New("getLatestURLFormat can not be empty")
	} else {
		asString := getLatest.AsString()
		count := strings.Count(asString, "%s")
		if count != ParamsCounts {
			return errors.New("getLatestURLFormat must contains 4 \"%s\" placeholders for tenant, project, " +
				"stack or cluster. Current format:" + asString)
		}
		b.getLatestURLFormat = asString
	}

	return nil
}

// StateStorage return a StateStorage to manage http State
func (b *HTTPBackend) StateStorage() states.StateStorage {
	return &HTTPState{
		urlPrefix:          b.urlPrefix,
		applyURLFormat:     b.applyURLFormat,
		getLatestURLFormat: b.getLatestURLFormat,
	}
}
