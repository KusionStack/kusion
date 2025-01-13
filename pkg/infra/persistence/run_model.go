package persistence

import (
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"

	"gorm.io/gorm"
)

// RunModel is a DO used to map the entity to the database.
type RunModel struct {
	gorm.Model
	// RunType is the type of the run.
	Type string
	// StackID is the stack ID of the run.
	StackID uint
	// Stack is the stack of the run.
	Stack *StackModel
	// Workspace is the target workspace of the run.
	Workspace string
	// Status is the status of the run.
	Status string
	// Result is the result of the run.
	Result string
	// Logs is the logs of the run.
	Logs string
	// Trace is the trace of the run.
	Trace string
}

// The TableName method returns the name of the database table that the struct is mapped to.
func (m *RunModel) TableName() string {
	return "run"
}

// ToEntity converts the DO to an entity.
func (m *RunModel) ToEntity() (*entity.Run, error) {
	if m == nil {
		return nil, ErrRunModelNil
	}

	runType, err := constant.ParseRunType(m.Type)
	if err != nil {
		return nil, ErrFailedToGetRunType
	}

	runStatus, err := constant.ParseRunStatus(m.Status)
	if err != nil {
		return nil, ErrFailedToGetRunStatus
	}

	stackEntity, err := m.Stack.ToEntity()
	if err != nil {
		return nil, err
	}

	return &entity.Run{
		ID:                m.ID,
		Type:              runType,
		Stack:             stackEntity,
		Workspace:         m.Workspace,
		Status:            runStatus,
		Result:            m.Result,
		Trace:             m.Trace,
		Logs:              m.Logs,
		CreationTimestamp: m.CreatedAt,
		UpdateTimestamp:   m.UpdatedAt,
	}, nil
}

// FromEntity converts an entity to a DO.
func (m *RunModel) FromEntity(e *entity.Run) error {
	if m == nil {
		return ErrRunModelNil
	}

	if e.Stack != nil {
		m.StackID = e.Stack.ID
		m.Stack.FromEntity(e.Stack)
	}

	m.ID = e.ID
	m.Type = string(e.Type)
	m.Workspace = e.Workspace
	m.Status = string(e.Status)
	m.Result = e.Result
	m.Logs = e.Logs
	m.Trace = e.Trace
	m.CreatedAt = e.CreationTimestamp
	m.UpdatedAt = e.UpdateTimestamp

	return nil
}
