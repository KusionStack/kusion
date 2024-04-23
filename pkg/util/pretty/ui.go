package pretty

import "github.com/pterm/pterm"

type UI struct {
	SpinnerPrinter           *pterm.SpinnerPrinter
	ProgressbarPrinter       *pterm.ProgressbarPrinter
	InteractiveSelectPrinter *pterm.InteractiveSelectPrinter
}

// DefaultUI returns a UI for Kusion CLI display with default
// SpinnerPrinter, ProgressbarPrinter and InteractiveSelectPrinter.
func DefaultUI() *UI {
	return &UI{
		SpinnerPrinter:           &SpinnerT,
		ProgressbarPrinter:       &pterm.DefaultProgressbar,
		InteractiveSelectPrinter: &pterm.DefaultInteractiveSelect,
	}
}
