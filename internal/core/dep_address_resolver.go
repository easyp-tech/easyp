package core

import (
	"strings"
)

type MirrorConfig struct {
	Origin string
	Use    string
}

type DepAddressResolver struct {
	mirrors []MirrorConfig
}

func NewDepAddressResolver(mirrors []MirrorConfig) DepAddressResolver {
	return DepAddressResolver{
		mirrors: mirrors,
	}
}

func (r DepAddressResolver) Resolve(requestedModule string) string {
	requestedModule = r.useMirrors(requestedModule)
	return requestedModule
}

func (r DepAddressResolver) useMirrors(requestedModule string) string {
	for _, mirror := range r.mirrors {
		requestedModule = strings.Replace(requestedModule, mirror.Origin, mirror.Use, 1)
	}

	return requestedModule
}
