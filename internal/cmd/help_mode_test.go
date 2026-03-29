package cmd

import (
	"errors"
	"testing"
)

func TestHelpModeForError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want HelpMode
	}{
		{name: "nil", err: nil, want: HelpModeNone},
		{name: "unknown command", err: errors.New("unknown command \"x\" for \"things3-cli\""), want: HelpModeRoot},
		{name: "unknown flag", err: errors.New("unknown flag: --abc"), want: HelpModeCommand},
		{name: "required flag", err: errors.New("required flag(s) \"id\" not set"), want: HelpModeCommand},
		{name: "argument required", err: errors.New("--id is required"), want: HelpModeCommand},
		{name: "mutually exclusive", err: errors.New("--project and --area are mutually exclusive"), want: HelpModeCommand},
		{name: "internal error", err: errors.New("query todos: database is locked"), want: HelpModeNone},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := HelpModeForError(test.err)
			if got != test.want {
				t.Fatalf("HelpModeForError() = %v, want %v", got, test.want)
			}
		})
	}
}
