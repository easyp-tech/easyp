//go:build !unix

package gitcommand

import "os/exec"

func configureCancellation(_ *exec.Cmd) {}
