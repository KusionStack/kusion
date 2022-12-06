package pretty

import "github.com/pterm/pterm"

// Pretty prefix printer style.
//
// Usage:
//
//	pretty.Info.Println("Hello")
//	pretty.Warning.Println("Hello")
var (
	// Info returns a PrefixPrinter, which can be used to print text with an "info" Prefix.
	Info = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.InfoMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.InfoMessageStyle,
			Text:  "ℹ️",
		},
	}

	// Warning returns a PrefixPrinter, which can be used to print text with a "warning" Prefix.
	Warning = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.WarningMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.WarningMessageStyle,
			Text:  "❗",
		},
	}

	// Success returns a PrefixPrinter, which can be used to print text with a "success" Prefix.
	Success = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.SuccessMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessMessageStyle,
			Text:  "✅",
		},
	}

	// Error returns a PrefixPrinter, which can be used to print text with an "error" Prefix.
	Error = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.ErrorMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorMessageStyle,
			Text:  "❌",
		},
	}

	// Fatal returns a PrefixPrinter, which can be used to print text with an "fatal" Prefix.
	// NOTICE: Fatal terminates the application immediately!
	Fatal = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.FatalMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.FatalMessageStyle,
			Text:  "💣",
		},
		Fatal: true,
	}

	// FatalN returns a PrefixPrinter, which can be used to print text with an "fatal" Prefix.
	FatalN = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.FatalMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.FatalMessageStyle,
			Text:  "💣",
		},
	}

	// Debug Prints debug messages. By default, it will only print if PrintDebugMessages is true.
	// You can change PrintDebugMessages with EnableDebugMessages and DisableDebugMessages, or by setting the variable itself.
	Debug = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.DebugMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.DebugMessageStyle,
			Text:  "⭕",
		},
		Debugger: true,
	}

	// Check returns a PrefixPrinter, which can be used to print text with a "mark check" Prefix.
	Check = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.SuccessMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessMessageStyle,
			Text:  "✅",
		},
	}
)
