package main

import (
	"fmt"
	"os"

	"github.com/Leechael/things3--cli/internal/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	executedCmd, err := rootCmd.ExecuteC()

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())

		switch cmd.HelpModeForError(err) {
		case cmd.HelpModeRoot:
			_ = rootCmd.Help()
		case cmd.HelpModeCommand:
			if executedCmd != nil {
				_ = executedCmd.Help()
			} else {
				_ = rootCmd.Help()
			}
		}
	}

	os.Exit(cmd.ExitCode(err))
}
