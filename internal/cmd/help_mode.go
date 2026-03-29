package cmd

import "strings"

// HelpMode controls whether help text should be shown for a command error.
type HelpMode int

const (
	HelpModeNone HelpMode = iota
	HelpModeRoot
	HelpModeCommand
)

// HelpModeForError classifies Cobra/runtime errors into help rendering behavior.
func HelpModeForError(err error) HelpMode {
	if err == nil {
		return HelpModeNone
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))

	if strings.Contains(message, "unknown command") {
		return HelpModeRoot
	}

	if strings.Contains(message, "unknown flag") ||
		strings.Contains(message, "unknown shorthand flag") ||
		strings.Contains(message, "flag needs an argument") ||
		strings.Contains(message, "required flag(s)") ||
		strings.Contains(message, "requires at least") ||
		strings.Contains(message, "requires at most") ||
		strings.Contains(message, "accepts ") ||
		strings.Contains(message, "--") && strings.Contains(message, "is required") ||
		strings.Contains(message, "at least one of --") ||
		strings.Contains(message, "either --") ||
		strings.Contains(message, "mutually exclusive") {
		return HelpModeCommand
	}

	return HelpModeNone
}
