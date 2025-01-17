package state

import "kusionstack.io/kusion/pkg/engine/operation/models"

type LineSummary struct {
	created, updated, deleted int
}

func (ls *LineSummary) Count(op models.ActionType) {
	switch op {
	case models.Create:
		ls.created++
	case models.Update:
		ls.updated++
	case models.Delete:
		ls.deleted++
	}
}

func (ls *LineSummary) GetCreated() int {
	return ls.created
}

func (ls *LineSummary) GetUpdated() int {
	return ls.updated
}

func (ls *LineSummary) GetDeleted() int {
	return ls.deleted
}
