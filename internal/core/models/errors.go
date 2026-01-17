package models

import (
	"errors"
)

var (
	ErrVersionNotFound        = errors.New("version not found")
	ErrFileNotFound           = errors.New("file not found")
	ErrModuleNotInstalled     = errors.New("module not installed")
	ErrModuleInfoFileNotFound = errors.New("module info file not found")
	ErrHashDependencyMismatch = errors.New("hash dependency mismatch")
)
