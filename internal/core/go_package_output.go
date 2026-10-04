package core

import (
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// goPackageOutputDirs uses the final descriptor options, including managed
// disables and overrides. Packages outside the prefix keep their plugin paths.
func goPackageOutputDirs(prefix string, files []string, descriptors []*descriptorpb.FileDescriptorProto) map[string]string {
	if template, _, marker := strings.Cut(prefix, "{{"); marker {
		separator := strings.LastIndex(template, "/")
		if separator < 0 {
			return nil
		}
		prefix = template[:separator]
	}
	prefix = strings.TrimRight(prefix, "/")
	if prefix == "" {
		return nil
	}
	targets := make(map[string]bool, len(files))
	for _, file := range files {
		targets[file] = true
	}
	directories := make(map[string]string)
	for _, descriptor := range descriptors {
		if !targets[descriptor.GetName()] {
			continue
		}
		importPath, _, _ := strings.Cut(descriptor.GetOptions().GetGoPackage(), ";")
		if importPath == prefix {
			directories[descriptor.GetName()] = "."
			continue
		}
		if directory, ok := strings.CutPrefix(importPath, prefix+"/"); ok {
			directories[descriptor.GetName()] = directory
		}
	}
	return directories
}

func goPackageOutputPath(file *pluginpb.CodeGeneratorResponse_File, directories map[string]string) string {
	name := file.GetName()
	if len(directories) == 0 || !strings.HasSuffix(name, ".go") {
		return name
	}
	// Go and Go gRPC headers disambiguate user_grpc.proto from the gRPC
	// companion of user.proto. Plugins without that header use the longest basename.
	var declaredSource string
	for line := range strings.Lines(file.GetContent()) {
		if strings.HasPrefix(line, "package ") {
			break
		}
		if source, ok := strings.CutPrefix(strings.TrimSpace(line), "// source: "); ok {
			declaredSource = strings.TrimSpace(source)
			break
		}
	}
	outputDir := filepath.ToSlash(filepath.Dir(name))
	outputBase := filepath.Base(name)
	longestBase := ""
	output := name
	for source, directory := range directories {
		if declaredSource != "" && source != declaredSource {
			continue
		}
		if outputDir != filepath.ToSlash(filepath.Dir(source)) {
			continue
		}
		base := strings.TrimSuffix(filepath.Base(source), ".proto")
		if !strings.HasPrefix(outputBase, base+".") && !strings.HasPrefix(outputBase, base+"_") {
			continue
		}
		if len(base) > len(longestBase) {
			longestBase = base
			output = filepath.ToSlash(filepath.Join(directory, outputBase))
		}
	}
	return output
}
