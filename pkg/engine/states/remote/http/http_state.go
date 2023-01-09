package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/log"
)

// HTTPState represent a remote state that can be requested by HTTP.
// This state is designed to provide a generic way to manipulate State in third-party services
//
// Some url formats are given to bring relative flexibility for third-party services to implement their own State HTTP service and these
// formats MUST contain 4 "%s" placeholders for tenant, project and stack, since we will replace this format with fmt.Sprintf()
// Let's get applyURLFormat as an example to demonstrate how this suffix format works.
//
//	 Example:
//
//		urlPrefix = "http://kusionstack.io"
//		applyURLFormat = "/apis/v1/tenants/%s/projects/%s/stacks/%s/clusters/%s/states/"
//		tenant = "t"
//		project = "p"
//		stack = "s"
//	 cluster = "c"
//		the final request URL = "http://kusionstack.io/apis/v1/tenants/t/projects/p/stacks/s/clusters/c/states"
type HTTPState struct {
	// urlPrefix is the prefix added in front of all request URLs. e.g. "http://kusionstack.io/"
	urlPrefix string

	// applyURLFormat is the suffix url format to apply a state
	applyURLFormat string

	// getLatestURLFormat is the suffix url format to get the latest state
	getLatestURLFormat string
}

const ParamsCounts = 4

// GetLatestState is an implementation of StateStorage.GetLatestState
func (s *HTTPState) GetLatestState(query *states.StateQuery) (*states.State, error) {
	url := fmt.Sprintf("%s"+s.getLatestURLFormat, s.urlPrefix, query.Tenant, query.Project, query.Stack, query.Cluster)
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

	state := &states.State{}
	resBody, _ := io.ReadAll(res.Body)
	err = json.Unmarshal(resBody, state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// Apply is an implementation of StateStorage.Apply
func (s *HTTPState) Apply(state *states.State) error {
	jsonState, err := json.Marshal(state)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s"+s.applyURLFormat, s.urlPrefix, state.Tenant, state.Project, state.Stack, state.Cluster)

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
