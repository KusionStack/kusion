package states

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"kusionstack.io/kusion/pkg/log"
)

// HTTPState represent a remote state that can be requested by HTTP.
// This state is designed to provide a generic way to manipulate State in third-party services
//
// Some url formats are given to bring relative flexibility for third-party services to implement their own State HTTP service and these
// formats MUST contain 3 "%s" placeholders for tenant, project and stack, since we will replace this format with fmt.Sprintf()
// Let's get applyURLFormat as an example to demonstrate how this suffix format works.
//
//
//  Example:
//
//	urlPrefix = "http://kusionstack.io"
//	applyURLFormat = "/apis/v1/tenants/%s/projects/%s/stacks/%s/states"
//	tenant = "t"
//	project = "p"
//	stack = "s"
//	the final request URL = "http://kusionstack.io/apis/v1/tenants/t/projects/p/stacks/s/states"
type HTTPState struct {
	// urlPrefix is the prefix added in front of all request URLs. e.g. "http://kusionstack.io/"
	urlPrefix string

	// applyURLFormat is the suffix url format to apply a state
	applyURLFormat string

	// getLatestURLFormat is the suffix url format to get the latest state
	getLatestURLFormat string
}

// NewHTTPState builds a new HTTPState with ConfigSchema() and validates params with Configure()
func NewHTTPState(params map[string]interface{}) (*HTTPState, error) {
	h := &HTTPState{}
	ctyValue, err := gocty.ToCtyValue(params, (&HTTPState{}).ConfigSchema())
	if err != nil {
		return nil, err
	}

	if e := h.Configure(ctyValue); e != nil {
		return nil, e
	}
	return h, nil
}

// ConfigSchema is an implementation of StateStorage.ConfigSchema
func (s *HTTPState) ConfigSchema() cty.Type {
	config := map[string]cty.Type{
		"urlPrefix":          cty.String,
		"applyURLFormat":     cty.String,
		"getLatestURLFormat": cty.String,
	}
	return cty.Object(config)
}

// Configure is an implementation of StateStorage.Configure
func (s *HTTPState) Configure(obj cty.Value) error {
	var url cty.Value

	if url = obj.GetAttr("urlPrefix"); url.IsNull() || url.AsString() == "" {
		return errors.New("urlPrefix can not be empty")
	} else {
		s.urlPrefix = url.AsString()
	}

	if applyFormat := obj.GetAttr("applyURLFormat"); applyFormat.IsNull() || url.AsString() == "" {
		return errors.New("applyURLFormat can not be empty")
	} else {
		asString := applyFormat.AsString()
		count := strings.Count(asString, "%s")
		if count != 3 {
			return errors.New("applyURLFormat must contains 3 \"%s\" placeholders for tenant, project, stack. Current format:" + asString)
		}
		s.applyURLFormat = asString
	}

	if getLatest := obj.GetAttr("getLatestURLFormat"); getLatest.IsNull() && getLatest.AsString() == "" {
		return errors.New("getLatestURLFormat can not be empty")
	} else {
		asString := getLatest.AsString()
		count := strings.Count(asString, "%s")
		if count != 3 {
			return errors.New("getLatestURLFormat must contains 3 \"%s\" placeholders for tenant, project, stack. Current format:" + asString)
		}
		s.getLatestURLFormat = asString
	}

	return nil
}

// GetLatestState is an implementation of StateStorage.GetLatestState
func (s *HTTPState) GetLatestState(query *StateQuery) (*State, error) {
	url := fmt.Sprintf("%s"+s.getLatestURLFormat, s.urlPrefix, query.Tenant, query.Project, query.Stack)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		log.Info("Can't find the latest state by request:%s", url)
		return nil, nil
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("get the latest state failed. StatusCode:%v, Status:%s", res.StatusCode, res.Status)
	}

	state := &State{}
	resBody, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(resBody, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// Apply is an implementation of StateStorage.Apply
func (s *HTTPState) Apply(state *State) error {
	jsonState, err := json.Marshal(state)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s"+s.applyURLFormat, s.urlPrefix, state.Tenant, state.Project, state.Stack)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonState)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("apply state failed. StatusCode:%v, Status:%s", res.StatusCode, res.Status)
	}
	defer res.Body.Close()

	return nil
}

// Delete is not support now
func (s *HTTPState) Delete(id string) error {
	return errors.New("not supported")
}
