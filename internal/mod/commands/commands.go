package commands

import (
	"github.com/easyp-tech/easyp/internal/mod/dependency"
)

type (
	Dirs interface {
		CacheDir(dep dependency.Dependency) (string, error)
	}

	Commands struct {
		dirs Dirs
	}
)

func New(dirs Dirs) *Commands {
	return &Commands{
		dirs: dirs,
	}
}
