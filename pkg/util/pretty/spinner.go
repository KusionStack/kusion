package pretty

import (
	"time"

	"github.com/pterm/pterm"
)

// Spinner style.
//
// Usage:
//
//	sp, _ := pretty.Spinner.Start("Starting ...")
//	time.Sleep(time.Second * 3)
//	sp.Success("Done")
var Spinner = pterm.SpinnerPrinter{
	Sequence:            []string{" ⣾ ", " ⣽ ", " ⣻ ", " ⢿ ", " ⡿ ", " ⣟ ", " ⣯ ", " ⣷ "},
	Style:               &pterm.ThemeDefault.SpinnerStyle,
	Delay:               time.Millisecond * 100,
	ShowTimer:           true,
	TimerRoundingFactor: time.Second,
	TimerStyle:          &pterm.ThemeDefault.TimerStyle,
	MessageStyle:        &pterm.ThemeDefault.InfoMessageStyle,
	SuccessPrinter:      &Success,
	FailPrinter:         &Error,
	WarningPrinter:      &Warning,
}
