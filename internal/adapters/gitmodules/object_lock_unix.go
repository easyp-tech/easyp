//go:build !windows

package gitmodules

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func tryObjectLock(file *os.File) (bool, error) {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
		return false, nil
	}
	return err == nil, err
}
func releaseObjectLock(file *os.File) error { return unix.Flock(int(file.Fd()), unix.LOCK_UN) }
