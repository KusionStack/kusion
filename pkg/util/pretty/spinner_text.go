package pretty

import (
	"time"

	"github.com/pterm/pterm"
)

// SpinnerT text style.
//
// Usage:
//
//	sp, _ := pretty.SpinnerT.Start("Starting ...")
//	time.Sleep(time.Second * 3)
//	sp.Success("Done")
var SpinnerT = pterm.SpinnerPrinter{
	Sequence:            []string{" ⣾ ", " ⣽ ", " ⣻ ", " ⢿ ", " ⡿ ", " ⣟ ", " ⣯ ", " ⣷ "},
	Style:               &pterm.ThemeDefault.SpinnerStyle,
	Delay:               time.Millisecond * 100,
	ShowTimer:           true,
	TimerRoundingFactor: time.Second,
	TimerStyle:          &pterm.ThemeDefault.TimerStyle,
	MessageStyle:        &pterm.ThemeDefault.InfoMessageStyle,
	SuccessPrinter:      &SuccessT,
	FailPrinter:         &ErrorT,
	WarningPrinter:      &WarningT,
}
