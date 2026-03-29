package cmd

import "github.com/spf13/pflag"

func stringPointerIfChanged(flags *pflag.FlagSet, name string, value string) *string {
	if !flags.Changed(name) {
		return nil
	}
	result := value
	return &result
}

func boolPointerIfChanged(flags *pflag.FlagSet, name string, value bool) *bool {
	if !flags.Changed(name) {
		return nil
	}
	result := value
	return &result
}
