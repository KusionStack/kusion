package list

import (
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/backend/storages"
	workspacestorages "kusionstack.io/kusion/pkg/workspace/storages"
)

var projects = map[string][]string{
	"default": {"project1", "project2"},
	"dev":     {"project3", "project4"},
}

var workspaces = []string{"default", "dev"}

func TestNewFlags(t *testing.T) {
	tests := []struct {
		name string
		want *Flags
	}{
		{
			name: "Create new flags successfully",
			want: &Flags{
				Backend:   new(string),
				Workspace: &[]string{},
				All:       false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFlags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlags_AddFlags(t *testing.T) {
	type fields struct {
		Backend   *string
		Workspace *[]string
		All       bool
	}
	type args struct {
		cmd *cobra.Command
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Add flags successfully",
			fields: fields{
				Backend:   new(string),
				Workspace: &[]string{},
				All:       false,
			},
			args: args{
				cmd: &cobra.Command{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Flags{
				Backend:   tt.fields.Backend,
				Workspace: tt.fields.Workspace,
				All:       tt.fields.All,
			}
			f.AddFlags(tt.args.cmd)
		})
	}
}

func TestFlags_ToOptions(t *testing.T) {
	type fields struct {
		Backend   *string
		Workspace *[]string
		All       bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Options
		wantErr bool
	}{
		{
			name: "Empty workspace and backend",
			fields: fields{
				Backend:   new(string),
				Workspace: &[]string{},
				All:       false,
			},
			want: &Options{
				projects:         projects,
				Workspace:        []string{"default"},
				CurrentWorkspace: "default",
			},
			wantErr: false,
		},
		{
			name: "All workspace",
			fields: fields{
				Backend:   new(string),
				Workspace: &[]string{},
				All:       true,
			},
			want: &Options{
				projects:         projects,
				Workspace:        []string{"default", "dev"},
				CurrentWorkspace: "default",
			},
			wantErr: false,
		},
		{
			name: "Multi workspace",
			fields: fields{
				Backend:   new(string),
				Workspace: &[]string{"dev", "default"},
				All:       false,
			},
			want: &Options{
				projects:         projects,
				Workspace:        []string{"dev", "dev"},
				CurrentWorkspace: "default",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockey.PatchConvey("mock workspaceStorage", t, func() {
				flag := &Flags{
					Backend:   tt.fields.Backend,
					Workspace: tt.fields.Workspace,
					All:       tt.fields.All,
				}
				mockey.Mock((*storages.LocalStorage).WorkspaceStorage).Return(&workspacestorages.LocalStorage{}, nil).Build()
				mockey.Mock((*workspacestorages.LocalStorage).GetCurrent).Return("default", nil).Build()
				mockey.Mock((*workspacestorages.LocalStorage).GetNames).Return(workspaces, nil).Build()
				mockey.Mock((*workspacestorages.LocalStorage).Get).Return(&v1.Workspace{Name: "dev"}, nil).Build()
				mockey.Mock((*storages.LocalStorage).ProjectStorage).Return(projects, nil).Build()
				options, err := flag.ToOptions()
				assert.NoError(t, err)
				assert.Equal(t, tt.want, options)
			})
		})
	}
}

func TestOptions_Validate(t *testing.T) {
	type fields struct {
		projects         map[string][]string
		Workspace        []string
		CurrentWorkspace string
	}
	type args struct {
		cmd  *cobra.Command
		args []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "No error",
			fields: fields{},
			args: args{
				cmd:  &cobra.Command{},
				args: []string{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Options{
				projects:         tt.fields.projects,
				Workspace:        tt.fields.Workspace,
				CurrentWorkspace: tt.fields.CurrentWorkspace,
			}
			if err := o.Validate(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("Options.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOptions_Run(t *testing.T) {
	type fields struct {
		projects         map[string][]string
		Workspace        []string
		CurrentWorkspace string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Print projects successfully",
			fields: fields{
				projects:         projects,
				Workspace:        workspaces,
				CurrentWorkspace: "default",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Options{
				projects:         tt.fields.projects,
				Workspace:        tt.fields.Workspace,
				CurrentWorkspace: tt.fields.CurrentWorkspace,
			}
			if err := o.Run(); (err != nil) != tt.wantErr {
				t.Errorf("Options.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
