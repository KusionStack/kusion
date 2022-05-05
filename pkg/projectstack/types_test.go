package projectstack

import (
	"reflect"
	"testing"

	"github.com/pterm/pterm"
)

func TestNewProject(t *testing.T) {
	type args struct {
		config *ProjectConfiguration
		path   string
		stacks []*Stack
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
				stacks: []*Stack{},
			},
			want: &Project{
				ProjectConfiguration: ProjectConfiguration{},
				Path:                 "test",
				Stacks:               []*Stack{},
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
		Stacks               []*Stack
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
		Stacks               []*Stack
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
		Stacks               []*Stack
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
				Stacks: []*Stack{
					{
						StackConfiguration: StackConfiguration{
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

func TestNewStack(t *testing.T) {
	type args struct {
		config *StackConfiguration
		path   string
	}
	tests := []struct {
		name string
		args args
		want *Stack
	}{
		{
			name: "success",
			args: args{
				config: &StackConfiguration{
					Name: TestStackA,
				},
				path: TestStackPathAA,
			},
			want: &Stack{
				StackConfiguration: StackConfiguration{
					Name: TestStackA,
				},
				Path: TestStackPathAA,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStack(tt.args.config, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStack_GetName(t *testing.T) {
	type fields struct {
		StackConfiguration StackConfiguration
		Path               string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				StackConfiguration: StackConfiguration{
					Name: TestStackA,
				},
				Path: TestStackPathAA,
			},
			want: TestStackA,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stack{
				StackConfiguration: tt.fields.StackConfiguration,
				Path:               tt.fields.Path,
			}
			if got := s.GetName(); got != tt.want {
				t.Errorf("Stack.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStack_GetPath(t *testing.T) {
	type fields struct {
		StackConfiguration StackConfiguration
		Path               string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				StackConfiguration: StackConfiguration{
					Name: TestStackA,
				},
				Path: TestStackPathAA,
			},
			want: TestStackPathAA,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stack{
				StackConfiguration: tt.fields.StackConfiguration,
				Path:               tt.fields.Path,
			}
			if got := s.GetPath(); got != tt.want {
				t.Errorf("Stack.GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStack_TableReport(t *testing.T) {
	type fields struct {
		StackConfiguration StackConfiguration
		Path               string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				StackConfiguration: StackConfiguration{
					Name: TestStackA,
				},
				Path: TestStackPathAA,
			},
			want: `┌────────────────────────────────────────────┐
| Type       | Name                          |
| Stack Name | dev                           |
| Stack Path | testdata/appops/http-echo/dev |
└────────────────────────────────────────────┘`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stack{
				StackConfiguration: tt.fields.StackConfiguration,
				Path:               tt.fields.Path,
			}
			got := pterm.RemoveColorFromString(s.TableReport())
			if got != tt.want {
				t.Errorf("Stack.TableReport() = %v, want %v", got, tt.want)
			}
		})
	}
}
