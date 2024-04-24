package terminal

import (
	"github.com/pterm/pterm"
	"kusionstack.io/kusion/pkg/util/pretty"
)

type UI struct {
	SpinnerPrinter           *pterm.SpinnerPrinter
	ProgressbarPrinter       *pterm.ProgressbarPrinter
	InteractiveSelectPrinter *pterm.InteractiveSelectPrinter
}

// DefaultUI returns a UI for Kusion CLI display with default
// SpinnerPrinter, ProgressbarPrinter and InteractiveSelectPrinter.
func DefaultUI() *UI {
	return &UI{
		SpinnerPrinter:           &pretty.SpinnerT,
		ProgressbarPrinter:       &pterm.DefaultProgressbar,
		InteractiveSelectPrinter: &pterm.DefaultInteractiveSelect,
	}
}
