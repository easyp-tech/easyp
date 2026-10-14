package main

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigFlagIsCommandLocal(t *testing.T) {
	app := newApp()

	for _, flag := range app.Flags {
		require.NotContains(t, flag.Names(), "config")
		require.NotContains(t, flag.Names(), "cfg")
	}
	withConfig := map[string]bool{"lint": true, "breaking": true, "generate": true, "validate-config": true}
	for _, command := range app.Commands {
		var names []string
		for _, flag := range command.Flags {
			names = append(names, flag.Names()...)
		}
		if withConfig[command.Name] {
			require.Contains(t, names, "config", command.Name)
		} else {
			require.NotContains(t, names, "config", command.Name)
		}
	}
}

func TestGlobalConfigFlagIsRejected(t *testing.T) {
	app := newApp()
	app.Writer, app.ErrWriter = io.Discard, io.Discard

	err := app.Run([]string{"easyp", "--cfg", "easyp.yaml", "lint"})

	require.ErrorContains(t, err, "flag provided but not defined: -cfg")
}
