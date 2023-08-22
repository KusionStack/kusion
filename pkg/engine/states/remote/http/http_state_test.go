package http

import (
	"fmt"
	"github.com/bytedance/mockey"
	"io"
	"net/http"
	"strings"
	"testing"

	"kusionstack.io/kusion/pkg/engine/states"

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
		mockey.PatchConvey(tt.name, t, func() {
			s := &HTTPState{
				urlPrefix:          tt.fields.urlPrefix,
				applyURLFormat:     tt.fields.applyURLFormat,
				getLatestURLFormat: tt.fields.getLatestURLFormat,
			}
			mockey.Mock((*http.Client).Do).To(tt.mockFunc).Build()
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
		mockey.PatchConvey(tt.name, t, func() {
			s := &HTTPState{
				urlPrefix:          tt.fields.urlPrefix,
				applyURLFormat:     tt.fields.applyURLFormat,
				getLatestURLFormat: tt.fields.getLatestURLFormat,
			}
			mockey.Mock((*http.Client).Do).To(tt.mockFunc).Build()

			got, err := s.GetLatestState(tt.args.query)
			if !tt.wantErr(t, err, fmt.Sprintf("GetLatestState(%v)", tt.args.query)) {
				t.Errorf("wantErrFuncFailed:%v", err)
			}
			assert.Equalf(t, tt.want, got, "GetLatestState(%v)", tt.args.query)
		})
	}
}
