package internal

import (
	"fmt"
	"os"
)

var colorsEnabled bool

func init() {
	_, noColor := os.LookupEnv("NO_COLOR")
	forceColor := os.Getenv("FORCE_COLOR")
	colorsEnabled = !noColor && forceColor != "0"
}

func wrap(code, resetCode int) func(string) string {
	return func(text string) string {
		if colorsEnabled {
			return fmt.Sprintf("\x1b[%dm%s\x1b[%dm", code, text, resetCode)
		}
		return text
	}
}

// Color functions
var (
	Green   = wrap(32, 39)
	Red     = wrap(31, 39)
	Yellow  = wrap(33, 39)
	Blue    = wrap(34, 39)
	Cyan    = wrap(36, 39)
	Magenta = wrap(35, 39)
	Gray    = wrap(90, 39)
)

// Style functions
var (
	Bold      = wrap(1, 22)
	Dim       = wrap(2, 22)
	Italic    = wrap(3, 23)
	Underline = wrap(4, 24)
)
