package project

import (
	"reflect"
	"testing"

	"github.com/pterm/pterm"

	"kusionstack.io/kusion/pkg/apis/stack"
)

func TestNewProject(t *testing.T) {
	type args struct {
		config *ProjectConfiguration
		path   string
		stacks []*stack.Stack
	}
	tests := []struct {
		name string
		args args
		want *Project
	}{
		{
			name: "success",
			args: args{
				config: &ProjectConfiguration{},
				path:   "test",
				stacks: []*stack.Stack{},
			},
			want: &Project{
				ProjectConfiguration: ProjectConfiguration{},
				Path:                 "test",
				Stacks:               []*stack.Stack{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProject(tt.args.config, tt.args.path, tt.args.stacks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_GetName(t *testing.T) {
	type fields struct {
		ProjectConfiguration ProjectConfiguration
		Path                 string
		Stacks               []*stack.Stack
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				ProjectConfiguration: ProjectConfiguration{
					Name: "test",
				},
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Project{
				ProjectConfiguration: tt.fields.ProjectConfiguration,
				Path:                 tt.fields.Path,
				Stacks:               tt.fields.Stacks,
			}
			if got := p.GetName(); got != tt.want {
				t.Errorf("Project.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_GetPath(t *testing.T) {
	type fields struct {
		ProjectConfiguration ProjectConfiguration
		Path                 string
		Stacks               []*stack.Stack
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				Path: "test",
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Project{
				ProjectConfiguration: tt.fields.ProjectConfiguration,
				Path:                 tt.fields.Path,
				Stacks:               tt.fields.Stacks,
			}
			if got := p.GetPath(); got != tt.want {
				t.Errorf("Project.GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_TableReport(t *testing.T) {
	type fields struct {
		ProjectConfiguration ProjectConfiguration
		Path                 string
		Stacks               []*stack.Stack
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				ProjectConfiguration: ProjectConfiguration{
					Name:   TestProjectA,
					Tenant: "main",
				},
				Path: TestProjectPathA,
				Stacks: []*stack.Stack{
					{
						Configuration: stack.Configuration{
							Name: TestStackA,
						},
						Path: TestStackPathAA,
					},
				},
			},
			want: `┌──────────────────────────────────────────┐
| Type         | Name                      |
| Project Name | http-echo                 |
| Project Path | testdata/appops/http-echo |
| Tenant       | main                      |
| Stacks       | dev                       |
└──────────────────────────────────────────┘`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Project{
				ProjectConfiguration: tt.fields.ProjectConfiguration,
				Path:                 tt.fields.Path,
				Stacks:               tt.fields.Stacks,
			}
			got := pterm.RemoveColorFromString(p.TableReport())
			if got != tt.want {
				t.Errorf("Project.TableReport() = %v, want %v", got, tt.want)
			}
		})
	}
}
