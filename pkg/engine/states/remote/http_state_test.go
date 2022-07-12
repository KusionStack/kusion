package remote

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"kusionstack.io/kusion/pkg/engine/states"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	json_util "kusionstack.io/kusion/pkg/util/json"
)

const (
	prefix = "https://kusionstack.io"
	format = "/apis/v1/tenants/%s/projects/%s/stacks/%s/cluster/%s/states/"
)

func TestHTTPState_Apply(t *testing.T) {
	type fields struct {
		urlPrefix          string
		applyURLFormat     string
		getLatestURLFormat string
	}
	type args struct {
		state *states.State
	}

	state := states.NewState()
	state.Tenant = "t"
	state.Project = "p"
	state.Stack = "s"
	state.Cluster = "c"

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  assert.ErrorAssertionFunc
		mockFunc interface{}
	}{
		{
			name: "apply",
			fields: fields{
				urlPrefix:          prefix,
				applyURLFormat:     format,
				getLatestURLFormat: format,
			},
			args: args{state: state},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
			mockFunc: func(c *http.Client, req *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "Success",
					StatusCode: 200,
					Body:       http.NoBody,
				}, nil
			},
		},
		{
			name: "apply_error",
			fields: fields{
				urlPrefix:          prefix,
				applyURLFormat:     format,
				getLatestURLFormat: format,
			},
			args: args{state: state},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err != nil && strings.Contains(err.Error(), "apply state failed")
			},
			mockFunc: func(c *http.Client, req *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "NotFound",
					StatusCode: 404,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPState{
				urlPrefix:          tt.fields.urlPrefix,
				applyURLFormat:     tt.fields.applyURLFormat,
				getLatestURLFormat: tt.fields.getLatestURLFormat,
			}
			monkey.Patch((*http.Client).Do, tt.mockFunc)
			err := s.Apply(tt.args.state)
			if !tt.wantErr(t, err, fmt.Sprintf("Apply(%v)", tt.args.state)) {
				t.Errorf("wantErrFuncFailed:%v", err)
			}
		})
	}
}

func TestHTTPState_GetLatestState(t *testing.T) {
	type fields struct {
		urlPrefix          string
		applyURLFormat     string
		getLatestURLFormat string
	}
	type args struct {
		query *states.StateQuery
	}

	state := states.NewState()
	state.Tenant = "t"
	state.Project = "p"
	state.Stack = "s"
	state.Cluster = "c"

	query := &states.StateQuery{
		Tenant:  "t",
		Project: "p",
		Stack:   "s",
		Cluster: "c",
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     *states.State
		wantErr  assert.ErrorAssertionFunc
		mockFunc interface{}
	}{
		{
			name: "GetLatestState",
			fields: fields{
				urlPrefix:          prefix,
				applyURLFormat:     format,
				getLatestURLFormat: format,
			},
			args: args{query: query},
			want: state,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
			mockFunc: func(c *http.Client, req *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "Success",
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(json_util.Marshal2String(state))),
				}, nil
			},
		},
		{
			name: "GetLatestStateNotFound",
			fields: fields{
				urlPrefix:          prefix,
				applyURLFormat:     format,
				getLatestURLFormat: format,
			},
			want: nil,
			args: args{query: query},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
			mockFunc: func(c *http.Client, req *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "NotFound",
					StatusCode: 404,
					Body:       http.NoBody,
				}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPState{
				urlPrefix:          tt.fields.urlPrefix,
				applyURLFormat:     tt.fields.applyURLFormat,
				getLatestURLFormat: tt.fields.getLatestURLFormat,
			}
			monkey.Patch((*http.Client).Do, tt.mockFunc)

			got, err := s.GetLatestState(tt.args.query)
			if !tt.wantErr(t, err, fmt.Sprintf("GetLatestState(%v)", tt.args.query)) {
				t.Errorf("wantErrFuncFailed:%v", err)
			}
			assert.Equalf(t, tt.want, got, "GetLatestState(%v)", tt.args.query)
		})
	}
}

func TestNewHTTPState(t *testing.T) {
	type args struct {
		params map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *HTTPState
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "NewState",
			args: args{
				params: map[string]interface{}{
					"urlPrefix":          prefix,
					"applyURLFormat":     format,
					"getLatestURLFormat": format,
				},
			},
			want: &HTTPState{
				urlPrefix:          prefix,
				applyURLFormat:     format,
				getLatestURLFormat: format,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "Empty value",
			args: args{
				params: map[string]interface{}{
					"urlPrefix":          nil,
					"applyURLFormat":     format,
					"getLatestURLFormat": format,
				},
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return strings.Contains(err.Error(), "urlPrefix can not be empty")
			},
		},
		{
			name: "Invalidate format",
			args: args{
				params: map[string]interface{}{
					"urlPrefix":          prefix,
					"applyURLFormat":     "invalidate_format",
					"getLatestURLFormat": format,
				},
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return strings.Contains(err.Error(), "applyURLFormat must contains 4 \"%s\" placeholders for tenant, project, "+
					"stack and cluster")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHTTPState(tt.args.params)
			if !tt.wantErr(t, err, fmt.Sprintf("NewHTTPState(%v)", tt.args.params)) {
				t.Errorf("wantErrFuncFailed:%v", err)
			}
			assert.Equalf(t, tt.want, got, "NewHTTPState(%v)", tt.args.params)
		})
	}
}
