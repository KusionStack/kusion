package operation

import (
	"reflect"
	"testing"

	"kusionstack.io/kusion/pkg/engine/models"

	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/pretty"
)

var (
	TestChangeStepOpCreate   = NewChangeStep("id", Create, nil, nil)
	TestChangeStepOpDelete   = NewChangeStep("id", Delete, nil, nil)
	TestChangeStepOpUpdate   = NewChangeStep("id", Update, nil, nil)
	TestChangeStepOpUnChange = NewChangeStep("id", UnChange, nil, nil)
	TestStepKeys             = []string{"test-key-1", "test-key-2", "test-key-3", "test-key-4"}
	TestChangeSteps          = map[string]*ChangeStep{
		"test-key-1": TestChangeStepOpCreate,
		"test-key-2": TestChangeStepOpDelete,
		"test-key-3": TestChangeStepOpUpdate,
		"test-key-4": TestChangeStepOpUnChange,
	}
)

func TestOpType_Ing(t *testing.T) {
	tests := []struct {
		name string
		op   ActionType
		want string
	}{
		{
			name: "t1",
			op:   Create,
			want: "Creating",
		},
		{
			name: "t2",
			op:   Delete,
			want: "Deleting",
		},
		{
			name: "t3",
			op:   Update,
			want: "Updating",
		},
		{
			name: "t4",
			op:   UnChange,
			want: "Unchanged",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.Ing(); got != tt.want {
				t.Errorf("ActionType.Ing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpType_PrettyString(t *testing.T) {
	tests := []struct {
		name string
		op   ActionType
		want string
	}{
		{
			name: "t1",
			op:   Create,
			want: pretty.Green(Create.Ing()),
		},
		{
			name: "t2",
			op:   Delete,
			want: pretty.Red(Delete.Ing()),
		},
		{
			name: "t3",
			op:   Update,
			want: pretty.Blue(Update.Ing()),
		},
		{
			name: "t4",
			op:   UnChange,
			want: pretty.Gray(UnChange.Ing()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.PrettyString(); got != tt.want {
				t.Errorf("ActionType.PrettyString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeStep_Diff(t *testing.T) {
	type fields struct {
		ID  string
		Op  ActionType
		Old interface{}
		New interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "t1",
			fields: fields{
				ID:  "id",
				Op:  Create,
				Old: nil,
				New: nil,
			},
			want: `[32;1m[32;1mID: [0m[0m[32mid[0m
[32m[0m[32;1m[32;1mPlan: [0m[0m[32mCreating[0m
[32;1m[32;1mDiff: [0m[0m

`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &ChangeStep{
				ID:     tt.fields.ID,
				Action: tt.fields.Op,
				Old:    tt.fields.Old,
				New:    tt.fields.New,
			}
			got, err := cs.Diff()
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangeStep.Diff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ChangeStep.Diff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewChangeStep(t *testing.T) {
	type args struct {
		id  string
		op  ActionType
		old interface{}
		new interface{}
	}
	tests := []struct {
		name string
		args args
		want *ChangeStep
	}{
		{
			name: "t1",
			args: args{
				id:  "id",
				op:  Create,
				old: nil,
				new: nil,
			},
			want: &ChangeStep{
				ID:     "id",
				Action: Create,
				Old:    nil,
				New:    nil,
			},
		},
		{
			name: "t2",
			args: args{
				id:  "id[0]",
				op:  Create,
				old: nil,
				new: nil,
			},
			want: &ChangeStep{
				ID:     "id[0]",
				Action: Create,
				Old:    nil,
				New:    nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewChangeStep(tt.args.id, tt.args.op, tt.args.old, tt.args.new); !reflect.DeepEqual(got,
				tt.want) {
				t.Errorf("NewChangeStep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChanges_Get(t *testing.T) {
	type fields struct {
		order   *ChangeOrder
		project *projectstack.Project
		stack   *projectstack.Stack
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ChangeStep
	}{
		{
			name: "t1",
			fields: fields{
				order: &ChangeOrder{
					ChangeSteps: map[string]*ChangeStep{
						"test-key": TestChangeStepOpCreate,
					},
				},
			},
			args: args{
				key: "test-key",
			},
			want: TestChangeStepOpCreate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Changes{
				ChangeOrder: tt.fields.order,
				project:     tt.fields.project,
				stack:       tt.fields.stack,
			}
			if got := p.Get(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Changes.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChanges_Values(t *testing.T) {
	type fields struct {
		order   *ChangeOrder
		project *projectstack.Project
		stack   *projectstack.Stack
	}
	type args struct {
		filters []ChangeStepFilterFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*ChangeStep
	}{
		{
			name: "filter-opcreate",
			fields: fields{
				order: &ChangeOrder{
					StepKeys:    TestStepKeys,
					ChangeSteps: TestChangeSteps,
				},
			},
			args: args{
				filters: []ChangeStepFilterFunc{CreateChangeStepFilter},
			},
			want: []*ChangeStep{TestChangeStepOpCreate},
		},
		{
			name: "filter-opdelete",
			fields: fields{
				order: &ChangeOrder{
					StepKeys:    TestStepKeys,
					ChangeSteps: TestChangeSteps,
				},
			},
			args: args{
				filters: []ChangeStepFilterFunc{DeleteChangeStepFilter},
			},
			want: []*ChangeStep{TestChangeStepOpDelete},
		},
		{
			name: "filter-opupdate",
			fields: fields{
				order: &ChangeOrder{
					StepKeys:    TestStepKeys,
					ChangeSteps: TestChangeSteps,
				},
			},
			args: args{
				filters: []ChangeStepFilterFunc{UpdateChangeStepFilter},
			},
			want: []*ChangeStep{TestChangeStepOpUpdate},
		},
		{
			name: "filter-opunchange",
			fields: fields{
				order: &ChangeOrder{
					StepKeys:    TestStepKeys,
					ChangeSteps: TestChangeSteps,
				},
			},
			args: args{
				filters: []ChangeStepFilterFunc{UnChangeChangeStepFilter},
			},
			want: []*ChangeStep{TestChangeStepOpUnChange},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Changes{
				ChangeOrder: tt.fields.order,
				project:     tt.fields.project,
				stack:       tt.fields.stack,
			}
			if got := p.Values(tt.args.filters...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Changes.Values() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChanges_Stack(t *testing.T) {
	type fields struct {
		order   *ChangeOrder
		project *projectstack.Project
		stack   *projectstack.Stack
	}
	tests := []struct {
		name   string
		fields fields
		want   *projectstack.Stack
	}{
		{
			name: "t1",
			fields: fields{
				order:   &ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*ChangeStep{}},
				project: &projectstack.Project{},
				stack: &projectstack.Stack{
					StackConfiguration: projectstack.StackConfiguration{
						Name: "test-name",
					},
					Path: "test-path",
				},
			},
			want: &projectstack.Stack{
				StackConfiguration: projectstack.StackConfiguration{
					Name: "test-name",
				},
				Path: "test-path",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Changes{
				ChangeOrder: tt.fields.order,
				project:     tt.fields.project,
				stack:       tt.fields.stack,
			}
			if got := p.Stack(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Changes.Stack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChanges_Project(t *testing.T) {
	type fields struct {
		order   *ChangeOrder
		project *projectstack.Project
		stack   *projectstack.Stack
	}
	tests := []struct {
		name   string
		fields fields
		want   *projectstack.Project
	}{
		{
			name: "t1",
			fields: fields{
				project: &projectstack.Project{
					ProjectConfiguration: projectstack.ProjectConfiguration{
						Name:   "test-name",
						Tenant: "test-tenant",
					},
					Path: "test-path",
				},
			},
			want: &projectstack.Project{
				ProjectConfiguration: projectstack.ProjectConfiguration{
					Name:   "test-name",
					Tenant: "test-tenant",
				},
				Path: "test-path",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewChanges(tt.fields.project, tt.fields.stack, tt.fields.order)
			if got := p.Project(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Changes.Project() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChanges_Diffs(t *testing.T) {
	type fields struct {
		order   *ChangeOrder
		project *projectstack.Project
		stack   *projectstack.Stack
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "t1",
			fields: fields{
				order: &ChangeOrder{
					StepKeys: []string{"test-key"},
					ChangeSteps: map[string]*ChangeStep{
						"test-key": TestChangeStepOpCreate,
					},
				},
			},
			want: `[32;1m[32;1mID: [0m[0m[32mid[0m
[32m[0m[32;1m[32;1mPlan: [0m[0m[32mCreating[0m
[32;1m[32;1mDiff: [0m[0m

`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Changes{
				ChangeOrder: tt.fields.order,
				project:     tt.fields.project,
				stack:       tt.fields.stack,
			}
			if got := p.Diffs(); got != tt.want {
				t.Errorf("Changes.Diffs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChanges_Preview(t *testing.T) {
	type fields struct {
		order   *ChangeOrder
		project *projectstack.Project
		stack   *projectstack.Stack
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "t1",
			fields: fields{
				order: &ChangeOrder{
					StepKeys: []string{"test-key"},
					ChangeSteps: map[string]*ChangeStep{
						"test-key": TestChangeStepOpCreate,
					},
				},
				stack: &projectstack.Stack{
					StackConfiguration: projectstack.StackConfiguration{
						Name: "test-name",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Changes{
				ChangeOrder: tt.fields.order,
				project:     tt.fields.project,
				stack:       tt.fields.stack,
			}
			p.Summary()
		})
	}
}

func Test_buildResourceStateMap(t *testing.T) {
	type args struct {
		rs []*models.Resource
	}
	tests := []struct {
		name string
		args args
		want map[string]*models.Resource
	}{
		{
			name: "t1",
			args: args{
				rs: []*models.Resource{},
			},
			want: map[string]*models.Resource{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildResourceStateMap(tt.args.rs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildResourceStateMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
