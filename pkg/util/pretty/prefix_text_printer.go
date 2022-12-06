package pretty

import "github.com/pterm/pterm"

// Pretty prefix text printer style.
//
// Usage:
//
//	pretty.InfoT.Println("Hello")
//	pretty.WarningT.Println("Hello")
var (
	// DebugT Prints debug messages. By default it will only print if PrintDebugMessages is true.
	// You can change PrintDebugMessages with EnableDebugMessages and DisableDebugMessages, or by setting the variable itself.
	DebugT = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.DebugMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.DebugMessageStyle,
			Text:  "#",
		},
		Debugger: true,
	}

	// InfoT returns a PrefixPrinter, which can be used to print text with an "info" Prefix.
	InfoT = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.SuccessMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessMessageStyle,
			Text:  "»",
		},
	}

	// WarningT returns a PrefixPrinter, which can be used to print text with a "warning" Prefix.
	WarningT = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.WarningMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.WarningMessageStyle,
			Text:  "!",
		},
	}

	// ErrorT returns a PrefixPrinter, which can be used to print text with an "error" Prefix.
	ErrorT = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.ErrorMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorMessageStyle,
			Text:  "✘",
		},
	}

	// FatalT returns a PrefixPrinter, which can be used to print text with an "fatal" Prefix.
	// NOTICE: Fatal terminates the application immediately!
	FatalT = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.FatalMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.FatalMessageStyle,
			Text:  "☒",
		},
		Fatal: true,
	}

	// FatalN returns a PrefixPrinter, which can be used to print text with an "fatal" Prefix.
	FatalNT = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.FatalMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.FatalMessageStyle,
			Text:  "☒",
		},
	}

	// SuccessT returns a PrefixPrinter, which can be used to print text with a "success" Prefix.
	SuccessT = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.SuccessMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessMessageStyle,
			Text:  "✔︎",
		},
	}

	// CheckT returns a PrefixPrinter, which can be used to print text with a "mark check" Prefix.
	CheckT = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.SuccessMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessMessageStyle,
			Text:  "☑",
		},
	}
)
