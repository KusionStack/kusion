package pretty

import (
	"fmt"
	"testing"

	"github.com/pterm/pterm"
)

func TestPrefixTextPrinter(t *testing.T) {
	fmt.Println("PrefixTextPrinter:")
	defer fmt.Println("")

	// Preset prefix printer style.
	InfoT.Println("This is a InfoT message")
	WarningT.Println("This is a WarningT message")
	SuccessT.Println("This is a SuccessT message")
	ErrorT.Println("This is a ErrorT message")
	FatalNT.Println("This is a FatalNT message")
	DebugT.Println("This is a DebugT message")

	// Custom prefix printer style.
	InfoT.WithMessageStyle(pterm.GrayBoxStyle).Println("This is a InfoT message with gray box style")
	InfoT.WithMessageStyle(pterm.NewStyle(pterm.BgLightWhite, pterm.FgBlue)).Println("This is a InfoT message with FgLightWhite and FgBlue")
	InfoT.WithShowLineNumber(true).Println("This is a InfoT message with line number")
}
