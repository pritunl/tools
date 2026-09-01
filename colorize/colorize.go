// Package colorize provides ANSI escape sequences for coloring terminal
// output and a helper for wrapping a string in foreground and background
// color codes.
//
// The color constants are untyped string constants and may be passed
// anywhere a Color is expected:
//
//	fmt.Println(colorize.String("error", colorize.RedBold, colorize.None))
package colorize

// Color is an ANSI escape sequence that sets a terminal text attribute such
// as a foreground color, background color or bold weight. The empty string
// (None) applies no attribute.
type Color string

// ANSI escape sequences for terminal text attributes. Constants without a
// suffix set the foreground color, the Bold variants set the foreground color
// with bold weight, and the Bg variants set the background color.
const (
	// None applies no color or attribute.
	None = ""
	// Bold enables bold weight without changing the color.
	Bold       = "\033[1m"
	Black      = "\033[0;30m"
	BlackBold  = "\033[1;30m"
	Gray       = "\033[0;90m"
	GrayBold   = "\033[1;90m"
	Red        = "\033[0;31m"
	RedBold    = "\033[1;31m"
	Green      = "\033[0;32m"
	GreenBold  = "\033[1;32m"
	Yellow     = "\033[0;33m"
	YellowBold = "\033[1;33m"
	Blue       = "\033[0;34m"
	BlueBold   = "\033[1;34m"
	Purple     = "\033[0;35m"
	PurpleBold = "\033[1;35m"
	Cyan       = "\033[0;36m"
	CyanBold   = "\033[1;36m"
	White      = "\033[0;37m"
	WhiteBold  = "\033[1;37m"
	BlackBg    = "\033[40m"
	GrayBg     = "\033[100m"
	RedBg      = "\033[41m"
	GreenBg    = "\033[42m"
	YellowBg   = "\033[43m"
	BlueBg     = "\033[44m"
	PurpleBg   = "\033[45m"
	CyanBg     = "\033[46m"
	WhiteBg    = "\033[47m"
)

// String returns input wrapped in the fg and bg escape sequences followed by
// an ANSI reset so that attributes do not leak into subsequent output. Pass
// None for fg or bg to leave that attribute unset.
func String(input string, fg Color, bg Color) (str string) {
	str = string(fg) + string(bg) + input + "\033[0m"
	return
}
