package gitmodules

import (
	"context"
	"fmt"
	"slices"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func (c *Cache) resolveBSRDependencies(ctx context.Context, module v1.Module) (v1.Module, []v1.BSRResolution, error) {
	if len(module.BSRDependencies) == 0 {
		return module, nil, nil
	}
	if c.bsrResolver == nil {
		return v1.Module{}, nil, fmt.Errorf("BSR resolver is not configured for %s", module.Name)
	}
	module.Requires = slices.Clone(module.Requires)
	var bindings []v1.BSRResolution
	resolved := make(map[v1.BSRDependency]v1.BSRResolution)
	for _, dependency := range module.BSRDependencies {
		if err := ctx.Err(); err != nil {
			return v1.Module{}, nil, fmt.Errorf("Err: %w", err)
		}
		// The same request can originate in several workspace configs. Resolve it
		// once, but retain each origin in the parent's lock entry.
		request := dependency
		request.Config = ""
		binding, ok := resolved[request]
		if ok {
			binding.Dependency = dependency
		} else {
			var err error
			binding, err = c.bsrResolver.Resolve(ctx, dependency)
			if err != nil {
				return v1.Module{}, nil, fmt.Errorf("Resolve: %s: %w", dependency.Config, err)
			}
			if binding.Dependency != dependency {
				return v1.Module{}, nil, fmt.Errorf("BSR resolver changed BSR dependency %s in %s", dependency.Module, dependency.Config)
			}
			if err := binding.Validate(); err != nil {
				return v1.Module{}, nil, fmt.Errorf("Validate: %w", err)
			}
			resolved[request] = binding
		}
		if !slices.Contains(bindings, binding) {
			bindings = append(bindings, binding)
		}
		if !slices.Contains(module.Requires, binding.Git) {
			module.Requires = append(module.Requires, binding.Git)
		}
	}
	return module, bindings, nil
}

func restoreBSRDependencies(module v1.Module, bindings []v1.BSRResolution) (v1.Module, error) {
	if len(module.BSRDependencies) > 0 && len(bindings) == 0 {
		return v1.Module{}, fmt.Errorf("%s: protobuf.lock has no BSR bindings; run easyp mod tidy", module.Name)
	}
	if len(bindings) != len(module.BSRDependencies) {
		return v1.Module{}, fmt.Errorf("%s: protobuf.lock BSR metadata does not match the dependency configs; run easyp mod tidy", module.Name)
	}
	byDependency := make(map[v1.BSRDependency]v1.Requirement, len(bindings))
	for _, binding := range bindings {
		if err := binding.Validate(); err != nil {
			return v1.Module{}, fmt.Errorf("Validate: %w", err)
		}
		if _, ok := byDependency[binding.Dependency]; ok {
			return v1.Module{}, fmt.Errorf("duplicate BSR binding for %s in %s", binding.Dependency.Module, binding.Dependency.Config)
		}
		byDependency[binding.Dependency] = binding.Git
	}
	module.Requires = slices.Clone(module.Requires)
	for _, dependency := range module.BSRDependencies {
		git, ok := byDependency[dependency]
		if !ok {
			return v1.Module{}, fmt.Errorf("%s: protobuf.lock BSR metadata for %s in %s does not match the dependency config; run easyp mod tidy", module.Name, dependency.Module, dependency.Config)
		}
		if !slices.Contains(module.Requires, git) {
			module.Requires = append(module.Requires, git)
		}
	}
	return module, nil
}
