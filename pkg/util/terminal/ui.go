package terminal

import (
	"github.com/liu-hm19/pterm"
	"kusionstack.io/kusion/pkg/util/pretty"
)

type UI struct {
	SpinnerPrinter           *pterm.SpinnerPrinter
	ProgressbarPrinter       *pterm.ProgressbarPrinter
	InteractiveSelectPrinter *pterm.InteractiveSelectPrinter
	MultiPrinter             *pterm.MultiPrinter
}

// DefaultUI returns a UI for Kusion CLI display with default
// SpinnerPrinter, ProgressbarPrinter and InteractiveSelectPrinter.
func DefaultUI() *UI {
	return &UI{
		SpinnerPrinter:           &pretty.SpinnerT,
		ProgressbarPrinter:       &pterm.DefaultProgressbar,
		InteractiveSelectPrinter: &pterm.DefaultInteractiveSelect,
		MultiPrinter:             &pterm.DefaultMultiPrinter,
	}
}
