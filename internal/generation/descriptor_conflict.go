package generation

import (
	"bytes"
	"fmt"
	"path/filepath"
	"slices"

	"google.golang.org/protobuf/reflect/protoreflect"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func (target preparedDescriptorTarget) origin(file string) string {
	owner, ok := target.owners[file]
	if !ok {
		return target.label + ", compiler-provided import"
	}
	version := target.versions[owner]
	if version == "" {
		version = "version unavailable (local source or unpinned metadata)"
	}
	return fmt.Sprintf("%s, source module %q, %s", target.label, owner, version)
}

// Version metadata enriches diagnostics only. Unused or missing locks must not
// prevent export of a graph already resolved and checked by the module layer.
func descriptorVersions(request Request, selected v1GenerationModule) map[string]string {
	versions := make(map[string]string)
	scope := selected.resolutionDir
	if scope == "" {
		scope = selected.directory
	}
	if scope == selected.directory {
		versions[selected.module.Name] = "workspace " + relativeDescriptorPath(request.WorkDir, selected.directory)
	}
	scopes := []string{scope}
	for _, directory := range scopes {
		lock, err := modules.ReadLock(filepath.Join(directory, v1.LockFile))
		if err == nil {
			for _, entry := range lock.Modules {
				versions[entry.Source] = fmt.Sprintf("version %s, commit %s", entry.Version, entry.Commit)
			}
		}
	}
	for _, directory := range scopes {
		module, err := modules.ReadModuleOrDefault(directory)
		if err != nil {
			continue
		}
		for _, replacement := range module.Replaces {
			versions[replacement.Module] = fmt.Sprintf("local replacement %q from %q", replacement.Target, relativeDescriptorPath(request.WorkDir, directory))
		}
	}
	return versions
}

// descriptorDifference reports a differing field, not just the proto filename.
func descriptorDifference(a, b protoreflect.Message, prefix string) string {
	fields := make(map[protoreflect.FieldNumber]protoreflect.FieldDescriptor)
	collect := func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		fields[fd.Number()] = fd
		return true
	}
	a.Range(collect)
	b.Range(collect)
	numbers := make([]int, 0, len(fields))
	for number := range fields {
		numbers = append(numbers, int(number))
	}
	slices.Sort(numbers)
	for _, number := range numbers {
		fd := fields[protoreflect.FieldNumber(number)]
		field := string(fd.Name())
		if prefix != "" {
			field = prefix + "." + field
		}
		x, y := a.Get(fd), b.Get(fd)
		if a.Has(fd) != b.Has(fd) {
			return fmt.Sprintf("%s: presence differs (%t versus %t)", field, a.Has(fd), b.Has(fd))
		}
		switch {
		case fd.IsList():
			left, right := x.List(), y.List()
			if left.Len() != right.Len() {
				return fmt.Sprintf("%s: length %d versus %d", field, left.Len(), right.Len())
			}
			for i := range left.Len() {
				if diff := descriptorValueDifference(left.Get(i), right.Get(i), fd, fmt.Sprintf("%s[%d]", field, i)); diff != "" {
					return diff
				}
			}
		case fd.IsMap():
			left, right := x.Map(), y.Map()
			keys := make(map[string]protoreflect.MapKey)
			collectKey := func(key protoreflect.MapKey, _ protoreflect.Value) bool {
				keys[key.String()] = key
				return true
			}
			left.Range(collectKey)
			right.Range(collectKey)
			names := make([]string, 0, len(keys))
			for name := range keys {
				names = append(names, name)
			}
			slices.Sort(names)
			for _, name := range names {
				key := keys[name]
				entry := fmt.Sprintf("%s[%q]", field, name)
				if left.Has(key) != right.Has(key) {
					return entry + ": presence differs"
				}
				if diff := descriptorValueDifference(left.Get(key), right.Get(key), fd.MapValue(), entry); diff != "" {
					return diff
				}
			}
		default:
			if diff := descriptorValueDifference(x, y, fd, field); diff != "" {
				return diff
			}
		}
	}
	if !bytes.Equal(a.GetUnknown(), b.GetUnknown()) {
		return prefix + ": unknown or extension fields differ"
	}
	return ""
}

func descriptorValueDifference(a, b protoreflect.Value, fd protoreflect.FieldDescriptor, path string) string {
	switch fd.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return descriptorDifference(a.Message(), b.Message(), path)
	case protoreflect.BytesKind:
		if !bytes.Equal(a.Bytes(), b.Bytes()) {
			return path + ": bytes differ"
		}
	default:
		if a.Interface() != b.Interface() {
			left, right := a.String(), b.String()
			if len(left) > 120 {
				left = left[:120] + "..."
			}
			if len(right) > 120 {
				right = right[:120] + "..."
			}
			return fmt.Sprintf("%s: %q versus %q", path, left, right)
		}
	}
	return ""
}
