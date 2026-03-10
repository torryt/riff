package internal

import "os"

// IsInteractive returns true when stdin is connected to a terminal,
// indicating that the user can answer interactive prompts.
func IsInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// If stdin is a character device (not a pipe/file), it's a terminal.
	return fi.Mode()&os.ModeCharDevice != 0
}
