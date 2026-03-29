package main

import (
	"os"

	"github.com/Leechael/things3--cli/internal/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	err := rootCmd.Execute()
	os.Exit(cmd.ExitCode(err))
}
