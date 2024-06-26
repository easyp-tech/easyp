package models

import (
	"errors"
)

var (
	ErrVersionNotFound    = errors.New("version not found")
	ErrFileNotFound       = errors.New("file not found")
	ErrModuleNotInstalled = errors.New("module not installed")
)
