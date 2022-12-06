package pretty

import "github.com/pterm/pterm"

// Pretty style, contains the color and the style.
//
// Usage:
//
//	var s1 string = pretty.GreenBold("Hello")
//	var s2 string = pretty.NormalBold("Hello %s", "World")
//	fmt.Println(s1, s2)
var (
	// Preset Color
	// Red is an alias for FgRed.Sprintf.
	Red = pterm.FgRed.Sprintf
	// Cyan is an alias for FgCyan.Sprintf.
	Cyan = pterm.FgCyan.Sprintf
	// Gray is an alias for FgGray.Sprintf.
	Gray = pterm.FgGray.Sprintf
	// Blue is an alias for FgBlue.Sprintf.
	Blue = pterm.FgBlue.Sprintf
	// Black is an alias for FgBlack.Sprintf.
	Black = pterm.FgBlack.Sprintf
	// Green is an alias for FgGreen.Sprintf.
	Green = pterm.FgGreen.Sprintf
	// White is an alias for FgWhite.Sprintf.
	White = pterm.FgWhite.Sprintf
	// Yellow is an alias for FgYellow.Sprintf.
	Yellow = pterm.FgYellow.Sprintf
	// Magenta is an alias for FgMagenta.Sprintf.
	Magenta = pterm.FgMagenta.Sprintf

	// Normal is an alias for FgDefault.Sprintf.
	Normal = pterm.FgDefault.Sprintf

	// LightRed is a shortcut for FgLightRed.Sprintf.
	LightRed = pterm.FgLightRed.Sprintf
	// LightCyan is a shortcut for FgLightCyan.Sprintf.
	LightCyan = pterm.FgLightCyan.Sprintf
	// LightBlue is a shortcut for FgLightBlue.Sprintf.
	LightBlue = pterm.FgLightBlue.Sprintf
	// LightGreen is a shortcut for FgLightGreen.Sprintf.
	LightGreen = pterm.FgLightGreen.Sprintf
	// LightWhite is a shortcut for FgLightWhite.Sprintf.
	LightWhite = pterm.FgLightWhite.Sprintf
	// LightYellow is a shortcut for FgLightYellow.Sprintf.
	LightYellow = pterm.FgLightYellow.Sprintf
	// LightMagenta is a shortcut for FgLightMagenta.Sprintf.
	LightMagenta = pterm.FgLightMagenta.Sprintf

	// Preset Style for Color and Bold
	// RedBold is an shortcut for Sprintf of Style with Red and Bold.
	RedBold = pterm.NewStyle(pterm.FgRed, pterm.Bold).Sprintf
	// CyanBold is an shortcut for Sprintf of Style with FgCyan and Bold.
	CyanBold = pterm.NewStyle(pterm.FgCyan, pterm.Bold).Sprintf
	// GrayBold is an shortcut for Sprintf of Style with FgGray and Bold.
	GrayBold = pterm.NewStyle(pterm.FgGray, pterm.Bold).Sprintf
	// BlueBold is an shortcut for Sprintf of Style with FgBlue and Bold.
	BlueBold = pterm.NewStyle(pterm.FgBlue, pterm.Bold).Sprintf
	// BlackBold is an shortcut for Sprintf of Style with FgBlack and Bold.
	BlackBold = pterm.NewStyle(pterm.FgBlack, pterm.Bold).Sprintf
	// GreenBold is an shortcut for Sprintf of Style with FgGreen and Bold.
	GreenBold = pterm.NewStyle(pterm.FgGreen, pterm.Bold).Sprintf
	// WhiteBold is an shortcut for Sprintf of Style with FgWhite and Bold.
	WhiteBold = pterm.NewStyle(pterm.FgWhite, pterm.Bold).Sprintf
	// YellowBold is an shortcut for Sprintf of Style with FgYellow and Bold.
	YellowBold = pterm.NewStyle(pterm.FgYellow, pterm.Bold).Sprintf
	// MagentaBold is an shortcut for Sprintf of Style with FgMagenta and Bold.
	MagentaBold = pterm.NewStyle(pterm.FgMagenta, pterm.Bold).Sprintf

	// NormalBold is an shortcut for Sprintf of Style with FgDefault and Bold.
	NormalBold = pterm.NewStyle(pterm.FgDefault, pterm.Bold).Sprintf

	// LightRedBold is an shortcut for Sprintf of Style with FgLightRed and Bold.
	LightRedBold = pterm.NewStyle(pterm.FgLightRed, pterm.Bold).Sprintf
	// LightCyanBold is an shortcut for Sprintf of Style with FgLightCyan and Bold.
	LightCyanBold = pterm.NewStyle(pterm.FgLightCyan, pterm.Bold).Sprintf
	// LightBlueBold is an shortcut for Sprintf of Style with FgLightBlue and Bold.
	LightBlueBold = pterm.NewStyle(pterm.FgLightBlue, pterm.Bold).Sprintf
	// LightGreenBold is an shortcut for Sprintf of Style with FgLightGreen and Bold.
	LightGreenBold = pterm.NewStyle(pterm.FgLightGreen, pterm.Bold).Sprintf
	// LightWhiteBold is an shortcut for Sprintf of Style with FgLightWhite and Bold.
	LightWhiteBold = pterm.NewStyle(pterm.FgLightWhite, pterm.Bold).Sprintf
	// LightYellowBold is an shortcut for Sprintf of Style with FgLightYellow and Bold.
	LightYellowBold = pterm.NewStyle(pterm.FgLightYellow, pterm.Bold).Sprintf
	// LightMagentaBold is an shortcut for Sprintf of Style with FgLightMagenta and Bold.
	LightMagentaBold = pterm.NewStyle(pterm.FgLightMagenta, pterm.Bold).Sprintf
)
